package locacao

import (
	"context"
	"time"
)

// Repositorio abstrai a persistencia do dominio. O servico depende apenas
// desta interface, o que permite rodar a API com armazenamento em memoria
// (desenvolvimento e testes) ou com PostgreSQL (containers e producao).
type Repositorio interface {
	CriarEmpresa(ctx context.Context, e Empresa) (Empresa, error)
	BuscarEmpresa(ctx context.Context, id string) (Empresa, error)

	CriarVeiculo(ctx context.Context, v Veiculo) (Veiculo, error)
	BuscarVeiculo(ctx context.Context, empresaID, veiculoID string) (Veiculo, error)
	ListarVeiculos(ctx context.Context, empresaID string) ([]Veiculo, error)

	CriarLocacao(ctx context.Context, l Locacao) (Locacao, error)
	AtualizarLocacao(ctx context.Context, l Locacao) (Locacao, error)
	BuscarLocacao(ctx context.Context, empresaID, locacaoID string) (Locacao, error)
	ListarLocacoes(ctx context.Context, empresaID string) ([]Locacao, error)
	// LocacoesAtivasDoVeiculo devolve os contratos ainda nao encerrados do veiculo,
	// usados para detectar conflito de reserva.
	LocacoesAtivasDoVeiculo(ctx context.Context, empresaID, veiculoID string) ([]Locacao, error)
}

// GeradorID produz identificadores para os registros criados.
type GeradorID func() string

// Relogio fornece o instante atual. Injetado para tornar os testes deterministicos.
type Relogio func() time.Time
