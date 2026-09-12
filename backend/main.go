// Comando api: sobe o servidor HTTP da plataforma DriveFlow.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	repo := armazenamento.NovaMemoria()
	log.Info("armazenamento em memoria ativo (dados nao persistem entre reinicios)")

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

func valorOuPadrao(chave, padrao string) string {
	if v := os.Getenv(chave); v != "" {
		return v
	}
	return padrao
}
