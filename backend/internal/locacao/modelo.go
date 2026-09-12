// Package locacao concentra o dominio da plataforma multi-tenant de locacao
// de veiculos: empresas, frota e contratos de locacao.
package locacao

import (
	"errors"
	"time"
)

// StatusVeiculo representa a situacao operacional de um veiculo da frota.
type StatusVeiculo string

const (
	VeiculoDisponivel StatusVeiculo = "disponivel"
	VeiculoManutencao StatusVeiculo = "manutencao"
)

// StatusLocacao representa o ciclo de vida de um contrato de locacao.
type StatusLocacao string

const (
	LocacaoAberta    StatusLocacao = "aberta"
	LocacaoEncerrada StatusLocacao = "encerrada"
)

// Empresa e o tenant da plataforma. Cada empresa enxerga apenas a propria frota.
type Empresa struct {
	ID     string    `json:"id"`
	Nome   string    `json:"nome"`
	CNPJ   string    `json:"cnpj"`
	Criada time.Time `json:"criada_em"`
}

// Veiculo e um item da frota e pertence sempre a uma unica empresa.
type Veiculo struct {
	ID           string        `json:"id"`
	EmpresaID    string        `json:"empresa_id"`
	Placa        string        `json:"placa"`
	Modelo       string        `json:"modelo"`
	Categoria    string        `json:"categoria"`
	TarifaDiaria int64         `json:"tarifa_diaria"` // em centavos
	Status       StatusVeiculo `json:"status"`
}

// Locacao e o contrato que reserva um veiculo por um periodo.
type Locacao struct {
	ID            string        `json:"id"`
	EmpresaID     string        `json:"empresa_id"`
	VeiculoID     string        `json:"veiculo_id"`
	Cliente       string        `json:"cliente"`
	Inicio        time.Time     `json:"inicio"`
	FimPrevisto   time.Time     `json:"fim_previsto"`
	DevolvidoEm   *time.Time    `json:"devolvido_em,omitempty"`
	ValorPrevisto int64         `json:"valor_previsto"` // em centavos
	ValorFinal    *int64        `json:"valor_final,omitempty"`
	Status        StatusLocacao `json:"status"`
}

// Erros de dominio. A camada HTTP os traduz em codigos de status.
var (
	ErrNaoEncontrado       = errors.New("recurso nao encontrado")
	ErrDadosInvalidos      = errors.New("dados invalidos")
	ErrConflitoReserva     = errors.New("veiculo ja reservado no periodo informado")
	ErrVeiculoIndisponivel = errors.New("veiculo indisponivel para locacao")
	ErrPlacaDuplicada      = errors.New("placa ja cadastrada para esta empresa")
	ErrLocacaoEncerrada    = errors.New("locacao ja encerrada")
)
