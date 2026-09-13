package entities

import "errors"

// Erros de dominio. A camada de delivery os traduz em codigos de status HTTP.
// As mensagens chegam ao usuario final pela interface, entao seguem em
// portugues, sem acento, como o restante do projeto.
var (
	ErrNotFound           = errors.New("recurso nao encontrado")
	ErrInvalidData        = errors.New("dados invalidos")
	ErrBookingConflict    = errors.New("veiculo ja reservado no periodo informado")
	ErrVehicleUnavailable = errors.New("veiculo indisponivel para locacao")
	ErrDuplicatePlate     = errors.New("placa ja cadastrada para esta empresa")
	ErrRentalClosed       = errors.New("locacao ja encerrada")
)
