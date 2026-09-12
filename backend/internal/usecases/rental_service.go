package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"driveflow/backend/internal/entities"
)

// RentalService cuida do ciclo de vida dos contratos de locacao.
type RentalService struct {
	repo  Repository
	newID IDGenerator
	now   Clock
}

// NewRentalService monta o servico. newID e now podem ser nil, caso em que
// assumem as implementacoes padrao.
func NewRentalService(repo Repository, newID IDGenerator, now Clock) *RentalService {
	return &RentalService{
		repo:  repo,
		newID: generatorOrDefault(newID),
		now:   clockOrDefault(now),
	}
}

// Reserve cria um contrato de locacao.
//
// Esta e a regra critica apontada no diagnostico do E1: duas empresas (ou dois
// atendentes da mesma empresa) nao podem reservar o mesmo veiculo em periodos
// que se sobrepoem.
func (s *RentalService) Reserve(ctx context.Context, companyID, vehicleID, customer string, start, expectedEnd time.Time) (entities.Rental, error) {
	customer = strings.TrimSpace(customer)
	if customer == "" {
		return entities.Rental{}, fmt.Errorf("%w: cliente e obrigatorio", entities.ErrInvalidData)
	}
	if !expectedEnd.After(start) {
		return entities.Rental{}, fmt.Errorf("%w: fim previsto deve ser posterior ao inicio", entities.ErrInvalidData)
	}

	// FindVehicle filtra por empresa: uma empresa nunca alcanca a frota da outra.
	vehicle, err := s.repo.FindVehicle(ctx, companyID, vehicleID)
	if err != nil {
		return entities.Rental{}, err
	}
	if vehicle.Status != entities.VehicleAvailable {
		return entities.Rental{}, fmt.Errorf("%w: veiculo em %s", entities.ErrVehicleUnavailable, vehicle.Status)
	}

	active, err := s.repo.ActiveRentalsForVehicle(ctx, companyID, vehicleID)
	if err != nil {
		return entities.Rental{}, err
	}
	for _, rental := range active {
		if entities.PeriodsOverlap(start, expectedEnd, rental.Start, rental.ExpectedEnd) {
			return entities.Rental{}, fmt.Errorf("%w: conflito com a locacao %s", entities.ErrBookingConflict, rental.ID)
		}
	}

	return s.repo.CreateRental(ctx, entities.Rental{
		ID:             s.newID(),
		CompanyID:      companyID,
		VehicleID:      vehicleID,
		Customer:       customer,
		Start:          start.UTC(),
		ExpectedEnd:    expectedEnd.UTC(),
		EstimatedTotal: entities.EstimatedTotal(vehicle.DailyRate, start, expectedEnd),
		Status:         entities.RentalOpen,
	})
}

// Return encerra o contrato e calcula o valor final com eventual multa.
// Um returnedAt zerado significa devolucao agora.
func (s *RentalService) Return(ctx context.Context, companyID, rentalID string, returnedAt time.Time) (entities.Rental, error) {
	rental, err := s.repo.FindRental(ctx, companyID, rentalID)
	if err != nil {
		return entities.Rental{}, err
	}
	if rental.Status == entities.RentalClosed {
		return entities.Rental{}, entities.ErrRentalClosed
	}

	vehicle, err := s.repo.FindVehicle(ctx, companyID, rental.VehicleID)
	if err != nil {
		return entities.Rental{}, err
	}

	if returnedAt.IsZero() {
		returnedAt = s.now()
	}
	returnedAt = returnedAt.UTC()
	if returnedAt.Before(rental.Start) {
		return entities.Rental{}, fmt.Errorf("%w: devolucao anterior ao inicio da locacao", entities.ErrInvalidData)
	}

	total := entities.FinalTotal(vehicle.DailyRate, rental.Start, rental.ExpectedEnd, returnedAt)
	rental.ReturnedAt = &returnedAt
	rental.FinalTotal = &total
	rental.Status = entities.RentalClosed

	return s.repo.UpdateRental(ctx, rental)
}

// List devolve os contratos de uma empresa.
func (s *RentalService) List(ctx context.Context, companyID string) ([]entities.Rental, error) {
	if _, err := s.repo.FindCompany(ctx, companyID); err != nil {
		return nil, err
	}
	return s.repo.ListRentals(ctx, companyID)
}
