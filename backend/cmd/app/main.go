// Comando app: sobe o servidor HTTP da plataforma DriveFlow.
//
// E o unico lugar que conhece todas as camadas ao mesmo tempo: le a
// configuracao, escolhe a implementacao do repositorio, monta os casos de uso
// e entrega tudo pronto para a camada de delivery.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"driveflow/backend/configs"
	httpdelivery "driveflow/backend/internal/delivery/http"
	"driveflow/backend/internal/repository"
	"driveflow/backend/internal/usecases"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(configs.Load(), log); err != nil {
		log.Error("api encerrada com falha", "erro", err)
		os.Exit(1)
	}
}

func run(cfg configs.Config, log *slog.Logger) error {
	repo, closeRepo, err := openRepository(cfg, log)
	if err != nil {
		return err
	}
	defer closeRepo()

	server := &http.Server{
		Addr:              cfg.Address(),
		Handler:           newHandler(repo, log),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	// Encerramento gracioso: para de aceitar conexoes e aguarda as em andamento.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	failures := make(chan error, 1)
	go func() {
		log.Info("api ouvindo", "porta", cfg.Port, "versao", httpdelivery.Version)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failures <- err
		}
	}()

	select {
	case err := <-failures:
		return err
	case <-shutdown:
		log.Info("sinal recebido, encerrando a api")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	}
}

// newHandler injeta o repositorio nos casos de uso e os casos de uso na
// camada HTTP. E aqui que as dependencias apontam para dentro.
func newHandler(repo usecases.Repository, log *slog.Logger) http.Handler {
	return httpdelivery.NewServer(
		usecases.NewCompanyService(repo, nil, nil),
		usecases.NewVehicleService(repo, nil),
		usecases.NewRentalService(repo, nil, nil),
		log,
	)
}

// openRepository escolhe a persistencia pela configuracao.
//
// Com DATABASE_URL definida (o caso do docker compose, onde existe o container
// db), a api conecta no PostgreSQL e aplica as migrations pendentes antes de
// aceitar requisicoes. Sem ela, sobe com armazenamento em memoria, o que
// mantem `go run ./cmd/app` e `go test ./...` executaveis sem depender de
// container.
func openRepository(cfg configs.Config, log *slog.Logger) (usecases.Repository, func(), error) {
	if !cfg.UsesDatabase() {
		log.Warn("DATABASE_URL nao definida: usando armazenamento em memoria (os dados nao persistem)")
		return repository.NewMemoryRepository(), func() {}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	db, err := repository.Connect(ctx, cfg.DatabaseURL, 45*time.Second)
	if err != nil {
		return nil, nil, err
	}
	log.Info("conectado ao postgres")

	if err := repository.ApplyMigrations(db, log); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	return repository.NewPostgresRepository(db), func() { closeDatabase(db, log) }, nil
}

func closeDatabase(db *sql.DB, log *slog.Logger) {
	if err := db.Close(); err != nil {
		log.Error("falha ao fechar conexao com o banco", "erro", err)
	}
}
