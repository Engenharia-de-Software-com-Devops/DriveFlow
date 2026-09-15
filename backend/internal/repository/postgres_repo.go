package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"driveflow/backend/internal/entities"
)

// PostgresRepository implementa usecases.Repository sobre o container db.
//
// As tabelas e colunas continuam em portugues: o schema ja foi aplicado e,
// pela convencao do projeto, migration aplicada nao e editada. A traducao para
// os campos em ingles acontece aqui, na fronteira com o banco.
type PostgresRepository struct {
	db *sql.DB
}

// Connect abre o pool de conexoes e espera o banco aceitar consultas.
// O container da api pode subir antes do container do db, entao a conexao
// e tentada por ate `wait` antes de desistir.
func Connect(ctx context.Context, url string, wait time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("abrir conexao: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	deadline := time.Now().Add(wait)
	for {
		pingErr := db.PingContext(ctx)
		if pingErr == nil {
			return db, nil
		}
		if ctx.Err() != nil || time.Now().After(deadline) {
			_ = db.Close()
			return nil, fmt.Errorf("banco indisponivel apos %s: %w", wait, pingErr)
		}
		select {
		case <-ctx.Done():
			_ = db.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// NewPostgresRepository envolve um pool ja conectado.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) CreateCompany(ctx context.Context, c entities.Company) (entities.Company, error) {
	const query = `
		INSERT INTO empresas (id, nome, cnpj, criada_em)
		VALUES ($1, $2, $3, $4)`

	_, err := p.db.ExecContext(ctx, query, c.ID, c.Name, c.TaxID, c.CreatedAt)
	if err != nil {
		if violatesConstraint(err, "empresas_cnpj_key") {
			return entities.Company{}, fmt.Errorf("%w: cnpj ja cadastrado", entities.ErrInvalidData)
		}
		return entities.Company{}, fmt.Errorf("inserir empresa: %w", err)
	}
	return c, nil
}

func (p *PostgresRepository) FindCompany(ctx context.Context, companyID string) (entities.Company, error) {
	const query = `SELECT id, nome, cnpj, criada_em FROM empresas WHERE id = $1`

	var c entities.Company
	err := p.db.QueryRowContext(ctx, query, companyID).Scan(&c.ID, &c.Name, &c.TaxID, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entities.Company{}, fmt.Errorf("%w: empresa %s", entities.ErrNotFound, companyID)
	}
	if err != nil {
		return entities.Company{}, fmt.Errorf("buscar empresa: %w", err)
	}
	return c, nil
}

func (p *PostgresRepository) CreateVehicle(ctx context.Context, v entities.Vehicle) (entities.Vehicle, error) {
	const query = `
		INSERT INTO veiculos (id, empresa_id, placa, modelo, categoria, tarifa_diaria, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := p.db.ExecContext(ctx, query,
		v.ID, v.CompanyID, v.Plate, v.Model, v.Category, v.DailyRate, string(v.Status))
	if err != nil {
		if violatesConstraint(err, "veiculos_placa_por_empresa") {
			return entities.Vehicle{}, entities.ErrDuplicatePlate
		}
		return entities.Vehicle{}, fmt.Errorf("inserir veiculo: %w", err)
	}
	return v, nil
}

// FindVehicle filtra por empresa_id na propria consulta: o isolamento entre
// tenants nao depende de o chamador lembrar de conferir a empresa.
func (p *PostgresRepository) FindVehicle(ctx context.Context, companyID, vehicleID string) (entities.Vehicle, error) {
	const query = vehicleColumns + ` WHERE id = $1 AND empresa_id = $2`

	var v entities.Vehicle
	var status string
	err := p.db.QueryRowContext(ctx, query, vehicleID, companyID).
		Scan(&v.ID, &v.CompanyID, &v.Plate, &v.Model, &v.Category, &v.DailyRate, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return entities.Vehicle{}, fmt.Errorf("%w: veiculo %s", entities.ErrNotFound, vehicleID)
	}
	if err != nil {
		return entities.Vehicle{}, fmt.Errorf("buscar veiculo: %w", err)
	}
	v.Status = entities.VehicleStatus(status)
	return v, nil
}

func (p *PostgresRepository) ListVehicles(ctx context.Context, companyID string) ([]entities.Vehicle, error) {
	const query = vehicleColumns + ` WHERE empresa_id = $1 ORDER BY placa`

	rows, err := p.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("listar veiculos: %w", err)
	}
	defer func() { _ = rows.Close() }()

	fleet := make([]entities.Vehicle, 0)
	for rows.Next() {
		var v entities.Vehicle
		var status string
		if err := rows.Scan(&v.ID, &v.CompanyID, &v.Plate, &v.Model, &v.Category, &v.DailyRate, &status); err != nil {
			return nil, fmt.Errorf("ler veiculo: %w", err)
		}
		v.Status = entities.VehicleStatus(status)
		fleet = append(fleet, v)
	}
	return fleet, rows.Err()
}

func (p *PostgresRepository) CreateRental(ctx context.Context, r entities.Rental) (entities.Rental, error) {
	const query = `
		INSERT INTO locacoes (id, empresa_id, veiculo_id, cliente, inicio, fim_previsto, valor_previsto, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := p.db.ExecContext(ctx, query,
		r.ID, r.CompanyID, r.VehicleID, r.Customer, r.Start, r.ExpectedEnd, r.EstimatedTotal, string(r.Status))
	if err != nil {
		// Rede de seguranca contra duas requisicoes simultaneas: mesmo que as
		// duas passem pela checagem de dominio, o banco recusa a segunda.
		if violatesConstraint(err, "locacoes_sem_sobreposicao") {
			return entities.Rental{}, fmt.Errorf("%w: periodo bloqueado pelo banco", entities.ErrBookingConflict)
		}
		return entities.Rental{}, fmt.Errorf("inserir locacao: %w", err)
	}
	return r, nil
}

func (p *PostgresRepository) UpdateRental(ctx context.Context, r entities.Rental) (entities.Rental, error) {
	const query = `
		UPDATE locacoes
		SET cliente = $3, inicio = $4, fim_previsto = $5,
		    devolvido_em = $6, valor_previsto = $7, valor_final = $8, status = $9
		WHERE id = $1 AND empresa_id = $2`

	result, err := p.db.ExecContext(ctx, query,
		r.ID, r.CompanyID, r.Customer, r.Start, r.ExpectedEnd,
		nullIfNoTime(r.ReturnedAt), r.EstimatedTotal, nullIfNil(r.FinalTotal), string(r.Status))
	if err != nil {
		return entities.Rental{}, fmt.Errorf("atualizar locacao: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return entities.Rental{}, fmt.Errorf("conferir atualizacao: %w", err)
	}
	if affected == 0 {
		return entities.Rental{}, fmt.Errorf("%w: locacao %s", entities.ErrNotFound, r.ID)
	}
	return r, nil
}

func (p *PostgresRepository) FindRental(ctx context.Context, companyID, rentalID string) (entities.Rental, error) {
	const query = rentalColumns + ` WHERE id = $1 AND empresa_id = $2`

	r, err := scanRental(p.db.QueryRowContext(ctx, query, rentalID, companyID))
	if errors.Is(err, sql.ErrNoRows) {
		return entities.Rental{}, fmt.Errorf("%w: locacao %s", entities.ErrNotFound, rentalID)
	}
	if err != nil {
		return entities.Rental{}, fmt.Errorf("buscar locacao: %w", err)
	}
	return r, nil
}

func (p *PostgresRepository) ListRentals(ctx context.Context, companyID string) ([]entities.Rental, error) {
	const query = rentalColumns + ` WHERE empresa_id = $1 ORDER BY inicio`
	return p.queryRentals(ctx, query, companyID)
}

func (p *PostgresRepository) ActiveRentalsForVehicle(ctx context.Context, companyID, vehicleID string) ([]entities.Rental, error) {
	const query = rentalColumns + `
		WHERE empresa_id = $1 AND veiculo_id = $2 AND status = 'aberta'
		ORDER BY inicio`
	return p.queryRentals(ctx, query, companyID, vehicleID)
}

const vehicleColumns = `
	SELECT id, empresa_id, placa, modelo, categoria, tarifa_diaria, status
	FROM veiculos`

const rentalColumns = `
	SELECT id, empresa_id, veiculo_id, cliente, inicio, fim_previsto,
	       devolvido_em, valor_previsto, valor_final, status
	FROM locacoes`

func (p *PostgresRepository) queryRentals(ctx context.Context, query string, args ...any) ([]entities.Rental, error) {
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("consultar locacoes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	rentals := make([]entities.Rental, 0)
	for rows.Next() {
		r, err := scanRental(rows)
		if err != nil {
			return nil, fmt.Errorf("ler locacao: %w", err)
		}
		rentals = append(rentals, r)
	}
	return rentals, rows.Err()
}

// scanner cobre tanto *sql.Row quanto *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanRental(source scanner) (entities.Rental, error) {
	var r entities.Rental
	var returnedAt sql.NullTime
	var finalTotal sql.NullInt64
	var status string

	err := source.Scan(&r.ID, &r.CompanyID, &r.VehicleID, &r.Customer, &r.Start, &r.ExpectedEnd,
		&returnedAt, &r.EstimatedTotal, &finalTotal, &status)
	if err != nil {
		return entities.Rental{}, err
	}

	if returnedAt.Valid {
		instant := returnedAt.Time.UTC()
		r.ReturnedAt = &instant
	}
	if finalTotal.Valid {
		total := finalTotal.Int64
		r.FinalTotal = &total
	}
	r.Status = entities.RentalStatus(status)
	return r, nil
}

// violatesConstraint identifica a constraint que o Postgres recusou.
func violatesConstraint(err error, name string) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Constraint == name
	}
	return strings.Contains(err.Error(), name)
}

func nullIfNoTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func nullIfNil(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}
