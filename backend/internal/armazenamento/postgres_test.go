package armazenamento_test

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

	"driveflow/backend/internal/armazenamento"
	"driveflow/backend/internal/locacao"
)

// abrirPostgres conecta no banco indicado por DATABASE_URL e aplica as
// migrations. Sem a variavel definida o teste e ignorado, entao
// `go test ./...` continua rodando sem depender de container.
func abrirPostgres(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL nao definida: teste de integracao com o postgres ignorado")
	}

	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()

	db, err := armazenamento.Conectar(ctx, url, 20*time.Second)
	if err != nil {
		t.Fatalf("Conectar: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	silencioso := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := armazenamento.AplicarMigracoes(db, silencioso); err != nil {
		t.Fatalf("AplicarMigracoes: %v", err)
	}

	limpar(t, db)
	t.Cleanup(func() { limpar(t, db) })
	return db
}

func limpar(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`TRUNCATE locacoes, veiculos, empresas CASCADE`); err != nil {
		t.Fatalf("limpar tabelas: %v", err)
	}
}

func diaPG(d int) time.Time {
	return time.Date(2026, time.June, d, 10, 0, 0, 0, time.UTC)
}

// AplicarMigracoes precisa ser idempotente: o container da api roda na subida
// e pode reiniciar quantas vezes for preciso.
func TestAplicarMigracoesEhIdempotente(t *testing.T) {
	db := abrirPostgres(t)
	silencioso := slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := armazenamento.AplicarMigracoes(db, silencioso); err != nil {
		t.Fatalf("segunda aplicacao das migracoes falhou: %v", err)
	}

	var total int
	if err := db.QueryRow(`SELECT count(*) FROM schema_migracoes`).Scan(&total); err != nil {
		t.Fatalf("contar migracoes: %v", err)
	}

	esperadas, err := armazenamento.CarregarMigracoes()
	if err != nil {
		t.Fatalf("CarregarMigracoes: %v", err)
	}
	if total != len(esperadas) {
		t.Errorf("schema_migracoes tem %d linhas, esperado %d", total, len(esperadas))
	}
}

func TestPostgresFluxoCompleto(t *testing.T) {
	repo := armazenamento.NovoPostgres(abrirPostgres(t))
	servico := locacao.NovoServico(repo, nil, nil)
	ctx := context.Background()

	empresa, err := servico.CadastrarEmpresa(ctx, "Locadora Alfa", "12345678000190")
	if err != nil {
		t.Fatalf("CadastrarEmpresa: %v", err)
	}
	veiculo, err := servico.CadastrarVeiculo(ctx, empresa.ID, "ABC1D23", "Onix 1.0", "economico", 15000)
	if err != nil {
		t.Fatalf("CadastrarVeiculo: %v", err)
	}

	contrato, err := servico.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente A", diaPG(10), diaPG(13))
	if err != nil {
		t.Fatalf("ReservarVeiculo: %v", err)
	}
	if contrato.ValorPrevisto != 45000 {
		t.Errorf("valor previsto = %d, esperado 45000", contrato.ValorPrevisto)
	}

	if _, err := servico.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente B", diaPG(12), diaPG(16)); !errors.Is(err, locacao.ErrConflitoReserva) {
		t.Fatalf("esperado ErrConflitoReserva, obtido %v", err)
	}

	encerrado, err := servico.DevolverVeiculo(ctx, empresa.ID, contrato.ID, diaPG(14))
	if err != nil {
		t.Fatalf("DevolverVeiculo: %v", err)
	}
	if encerrado.ValorFinal == nil || *encerrado.ValorFinal != 64500 {
		t.Errorf("valor final = %v, esperado 64500", encerrado.ValorFinal)
	}

	contratos, err := servico.ListarLocacoes(ctx, empresa.ID)
	if err != nil {
		t.Fatalf("ListarLocacoes: %v", err)
	}
	if len(contratos) != 1 || contratos[0].Status != locacao.LocacaoEncerrada {
		t.Errorf("locacoes = %+v, esperado um contrato encerrado", contratos)
	}
}

// O banco precisa recusar contratos sobrepostos mesmo quando a checagem de
// dominio e contornada, cobrindo a corrida entre duas requisicoes simultaneas.
func TestPostgresBloqueiaSobreposicaoNoBanco(t *testing.T) {
	db := abrirPostgres(t)
	repo := armazenamento.NovoPostgres(db)
	ctx := context.Background()

	empresa, err := repo.CriarEmpresa(ctx, locacao.Empresa{
		ID: "emp-1", Nome: "Locadora Alfa", CNPJ: "12345678000190", Criada: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("CriarEmpresa: %v", err)
	}
	veiculo, err := repo.CriarVeiculo(ctx, locacao.Veiculo{
		ID: "vei-1", EmpresaID: empresa.ID, Placa: "ABC1D23", Modelo: "Onix",
		Categoria: "economico", TarifaDiaria: 15000, Status: locacao.VeiculoDisponivel,
	})
	if err != nil {
		t.Fatalf("CriarVeiculo: %v", err)
	}

	primeira := locacao.Locacao{
		ID: "loc-1", EmpresaID: empresa.ID, VeiculoID: veiculo.ID, Cliente: "Cliente A",
		Inicio: diaPG(10), FimPrevisto: diaPG(14), ValorPrevisto: 60000, Status: locacao.LocacaoAberta,
	}
	if _, err := repo.CriarLocacao(ctx, primeira); err != nil {
		t.Fatalf("primeira locacao: %v", err)
	}

	segunda := primeira
	segunda.ID = "loc-2"
	segunda.Cliente = "Cliente B"
	segunda.Inicio = diaPG(12)
	segunda.FimPrevisto = diaPG(16)

	if _, err := repo.CriarLocacao(ctx, segunda); !errors.Is(err, locacao.ErrConflitoReserva) {
		t.Fatalf("esperado ErrConflitoReserva vindo do banco, obtido %v", err)
	}
}

func TestPostgresIsolaTenants(t *testing.T) {
	repo := armazenamento.NovoPostgres(abrirPostgres(t))
	servico := locacao.NovoServico(repo, nil, nil)
	ctx := context.Background()

	alfa, err := servico.CadastrarEmpresa(ctx, "Locadora Alfa", "12345678000190")
	if err != nil {
		t.Fatalf("CadastrarEmpresa alfa: %v", err)
	}
	beta, err := servico.CadastrarEmpresa(ctx, "Locadora Beta", "98765432000121")
	if err != nil {
		t.Fatalf("CadastrarEmpresa beta: %v", err)
	}

	veiculoAlfa, err := servico.CadastrarVeiculo(ctx, alfa.ID, "ABC1D23", "Onix", "economico", 15000)
	if err != nil {
		t.Fatalf("CadastrarVeiculo: %v", err)
	}

	if _, err := servico.ReservarVeiculo(ctx, beta.ID, veiculoAlfa.ID, "Cliente B", diaPG(10), diaPG(12)); !errors.Is(err, locacao.ErrNaoEncontrado) {
		t.Fatalf("esperado ErrNaoEncontrado, obtido %v", err)
	}

	frotaBeta, err := servico.ListarFrota(ctx, beta.ID)
	if err != nil {
		t.Fatalf("ListarFrota: %v", err)
	}
	if len(frotaBeta) != 0 {
		t.Errorf("frota da Beta = %+v, esperada vazia", frotaBeta)
	}
}
