package entities

import "time"

// RentalStatus representa o ciclo de vida de um contrato de locacao.
//
// Assim como VehicleStatus, os valores sao persistidos e validados por
// constraint CHECK no banco, entao permanecem em portugues.
type RentalStatus string

const (
	RentalOpen   RentalStatus = "aberta"
	RentalClosed RentalStatus = "encerrada"
)

// Rental e o contrato que reserva um veiculo por um periodo.
type Rental struct {
	ID             string       `json:"id"`
	CompanyID      string       `json:"empresa_id"`
	VehicleID      string       `json:"veiculo_id"`
	Customer       string       `json:"cliente"`
	Start          time.Time    `json:"inicio"`
	ExpectedEnd    time.Time    `json:"fim_previsto"`
	ReturnedAt     *time.Time   `json:"devolvido_em,omitempty"`
	EstimatedTotal int64        `json:"valor_previsto"` // em centavos
	FinalTotal     *int64       `json:"valor_final,omitempty"`
	Status         RentalStatus `json:"status"`
}
