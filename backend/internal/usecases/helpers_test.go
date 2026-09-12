package usecases_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"driveflow/backend/internal/entities"
	"driveflow/backend/internal/repository"
	"driveflow/backend/internal/usecases"
)

// services reune os tres casos de uso montados sobre o mesmo repositorio,
// como acontece na composicao feita em cmd/app.
type services struct {
	companies *usecases.CompanyService
	vehicles  *usecases.VehicleService
	rentals   *usecases.RentalService
}

// newServices monta os servicos com id e relogio deterministicos.
func newServices(t *testing.T) services {
	t.Helper()

	counter := 0
	newID := func() string {
		counter++
		return fmt.Sprintf("id-%03d", counter)
	}
	now := func() time.Time { return time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC) }

	repo := repository.NewMemoryRepository()
	return services{
		companies: usecases.NewCompanyService(repo, newID, now),
		vehicles:  usecases.NewVehicleService(repo, newID),
		rentals:   usecases.NewRentalService(repo, newID, now),
	}
}

func day(d int) time.Time {
	return time.Date(2026, time.March, d, 10, 0, 0, 0, time.UTC)
}

func companyWithVehicle(t *testing.T, s services, name, taxID, plate string) (entities.Company, entities.Vehicle) {
	t.Helper()
	ctx := context.Background()

	company, err := s.companies.Register(ctx, name, taxID)
	if err != nil {
		t.Fatalf("Register empresa: %v", err)
	}
	vehicle, err := s.vehicles.Register(ctx, company.ID, plate, "Onix 1.0", "economico", 15000)
	if err != nil {
		t.Fatalf("Register veiculo: %v", err)
	}
	return company, vehicle
}
