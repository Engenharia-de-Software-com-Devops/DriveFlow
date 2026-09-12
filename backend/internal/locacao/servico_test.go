package locacao_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"driveflow/backend/internal/armazenamento"
	"driveflow/backend/internal/locacao"
)

func novoServico(t *testing.T) *locacao.Servico {
	t.Helper()

	contador := 0
	gerarID := func() string {
		contador++
		return fmt.Sprintf("id-%03d", contador)
	}
	agora := func() time.Time { return time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC) }

	return locacao.NovoServico(armazenamento.NovaMemoria(), gerarID, agora)
}

func dia(d int) time.Time {
	return time.Date(2026, time.March, d, 10, 0, 0, 0, time.UTC)
}

func empresaComVeiculo(t *testing.T, s *locacao.Servico, nome, cnpj, placa string) (locacao.Empresa, locacao.Veiculo) {
	t.Helper()
	ctx := context.Background()

	empresa, err := s.CadastrarEmpresa(ctx, nome, cnpj)
	if err != nil {
		t.Fatalf("CadastrarEmpresa: %v", err)
	}
	veiculo, err := s.CadastrarVeiculo(ctx, empresa.ID, placa, "Onix 1.0", "economico", 15000)
	if err != nil {
		t.Fatalf("CadastrarVeiculo: %v", err)
	}
	return empresa, veiculo
}

func TestCadastrarEmpresaValidaCNPJ(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	if _, err := s.CadastrarEmpresa(ctx, "Locadora Alfa", "123"); !errors.Is(err, locacao.ErrDadosInvalidos) {
		t.Fatalf("esperado ErrDadosInvalidos, obtido %v", err)
	}
	empresa, err := s.CadastrarEmpresa(ctx, "Locadora Alfa", "12.345.678/0001-90")
	if err != nil {
		t.Fatalf("CadastrarEmpresa: %v", err)
	}
	if empresa.CNPJ != "12345678000190" {
		t.Errorf("cnpj = %q, esperado apenas digitos", empresa.CNPJ)
	}
}

func TestCadastrarVeiculoValidaPlaca(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	empresa, err := s.CadastrarEmpresa(ctx, "Locadora Alfa", "12345678000190")
	if err != nil {
		t.Fatalf("CadastrarEmpresa: %v", err)
	}

	if _, err := s.CadastrarVeiculo(ctx, empresa.ID, "ABC12", "Onix", "economico", 15000); !errors.Is(err, locacao.ErrDadosInvalidos) {
		t.Fatalf("placa invalida deveria falhar, obtido %v", err)
	}

	// Placa no padrao Mercosul, informada com hifen e em minusculas.
	veiculo, err := s.CadastrarVeiculo(ctx, empresa.ID, "abc-1d23", "Onix", "economico", 15000)
	if err != nil {
		t.Fatalf("CadastrarVeiculo: %v", err)
	}
	if veiculo.Placa != "ABC1D23" {
		t.Errorf("placa = %q, esperado ABC1D23", veiculo.Placa)
	}
	if veiculo.Status != locacao.VeiculoDisponivel {
		t.Errorf("status = %q, esperado disponivel", veiculo.Status)
	}
}

func TestCadastrarVeiculoRejeitaPlacaDuplicada(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	empresa, _ := empresaComVeiculo(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")
	if _, err := s.CadastrarVeiculo(ctx, empresa.ID, "ABC1D23", "Onix", "economico", 15000); !errors.Is(err, locacao.ErrPlacaDuplicada) {
		t.Fatalf("esperado ErrPlacaDuplicada, obtido %v", err)
	}
}

// Gargalo 2 do diagnostico: reserva duplicada do mesmo veiculo.
func TestReservarVeiculoRejeitaPeriodoSobreposto(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	empresa, veiculo := empresaComVeiculo(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	if _, err := s.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente A", dia(10), dia(14)); err != nil {
		t.Fatalf("primeira reserva deveria ser aceita: %v", err)
	}

	_, err := s.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente B", dia(12), dia(16))
	if !errors.Is(err, locacao.ErrConflitoReserva) {
		t.Fatalf("esperado ErrConflitoReserva, obtido %v", err)
	}
}

func TestReservarVeiculoAceitaPeriodoSeguinte(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	empresa, veiculo := empresaComVeiculo(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	if _, err := s.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente A", dia(10), dia(14)); err != nil {
		t.Fatalf("primeira reserva: %v", err)
	}
	if _, err := s.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente B", dia(14), dia(16)); err != nil {
		t.Fatalf("reserva no periodo seguinte deveria ser aceita: %v", err)
	}
}

