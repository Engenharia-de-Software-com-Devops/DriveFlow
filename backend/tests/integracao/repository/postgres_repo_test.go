//go:build integracao

// Testes de integracao do repositorio PostgreSQL: sobem contra um banco real,
// entao so compilam com a tag `integracao` e ficam fora do `go test ./...` do
// dia a dia. Use `make testar-integracao`, que sobe o container e define a
// DATABASE_URL. Teste que roda sem banco vai em tests/unidade.
package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"driveflow/backend/internal/entities"
	"driveflow/backend/internal/repository"
	"driveflow/backend/tests/apoio"
)

// openPostgres conecta no banco indicado por DATABASE_URL e aplica as
// migrations. Como estes testes so compilam com a tag `integracao`, pedir a
// suite sem banco e erro de uso, nao motivo para ignorar em silencio.
func openPostgres(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL nao definida: rode `make testar-integracao`, que sobe o postgres do compose")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := repository.Connect(ctx, url, 20*time.Second)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	quiet := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := repository.ApplyMigrations(db, quiet); err != nil {
		t.Fatalf("ApplyMigrations: %v", err)
	}

	truncate(t, db)
	t.Cleanup(func() { truncate(t, db) })
	return db
}

func truncate(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`TRUNCATE locacoes, veiculos, empresas CASCADE`); err != nil {
		t.Fatalf("limpar tabelas: %v", err)
	}
}

func pgDay(d int) time.Time {
	return time.Date(2026, time.June, d, 10, 0, 0, 0, time.UTC)
}

// ApplyMigrations precisa ser idempotente: o container da api roda na subida
// e pode reiniciar quantas vezes for preciso.
func TestApplyMigrationsIsIdempotent(t *testing.T) {
	db := openPostgres(t)
	quiet := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := repository.ApplyMigrations(db, quiet); err != nil {
		t.Fatalf("segunda aplicacao das migracoes falhou: %v", err)
	}

	var total int
	if err := db.QueryRow(`SELECT count(*) FROM schema_migracoes`).Scan(&total); err != nil {
		t.Fatalf("contar migracoes: %v", err)
	}

	expected, err := repository.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations: %v", err)
	}
	if total != len(expected) {
		t.Errorf("schema_migracoes tem %d linhas, esperado %d", total, len(expected))
	}
}

func TestPostgresFullFlow(t *testing.T) {
	s := apoio.Padrao(repository.NewPostgresRepository(openPostgres(t)))
	ctx := context.Background()

	company, err := s.Companies.Register(ctx, "Locadora Alfa", "12345678000190")
	if err != nil {
		t.Fatalf("Register empresa: %v", err)
	}
	vehicle, err := s.Vehicles.Register(ctx, company.ID, "ABC1D23", "Onix 1.0", "economico", 15000)
	if err != nil {
		t.Fatalf("Register veiculo: %v", err)
	}

	rental, err := s.Rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente A", pgDay(10), pgDay(13))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if rental.EstimatedTotal != 45000 {
		t.Errorf("valor previsto = %d, esperado 45000", rental.EstimatedTotal)
	}

	if _, err := s.Rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente B", pgDay(12), pgDay(16)); !errors.Is(err, entities.ErrBookingConflict) {
		t.Fatalf("esperado ErrBookingConflict, obtido %v", err)
	}

	closed, err := s.Rentals.Return(ctx, company.ID, rental.ID, pgDay(14))
	if err != nil {
		t.Fatalf("Return: %v", err)
	}
	if closed.FinalTotal == nil || *closed.FinalTotal != 64500 {
		t.Errorf("valor final = %v, esperado 64500", closed.FinalTotal)
	}

	rentals, err := s.Rentals.List(ctx, company.ID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(rentals) != 1 || rentals[0].Status != entities.RentalClosed {
		t.Errorf("locacoes = %+v, esperado um contrato encerrado", rentals)
	}
}

// O banco precisa recusar contratos sobrepostos mesmo quando a checagem de
// dominio e contornada, cobrindo a corrida entre duas requisicoes simultaneas.
func TestPostgresBlocksOverlapInDatabase(t *testing.T) {
	repo := repository.NewPostgresRepository(openPostgres(t))
	ctx := context.Background()

	company, err := repo.CreateCompany(ctx, entities.Company{
		ID: "emp-1", Name: "Locadora Alfa", TaxID: "12345678000190", CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	vehicle, err := repo.CreateVehicle(ctx, entities.Vehicle{
		ID: "vei-1", CompanyID: company.ID, Plate: "ABC1D23", Model: "Onix",
		Category: "economico", DailyRate: 15000, Status: entities.VehicleAvailable,
	})
	if err != nil {
		t.Fatalf("CreateVehicle: %v", err)
	}

	first := entities.Rental{
		ID: "loc-1", CompanyID: company.ID, VehicleID: vehicle.ID, Customer: "Cliente A",
		Start: pgDay(10), ExpectedEnd: pgDay(14), EstimatedTotal: 60000, Status: entities.RentalOpen,
	}
	if _, err := repo.CreateRental(ctx, first); err != nil {
		t.Fatalf("primeira locacao: %v", err)
	}

	second := first
	second.ID = "loc-2"
	second.Customer = "Cliente B"
	second.Start = pgDay(12)
	second.ExpectedEnd = pgDay(16)

	if _, err := repo.CreateRental(ctx, second); !errors.Is(err, entities.ErrBookingConflict) {
		t.Fatalf("esperado ErrBookingConflict vindo do banco, obtido %v", err)
	}
}

func TestPostgresIsolatesTenants(t *testing.T) {
	s := apoio.Padrao(repository.NewPostgresRepository(openPostgres(t)))
	ctx := context.Background()

	alfa, err := s.Companies.Register(ctx, "Locadora Alfa", "12345678000190")
	if err != nil {
		t.Fatalf("Register alfa: %v", err)
	}
	beta, err := s.Companies.Register(ctx, "Locadora Beta", "98765432000121")
	if err != nil {
		t.Fatalf("Register beta: %v", err)
	}

	vehicleAlfa, err := s.Vehicles.Register(ctx, alfa.ID, "ABC1D23", "Onix", "economico", 15000)
	if err != nil {
		t.Fatalf("Register veiculo: %v", err)
	}

	if _, err := s.Rentals.Reserve(ctx, beta.ID, vehicleAlfa.ID, "Cliente B", pgDay(10), pgDay(12)); !errors.Is(err, entities.ErrNotFound) {
		t.Fatalf("esperado ErrNotFound, obtido %v", err)
	}

	fleetBeta, err := s.Vehicles.ListFleet(ctx, beta.ID)
	if err != nil {
		t.Fatalf("ListFleet: %v", err)
	}
	if len(fleetBeta) != 0 {
		t.Errorf("frota da Beta = %+v, esperada vazia", fleetBeta)
	}
}
