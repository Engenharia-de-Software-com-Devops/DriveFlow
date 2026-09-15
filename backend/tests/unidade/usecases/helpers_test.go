// Apoio dos testes de caso de uso: repositorio em memoria, relogio parado e o
// atalho de cadastro que quase todo teste aqui precisa antes de chegar na regra
// que esta sendo verificada.
package usecases_test

import (
	"context"
	"testing"
	"time"

	"driveflow/backend/internal/entities"
	"driveflow/backend/internal/repository"
	"driveflow/backend/tests/apoio"
)

// newServices monta os casos de uso sobre o repositorio em memoria, com id e
// relogio deterministicos.
func newServices(t *testing.T) apoio.Servicos {
	t.Helper()
	return apoio.Deterministico(repository.NewMemoryRepository())
}

func day(d int) time.Time {
	return time.Date(2026, time.March, d, 10, 0, 0, 0, time.UTC)
}

func companyWithVehicle(t *testing.T, s apoio.Servicos, name, taxID, plate string) (entities.Company, entities.Vehicle) {
	t.Helper()
	ctx := context.Background()

	company, err := s.Companies.Register(ctx, name, taxID)
	if err != nil {
		t.Fatalf("Register empresa: %v", err)
	}
	vehicle, err := s.Vehicles.Register(ctx, company.ID, plate, "Onix 1.0", "economico", 15000)
	if err != nil {
		t.Fatalf("Register veiculo: %v", err)
	}
	return company, vehicle
}
