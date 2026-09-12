package locacao

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// formatoPlaca aceita o padrao antigo (ABC1234) e o padrao Mercosul (ABC1D23).
var formatoPlaca = regexp.MustCompile(`^[A-Z]{3}[0-9][0-9A-Z][0-9]{2}$`)

// Servico aplica as regras de negocio da plataforma sobre o repositorio.
type Servico struct {
	repo    Repositorio
	gerarID GeradorID
	agora   Relogio
}

// NovoServico monta o servico. gerarID e agora podem ser nil, caso em que
// assumem as implementacoes padrao.
func NovoServico(repo Repositorio, gerarID GeradorID, agora Relogio) *Servico {
	if gerarID == nil {
		gerarID = IDAleatorio
	}
	if agora == nil {
		agora = time.Now
	}
	return &Servico{repo: repo, gerarID: gerarID, agora: agora}
}

// CadastrarEmpresa registra um novo tenant na plataforma.
func (s *Servico) CadastrarEmpresa(ctx context.Context, nome, cnpj string) (Empresa, error) {
	nome = strings.TrimSpace(nome)
	cnpj = apenasDigitos(cnpj)

	if nome == "" {
		return Empresa{}, fmt.Errorf("%w: nome da empresa e obrigatorio", ErrDadosInvalidos)
	}
	if len(cnpj) != 14 {
		return Empresa{}, fmt.Errorf("%w: cnpj deve conter 14 digitos", ErrDadosInvalidos)
	}

	return s.repo.CriarEmpresa(ctx, Empresa{
		ID:     s.gerarID(),
		Nome:   nome,
		CNPJ:   cnpj,
		Criada: s.agora().UTC(),
	})
}

// CadastrarVeiculo adiciona um veiculo a frota de uma empresa.
func (s *Servico) CadastrarVeiculo(ctx context.Context, empresaID, placa, modelo, categoria string, tarifaDiaria int64) (Veiculo, error) {
	if _, err := s.repo.BuscarEmpresa(ctx, empresaID); err != nil {
		return Veiculo{}, err
	}

	placa = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(placa), "-", ""))
	modelo = strings.TrimSpace(modelo)
	categoria = strings.TrimSpace(categoria)

	if !formatoPlaca.MatchString(placa) {
		return Veiculo{}, fmt.Errorf("%w: placa %q fora do padrao brasileiro", ErrDadosInvalidos, placa)
	}
	if modelo == "" {
		return Veiculo{}, fmt.Errorf("%w: modelo e obrigatorio", ErrDadosInvalidos)
	}
	if categoria == "" {
		return Veiculo{}, fmt.Errorf("%w: categoria e obrigatoria", ErrDadosInvalidos)
	}
	if tarifaDiaria <= 0 {
		return Veiculo{}, fmt.Errorf("%w: tarifa diaria deve ser maior que zero", ErrDadosInvalidos)
	}

	return s.repo.CriarVeiculo(ctx, Veiculo{
		ID:           s.gerarID(),
		EmpresaID:    empresaID,
		Placa:        placa,
		Modelo:       modelo,
		Categoria:    categoria,
		TarifaDiaria: tarifaDiaria,
		Status:       VeiculoDisponivel,
	})
}

// ListarFrota devolve os veiculos de uma empresa.
func (s *Servico) ListarFrota(ctx context.Context, empresaID string) ([]Veiculo, error) {
	if _, err := s.repo.BuscarEmpresa(ctx, empresaID); err != nil {
		return nil, err
	}
	return s.repo.ListarVeiculos(ctx, empresaID)
}

// ReservarVeiculo cria um contrato de locacao.
//
// Esta e a regra critica apontada no diagnostico do E1: duas empresas (ou dois
// atendentes da mesma empresa) nao podem reservar o mesmo veiculo em periodos
// que se sobrepoem.
func (s *Servico) ReservarVeiculo(ctx context.Context, empresaID, veiculoID, cliente string, inicio, fimPrevisto time.Time) (Locacao, error) {
	cliente = strings.TrimSpace(cliente)
	if cliente == "" {
		return Locacao{}, fmt.Errorf("%w: cliente e obrigatorio", ErrDadosInvalidos)
	}
	if !fimPrevisto.After(inicio) {
		return Locacao{}, fmt.Errorf("%w: fim previsto deve ser posterior ao inicio", ErrDadosInvalidos)
	}

	// BuscarVeiculo filtra por empresa: uma empresa nunca alcanca a frota da outra.
	veiculo, err := s.repo.BuscarVeiculo(ctx, empresaID, veiculoID)
	if err != nil {
		return Locacao{}, err
	}
	if veiculo.Status != VeiculoDisponivel {
		return Locacao{}, fmt.Errorf("%w: veiculo em %s", ErrVeiculoIndisponivel, veiculo.Status)
	}

	ativas, err := s.repo.LocacoesAtivasDoVeiculo(ctx, empresaID, veiculoID)
	if err != nil {
		return Locacao{}, err
	}
	for _, ativa := range ativas {
		if PeriodosSobrepostos(inicio, fimPrevisto, ativa.Inicio, ativa.FimPrevisto) {
			return Locacao{}, fmt.Errorf("%w: conflito com a locacao %s", ErrConflitoReserva, ativa.ID)
		}
	}

	return s.repo.CriarLocacao(ctx, Locacao{
		ID:            s.gerarID(),
		EmpresaID:     empresaID,
		VeiculoID:     veiculoID,
		Cliente:       cliente,
		Inicio:        inicio.UTC(),
		FimPrevisto:   fimPrevisto.UTC(),
		ValorPrevisto: ValorPrevisto(veiculo.TarifaDiaria, inicio, fimPrevisto),
		Status:        LocacaoAberta,
	})
}

// DevolverVeiculo encerra o contrato e calcula o valor final com eventual multa.
func (s *Servico) DevolverVeiculo(ctx context.Context, empresaID, locacaoID string, devolucao time.Time) (Locacao, error) {
	contrato, err := s.repo.BuscarLocacao(ctx, empresaID, locacaoID)
	if err != nil {
		return Locacao{}, err
	}
	if contrato.Status == LocacaoEncerrada {
		return Locacao{}, ErrLocacaoEncerrada
	}

	veiculo, err := s.repo.BuscarVeiculo(ctx, empresaID, contrato.VeiculoID)
	if err != nil {
		return Locacao{}, err
	}

	if devolucao.IsZero() {
		devolucao = s.agora()
	}
	devolucao = devolucao.UTC()
	if devolucao.Before(contrato.Inicio) {
		return Locacao{}, fmt.Errorf("%w: devolucao anterior ao inicio da locacao", ErrDadosInvalidos)
	}

	valor := ValorFinal(veiculo.TarifaDiaria, contrato.Inicio, contrato.FimPrevisto, devolucao)
	contrato.DevolvidoEm = &devolucao
	contrato.ValorFinal = &valor
	contrato.Status = LocacaoEncerrada

	return s.repo.AtualizarLocacao(ctx, contrato)
}

// ListarLocacoes devolve os contratos de uma empresa.
func (s *Servico) ListarLocacoes(ctx context.Context, empresaID string) ([]Locacao, error) {
	if _, err := s.repo.BuscarEmpresa(ctx, empresaID); err != nil {
		return nil, err
	}
	return s.repo.ListarLocacoes(ctx, empresaID)
}

func apenasDigitos(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
