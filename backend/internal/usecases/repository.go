// Package usecases aplica as regras de negocio da plataforma sobre as
// entidades de dominio.
//
// Depende apenas de entities e da porta Repository declarada aqui. Quem
// implementa a porta e a camada externa (internal/repository), o que mantem a
// direcao das dependencias apontando para dentro.
package usecases

import (
	"context"
	"time"

	"driveflow/backend/internal/entities"
	"driveflow/backend/pkg/id"
)

// Repository abstrai a persistencia do dominio. Os servicos dependem apenas
// desta interface, o que permite rodar a api com armazenamento em memoria
// (desenvolvimento e testes) ou com PostgreSQL (containers e producao).
type Repository interface {
	CreateCompany(ctx context.Context, c entities.Company) (entities.Company, error)
	FindCompany(ctx context.Context, companyID string) (entities.Company, error)
	ListCompanies(ctx context.Context) ([]entities.Company, error)

	CreateVehicle(ctx context.Context, v entities.Vehicle) (entities.Vehicle, error)
	FindVehicle(ctx context.Context, companyID, vehicleID string) (entities.Vehicle, error)
	ListVehicles(ctx context.Context, companyID string) ([]entities.Vehicle, error)

	CreateRental(ctx context.Context, r entities.Rental) (entities.Rental, error)
	UpdateRental(ctx context.Context, r entities.Rental) (entities.Rental, error)
	FindRental(ctx context.Context, companyID, rentalID string) (entities.Rental, error)
	ListRentals(ctx context.Context, companyID string) ([]entities.Rental, error)
	// ActiveRentalsForVehicle devolve os contratos ainda nao encerrados do
	// veiculo, usados para detectar conflito de reserva.
	ActiveRentalsForVehicle(ctx context.Context, companyID, vehicleID string) ([]entities.Rental, error)
}

// IDGenerator produz identificadores para os registros criados.
type IDGenerator func() string

// Clock fornece o instante atual. Injetado para tornar os testes deterministicos.
type Clock func() time.Time

// generatorOrDefault permite que os construtores recebam nil e ainda assim
// funcionem com a implementacao padrao.
func generatorOrDefault(newID IDGenerator) IDGenerator {
	if newID == nil {
		return id.New
	}
	return newID
}

func clockOrDefault(now Clock) Clock {
	if now == nil {
		return time.Now
	}
	return now
}