// Gargalo 2 do diagnostico: vazamento de frota entre empresas (multi-tenant).
func TestEmpresaNaoReservaVeiculoDeOutraEmpresa(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	_, veiculoAlfa := empresaComVeiculo(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")
	empresaBeta, _ := empresaComVeiculo(t, s, "Locadora Beta", "98765432000121", "XYZ9A88")

	_, err := s.ReservarVeiculo(ctx, empresaBeta.ID, veiculoAlfa.ID, "Cliente B", dia(10), dia(14))
	if !errors.Is(err, locacao.ErrNaoEncontrado) {
		t.Fatalf("esperado ErrNaoEncontrado, obtido %v", err)
	}
}

func TestListarFrotaIsolaPorEmpresa(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	empresaAlfa, _ := empresaComVeiculo(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")
	empresaBeta, _ := empresaComVeiculo(t, s, "Locadora Beta", "98765432000121", "XYZ9A88")

	frotaAlfa, err := s.ListarFrota(ctx, empresaAlfa.ID)
	if err != nil {
		t.Fatalf("ListarFrota: %v", err)
	}
	if len(frotaAlfa) != 1 || frotaAlfa[0].Placa != "ABC1D23" {
		t.Fatalf("frota da Alfa = %+v, esperado apenas ABC1D23", frotaAlfa)
	}

	frotaBeta, err := s.ListarFrota(ctx, empresaBeta.ID)
	if err != nil {
		t.Fatalf("ListarFrota: %v", err)
	}
	if len(frotaBeta) != 1 || frotaBeta[0].Placa != "XYZ9A88" {
		t.Fatalf("frota da Beta = %+v, esperado apenas XYZ9A88", frotaBeta)
	}
}

func TestReservarVeiculoRejeitaPeriodoInvalido(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	empresa, veiculo := empresaComVeiculo(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")
	if _, err := s.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente A", dia(14), dia(10)); !errors.Is(err, locacao.ErrDadosInvalidos) {
		t.Fatalf("esperado ErrDadosInvalidos, obtido %v", err)
	}
}

func TestDevolverVeiculoCalculaValorFinalComMulta(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	empresa, veiculo := empresaComVeiculo(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	contrato, err := s.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente A", dia(10), dia(13))
	if err != nil {
		t.Fatalf("ReservarVeiculo: %v", err)
	}
	if contrato.ValorPrevisto != 45000 {
		t.Fatalf("valor previsto = %d, esperado 45000", contrato.ValorPrevisto)
	}

	// Devolucao um dia apos o previsto: 45000 + 15000 + 30% de multa.
	encerrado, err := s.DevolverVeiculo(ctx, empresa.ID, contrato.ID, dia(14))
	if err != nil {
		t.Fatalf("DevolverVeiculo: %v", err)
	}
	if encerrado.Status != locacao.LocacaoEncerrada {
		t.Errorf("status = %q, esperado encerrada", encerrado.Status)
	}
	if encerrado.ValorFinal == nil || *encerrado.ValorFinal != 64500 {
		t.Errorf("valor final = %v, esperado 64500", encerrado.ValorFinal)
	}
}

func TestDevolverVeiculoLiberaParaNovaReserva(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	empresa, veiculo := empresaComVeiculo(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	contrato, err := s.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente A", dia(10), dia(14))
	if err != nil {
		t.Fatalf("ReservarVeiculo: %v", err)
	}
	if _, err := s.DevolverVeiculo(ctx, empresa.ID, contrato.ID, dia(12)); err != nil {
		t.Fatalf("DevolverVeiculo: %v", err)
	}
	if _, err := s.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente B", dia(11), dia(13)); err != nil {
		t.Fatalf("apos a devolucao o veiculo deveria estar livre: %v", err)
	}
}

func TestDevolverVeiculoDuasVezesFalha(t *testing.T) {
	s := novoServico(t)
	ctx := context.Background()

	empresa, veiculo := empresaComVeiculo(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	contrato, err := s.ReservarVeiculo(ctx, empresa.ID, veiculo.ID, "Cliente A", dia(10), dia(14))
	if err != nil {
		t.Fatalf("ReservarVeiculo: %v", err)
	}
	if _, err := s.DevolverVeiculo(ctx, empresa.ID, contrato.ID, dia(14)); err != nil {
		t.Fatalf("DevolverVeiculo: %v", err)
	}
	if _, err := s.DevolverVeiculo(ctx, empresa.ID, contrato.ID, dia(15)); !errors.Is(err, locacao.ErrLocacaoEncerrada) {
		t.Fatalf("esperado ErrLocacaoEncerrada, obtido %v", err)
	}
}
