// Package configs le a configuracao da aplicacao a partir do ambiente.
//
// Concentrar a leitura aqui mantem os os.Getenv fora das camadas de dominio e
// de persistencia: elas recebem os valores ja resolvidos.
package configs

import "os"

// Padroes usados quando a variavel de ambiente nao esta definida.
const (
	DefaultPort = "8080"
)

// Config reune tudo o que a api precisa saber do ambiente.
type Config struct {
	// Port e a porta HTTP em que a api escuta.
	Port string
	// DatabaseURL e a string de conexao do PostgreSQL. Vazia significa rodar
	// com o repositorio em memoria.
	DatabaseURL string
}

// Load monta a configuracao a partir das variaveis de ambiente.
func Load() Config {
	return Config{
		Port:        valueOrDefault("API_PORT", DefaultPort),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

// UsesDatabase indica se a api deve conectar no PostgreSQL. Sem DATABASE_URL a
// aplicacao sobe em memoria, o que mantem `go run ./cmd/app` e `go test ./...`
// executaveis sem depender de container.
func (c Config) UsesDatabase() bool {
	return c.DatabaseURL != ""
}

// Address devolve o endereco de escuta no formato aceito por net/http.
func (c Config) Address() string {
	return ":" + c.Port
}

func valueOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
