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
	"driveflow/backend/internal/usecases"
)

// openPostgres conecta no banco indicado por DATABASE_URL e aplica as
// migrations. Sem a variavel definida o teste e ignorado, entao
// `go test ./...` continua rodando sem depender de container.
func openPostgres(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL nao definida: teste de integracao com o postgres ignorado")
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

// pgServices monta os casos de uso sobre o repositorio PostgreSQL.
func pgServices(repo usecases.Repository) services {
	return services{
		companies: usecases.NewCompanyService(repo, nil, nil),
		vehicles:  usecases.NewVehicleService(repo, nil),
		rentals:   usecases.NewRentalService(repo, nil, nil),
	}
}

type services struct {
	companies *usecases.CompanyService
	vehicles  *usecases.VehicleService
	rentals   *usecases.RentalService
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
	s := pgServices(repository.NewPostgresRepository(openPostgres(t)))
	ctx := context.Background()

	company, err := s.companies.Register(ctx, "Locadora Alfa", "12345678000190")
	if err != nil {
		t.Fatalf("Register empresa: %v", err)
	}
	vehicle, err := s.vehicles.Register(ctx, company.ID, "ABC1D23", "Onix 1.0", "economico", 15000)
	if err != nil {
		t.Fatalf("Register veiculo: %v", err)
	}

	rental, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente A", pgDay(10), pgDay(13))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if rental.EstimatedTotal != 45000 {
		t.Errorf("valor previsto = %d, esperado 45000", rental.EstimatedTotal)
	}

	if _, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente B", pgDay(12), pgDay(16)); !errors.Is(err, entities.ErrBookingConflict) {
		t.Fatalf("esperado ErrBookingConflict, obtido %v", err)
	}

	closed, err := s.rentals.Return(ctx, company.ID, rental.ID, pgDay(14))
	if err != nil {
		t.Fatalf("Return: %v", err)
	}
	if closed.FinalTotal == nil || *closed.FinalTotal != 64500 {
		t.Errorf("valor final = %v, esperado 64500", closed.FinalTotal)
	}

	rentals, err := s.rentals.List(ctx, company.ID)
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
	s := pgServices(repository.NewPostgresRepository(openPostgres(t)))
	ctx := context.Background()

	alfa, err := s.companies.Register(ctx, "Locadora Alfa", "12345678000190")
	if err != nil {
		t.Fatalf("Register alfa: %v", err)
	}
	beta, err := s.companies.Register(ctx, "Locadora Beta", "98765432000121")
	if err != nil {
		t.Fatalf("Register beta: %v", err)
	}

	vehicleAlfa, err := s.vehicles.Register(ctx, alfa.ID, "ABC1D23", "Onix", "economico", 15000)
	if err != nil {
		t.Fatalf("Register veiculo: %v", err)
	}

	if _, err := s.rentals.Reserve(ctx, beta.ID, vehicleAlfa.ID, "Cliente B", pgDay(10), pgDay(12)); !errors.Is(err, entities.ErrNotFound) {
		t.Fatalf("esperado ErrNotFound, obtido %v", err)
	}

	fleetBeta, err := s.vehicles.ListFleet(ctx, beta.ID)
	if err != nil {
		t.Fatalf("ListFleet: %v", err)
	}
	if len(fleetBeta) != 0 {
		t.Errorf("frota da Beta = %+v, esperada vazia", fleetBeta)
	}
}
