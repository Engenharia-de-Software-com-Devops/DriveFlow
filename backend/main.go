// Comando api: sobe o servidor HTTP da plataforma DriveFlow.
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

	"driveflow/backend/internal/api"
	"driveflow/backend/internal/armazenamento"
	"driveflow/backend/internal/locacao"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := executar(log); err != nil {
		log.Error("api encerrada com falha", "erro", err)
		os.Exit(1)
	}
}

func executar(log *slog.Logger) error {
	endereco := valorOuPadrao("API_PORT", "8080")

	repo, fechar, err := abrirArmazenamento(log)
	if err != nil {
		return err
	}
	defer fechar()

	servidor := &http.Server{
		Addr:              ":" + endereco,
		Handler:           api.NovoServidor(locacao.NovoServico(repo, nil, nil), log),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	// Encerramento gracioso: para de aceitar conexoes e aguarda as em andamento.
	encerrar := make(chan os.Signal, 1)
	signal.Notify(encerrar, os.Interrupt, syscall.SIGTERM)

	falhas := make(chan error, 1)
	go func() {
		log.Info("api ouvindo", "porta", endereco, "versao", api.Versao)
		if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			falhas <- err
		}
	}()

	select {
	case err := <-falhas:
		return err
	case <-encerrar:
		log.Info("sinal recebido, encerrando a api")
		ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelar()
		return servidor.Shutdown(ctx)
	}
}

// abrirArmazenamento escolhe a persistencia pela variavel DATABASE_URL.
//
// Com DATABASE_URL definida (o caso do docker compose, onde existe o container
// db), a api conecta no PostgreSQL e aplica as migrations pendentes antes de
// aceitar requisicoes. Sem ela, sobe com armazenamento em memoria, o que
// mantem `go run .` e `go test ./...` executaveis sem depender de container.
func abrirArmazenamento(log *slog.Logger) (locacao.Repositorio, func(), error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Warn("DATABASE_URL nao definida: usando armazenamento em memoria (os dados nao persistem)")
		return armazenamento.NovaMemoria(), func() {}, nil
	}

	ctx, cancelar := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelar()

	db, err := armazenamento.Conectar(ctx, url, 45*time.Second)
	if err != nil {
		return nil, nil, err
	}
	log.Info("conectado ao postgres")

	if err := armazenamento.AplicarMigracoes(db, log); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	return armazenamento.NovoPostgres(db), func() { fecharBanco(db, log) }, nil
}

func fecharBanco(db *sql.DB, log *slog.Logger) {
	if err := db.Close(); err != nil {
		log.Error("falha ao fechar conexao com o banco", "erro", err)
	}
}

func valorOuPadrao(chave, padrao string) string {
	if v := os.Getenv(chave); v != "" {
		return v
	}
	return padrao
}
