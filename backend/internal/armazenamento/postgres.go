package armazenamento

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"driveflow/backend/internal/locacao"
)

// Postgres implementa locacao.Repositorio sobre o container db.
type Postgres struct {
	db *sql.DB
}

// Conectar abre o pool de conexoes e espera o banco aceitar consultas.
// O container da api pode subir antes do container do db, entao a conexao
// e tentada por ate `espera` antes de desistir.
func Conectar(ctx context.Context, url string, espera time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("abrir conexao: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	limite := time.Now().Add(espera)
	for {
		erroPing := db.PingContext(ctx)
		if erroPing == nil {
			return db, nil
		}
		if ctx.Err() != nil || time.Now().After(limite) {
			_ = db.Close()
			return nil, fmt.Errorf("banco indisponivel apos %s: %w", espera, erroPing)
		}
		select {
		case <-ctx.Done():
			_ = db.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// NovoPostgres envolve um pool ja conectado.
func NovoPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) CriarEmpresa(ctx context.Context, e locacao.Empresa) (locacao.Empresa, error) {
	const consulta = `
		INSERT INTO empresas (id, nome, cnpj, criada_em)
		VALUES ($1, $2, $3, $4)`

	_, err := p.db.ExecContext(ctx, consulta, e.ID, e.Nome, e.CNPJ, e.Criada)
	if err != nil {
		if violaRestricao(err, "empresas_cnpj_key") {
			return locacao.Empresa{}, fmt.Errorf("%w: cnpj ja cadastrado", locacao.ErrDadosInvalidos)
		}
		return locacao.Empresa{}, fmt.Errorf("inserir empresa: %w", err)
	}
	return e, nil
}

func (p *Postgres) BuscarEmpresa(ctx context.Context, id string) (locacao.Empresa, error) {
	const consulta = `SELECT id, nome, cnpj, criada_em FROM empresas WHERE id = $1`

	var e locacao.Empresa
	err := p.db.QueryRowContext(ctx, consulta, id).Scan(&e.ID, &e.Nome, &e.CNPJ, &e.Criada)
	if errors.Is(err, sql.ErrNoRows) {
		return locacao.Empresa{}, fmt.Errorf("%w: empresa %s", locacao.ErrNaoEncontrado, id)
	}
	if err != nil {
		return locacao.Empresa{}, fmt.Errorf("buscar empresa: %w", err)
	}
	return e, nil
}

func (p *Postgres) CriarVeiculo(ctx context.Context, v locacao.Veiculo) (locacao.Veiculo, error) {
	const consulta = `
		INSERT INTO veiculos (id, empresa_id, placa, modelo, categoria, tarifa_diaria, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := p.db.ExecContext(ctx, consulta,
		v.ID, v.EmpresaID, v.Placa, v.Modelo, v.Categoria, v.TarifaDiaria, string(v.Status))
	if err != nil {
		if violaRestricao(err, "veiculos_placa_por_empresa") {
			return locacao.Veiculo{}, locacao.ErrPlacaDuplicada
		}
		return locacao.Veiculo{}, fmt.Errorf("inserir veiculo: %w", err)
	}
	return v, nil
}

// BuscarVeiculo filtra por empresa_id na propria consulta: o isolamento entre
// tenants nao depende de o chamador lembrar de conferir a empresa.
func (p *Postgres) BuscarVeiculo(ctx context.Context, empresaID, veiculoID string) (locacao.Veiculo, error) {
	const consulta = `
		SELECT id, empresa_id, placa, modelo, categoria, tarifa_diaria, status
		FROM veiculos
		WHERE id = $1 AND empresa_id = $2`

	var v locacao.Veiculo
	var status string
	err := p.db.QueryRowContext(ctx, consulta, veiculoID, empresaID).
		Scan(&v.ID, &v.EmpresaID, &v.Placa, &v.Modelo, &v.Categoria, &v.TarifaDiaria, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return locacao.Veiculo{}, fmt.Errorf("%w: veiculo %s", locacao.ErrNaoEncontrado, veiculoID)
	}
	if err != nil {
		return locacao.Veiculo{}, fmt.Errorf("buscar veiculo: %w", err)
	}
	v.Status = locacao.StatusVeiculo(status)
	return v, nil
}

func (p *Postgres) ListarVeiculos(ctx context.Context, empresaID string) ([]locacao.Veiculo, error) {
	const consulta = `
		SELECT id, empresa_id, placa, modelo, categoria, tarifa_diaria, status
		FROM veiculos
		WHERE empresa_id = $1
		ORDER BY placa`

	linhas, err := p.db.QueryContext(ctx, consulta, empresaID)
	if err != nil {
		return nil, fmt.Errorf("listar veiculos: %w", err)
	}
	defer func() { _ = linhas.Close() }()

	frota := make([]locacao.Veiculo, 0)
	for linhas.Next() {
		var v locacao.Veiculo
		var status string
		if err := linhas.Scan(&v.ID, &v.EmpresaID, &v.Placa, &v.Modelo, &v.Categoria, &v.TarifaDiaria, &status); err != nil {
			return nil, fmt.Errorf("ler veiculo: %w", err)
		}
		v.Status = locacao.StatusVeiculo(status)
		frota = append(frota, v)
	}
	return frota, linhas.Err()
}

func (p *Postgres) CriarLocacao(ctx context.Context, l locacao.Locacao) (locacao.Locacao, error) {
	const consulta = `
		INSERT INTO locacoes (id, empresa_id, veiculo_id, cliente, inicio, fim_previsto, valor_previsto, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := p.db.ExecContext(ctx, consulta,
		l.ID, l.EmpresaID, l.VeiculoID, l.Cliente, l.Inicio, l.FimPrevisto, l.ValorPrevisto, string(l.Status))
	if err != nil {
		// Rede de seguranca contra duas requisicoes simultaneas: mesmo que as
		// duas passem pela checagem de dominio, o banco recusa a segunda.
		if violaRestricao(err, "locacoes_sem_sobreposicao") {
			return locacao.Locacao{}, fmt.Errorf("%w: periodo bloqueado pelo banco", locacao.ErrConflitoReserva)
		}
		return locacao.Locacao{}, fmt.Errorf("inserir locacao: %w", err)
	}
	return l, nil
}

func (p *Postgres) AtualizarLocacao(ctx context.Context, l locacao.Locacao) (locacao.Locacao, error) {
	const consulta = `
		UPDATE locacoes
		SET cliente = $3, inicio = $4, fim_previsto = $5,
		    devolvido_em = $6, valor_previsto = $7, valor_final = $8, status = $9
		WHERE id = $1 AND empresa_id = $2`

	resultado, err := p.db.ExecContext(ctx, consulta,
		l.ID, l.EmpresaID, l.Cliente, l.Inicio, l.FimPrevisto,
		nuloSeVazio(l.DevolvidoEm), l.ValorPrevisto, nuloSeNil(l.ValorFinal), string(l.Status))
	if err != nil {
		return locacao.Locacao{}, fmt.Errorf("atualizar locacao: %w", err)
	}

	afetadas, err := resultado.RowsAffected()
	if err != nil {
		return locacao.Locacao{}, fmt.Errorf("conferir atualizacao: %w", err)
	}
	if afetadas == 0 {
		return locacao.Locacao{}, fmt.Errorf("%w: locacao %s", locacao.ErrNaoEncontrado, l.ID)
	}
	return l, nil
}

func (p *Postgres) BuscarLocacao(ctx context.Context, empresaID, locacaoID string) (locacao.Locacao, error) {
	const consulta = colunasLocacao + ` WHERE id = $1 AND empresa_id = $2`

	l, err := lerLocacao(p.db.QueryRowContext(ctx, consulta, locacaoID, empresaID))
	if errors.Is(err, sql.ErrNoRows) {
		return locacao.Locacao{}, fmt.Errorf("%w: locacao %s", locacao.ErrNaoEncontrado, locacaoID)
	}
	if err != nil {
		return locacao.Locacao{}, fmt.Errorf("buscar locacao: %w", err)
	}
	return l, nil
}

func (p *Postgres) ListarLocacoes(ctx context.Context, empresaID string) ([]locacao.Locacao, error) {
	const consulta = colunasLocacao + ` WHERE empresa_id = $1 ORDER BY inicio`
	return p.consultarLocacoes(ctx, consulta, empresaID)
}

func (p *Postgres) LocacoesAtivasDoVeiculo(ctx context.Context, empresaID, veiculoID string) ([]locacao.Locacao, error) {
	const consulta = colunasLocacao + `
		WHERE empresa_id = $1 AND veiculo_id = $2 AND status = 'aberta'
		ORDER BY inicio`
	return p.consultarLocacoes(ctx, consulta, empresaID, veiculoID)
}

const colunasLocacao = `
	SELECT id, empresa_id, veiculo_id, cliente, inicio, fim_previsto,
	       devolvido_em, valor_previsto, valor_final, status
	FROM locacoes`

func (p *Postgres) consultarLocacoes(ctx context.Context, consulta string, args ...any) ([]locacao.Locacao, error) {
	linhas, err := p.db.QueryContext(ctx, consulta, args...)
	if err != nil {
		return nil, fmt.Errorf("consultar locacoes: %w", err)
	}
	defer func() { _ = linhas.Close() }()

	contratos := make([]locacao.Locacao, 0)
	for linhas.Next() {
		l, err := lerLocacao(linhas)
		if err != nil {
			return nil, fmt.Errorf("ler locacao: %w", err)
		}
		contratos = append(contratos, l)
	}
	return contratos, linhas.Err()
}

// leitor cobre tanto *sql.Row quanto *sql.Rows.
type leitor interface {
	Scan(destino ...any) error
}

func lerLocacao(origem leitor) (locacao.Locacao, error) {
	var l locacao.Locacao
	var devolvidoEm sql.NullTime
	var valorFinal sql.NullInt64
	var status string

	err := origem.Scan(&l.ID, &l.EmpresaID, &l.VeiculoID, &l.Cliente, &l.Inicio, &l.FimPrevisto,
		&devolvidoEm, &l.ValorPrevisto, &valorFinal, &status)
	if err != nil {
		return locacao.Locacao{}, err
	}

	if devolvidoEm.Valid {
		instante := devolvidoEm.Time.UTC()
		l.DevolvidoEm = &instante
	}
	if valorFinal.Valid {
		valor := valorFinal.Int64
		l.ValorFinal = &valor
	}
	l.Status = locacao.StatusLocacao(status)
	return l, nil
}

// violaRestricao identifica a constraint que o Postgres recusou.
func violaRestricao(err error, nome string) bool {
	var erroPG *pq.Error
	if errors.As(err, &erroPG) {
		return erroPG.Constraint == nome
	}
	return strings.Contains(err.Error(), nome)
}

func nuloSeVazio(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func nuloSeNil(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}
