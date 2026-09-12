// Package armazenamento reune as implementacoes de persistencia do dominio.
package armazenamento

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"driveflow/backend/internal/locacao"
)

// Memoria guarda os dados em estruturas na propria aplicacao.
// E a implementacao usada nos testes e no modo de desenvolvimento sem banco,
// o que mantem `go test ./...` reproduzivel sem depender de containers.
type Memoria struct {
	mu       sync.RWMutex
	empresas map[string]locacao.Empresa
	veiculos map[string]locacao.Veiculo
	locacoes map[string]locacao.Locacao
}

// NovaMemoria devolve um repositorio em memoria vazio.
func NovaMemoria() *Memoria {
	return &Memoria{
		empresas: make(map[string]locacao.Empresa),
		veiculos: make(map[string]locacao.Veiculo),
		locacoes: make(map[string]locacao.Locacao),
	}
}

func (m *Memoria) CriarEmpresa(_ context.Context, e locacao.Empresa) (locacao.Empresa, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existente := range m.empresas {
		if existente.CNPJ == e.CNPJ {
			return locacao.Empresa{}, fmt.Errorf("%w: cnpj ja cadastrado", locacao.ErrDadosInvalidos)
		}
	}
	m.empresas[e.ID] = e
	return e, nil
}

func (m *Memoria) BuscarEmpresa(_ context.Context, id string) (locacao.Empresa, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	e, ok := m.empresas[id]
	if !ok {
		return locacao.Empresa{}, fmt.Errorf("%w: empresa %s", locacao.ErrNaoEncontrado, id)
	}
	return e, nil
}

func (m *Memoria) CriarVeiculo(_ context.Context, v locacao.Veiculo) (locacao.Veiculo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existente := range m.veiculos {
		if existente.EmpresaID == v.EmpresaID && existente.Placa == v.Placa {
			return locacao.Veiculo{}, locacao.ErrPlacaDuplicada
		}
	}
	m.veiculos[v.ID] = v
	return v, nil
}

// BuscarVeiculo so encontra o veiculo se ele pertencer a empresa informada.
// E o ponto que garante o isolamento entre tenants.
func (m *Memoria) BuscarVeiculo(_ context.Context, empresaID, veiculoID string) (locacao.Veiculo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.veiculos[veiculoID]
	if !ok || v.EmpresaID != empresaID {
		return locacao.Veiculo{}, fmt.Errorf("%w: veiculo %s", locacao.ErrNaoEncontrado, veiculoID)
	}
	return v, nil
}

func (m *Memoria) ListarVeiculos(_ context.Context, empresaID string) ([]locacao.Veiculo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lista := make([]locacao.Veiculo, 0)
	for _, v := range m.veiculos {
		if v.EmpresaID == empresaID {
			lista = append(lista, v)
		}
	}
	sort.Slice(lista, func(i, j int) bool { return lista[i].Placa < lista[j].Placa })
	return lista, nil
}

func (m *Memoria) CriarLocacao(_ context.Context, l locacao.Locacao) (locacao.Locacao, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.locacoes[l.ID] = l
	return l, nil
}

func (m *Memoria) AtualizarLocacao(_ context.Context, l locacao.Locacao) (locacao.Locacao, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	atual, ok := m.locacoes[l.ID]
	if !ok || atual.EmpresaID != l.EmpresaID {
		return locacao.Locacao{}, fmt.Errorf("%w: locacao %s", locacao.ErrNaoEncontrado, l.ID)
	}
	m.locacoes[l.ID] = l
	return l, nil
}

func (m *Memoria) BuscarLocacao(_ context.Context, empresaID, locacaoID string) (locacao.Locacao, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	l, ok := m.locacoes[locacaoID]
	if !ok || l.EmpresaID != empresaID {
		return locacao.Locacao{}, fmt.Errorf("%w: locacao %s", locacao.ErrNaoEncontrado, locacaoID)
	}
	return l, nil
}

func (m *Memoria) ListarLocacoes(_ context.Context, empresaID string) ([]locacao.Locacao, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lista := make([]locacao.Locacao, 0)
	for _, l := range m.locacoes {
		if l.EmpresaID == empresaID {
			lista = append(lista, l)
		}
	}
	sort.Slice(lista, func(i, j int) bool { return lista[i].Inicio.Before(lista[j].Inicio) })
	return lista, nil
}

func (m *Memoria) LocacoesAtivasDoVeiculo(_ context.Context, empresaID, veiculoID string) ([]locacao.Locacao, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lista := make([]locacao.Locacao, 0)
	for _, l := range m.locacoes {
		if l.EmpresaID == empresaID && l.VeiculoID == veiculoID && l.Status == locacao.LocacaoAberta {
			lista = append(lista, l)
		}
	}
	sort.Slice(lista, func(i, j int) bool { return lista[i].Inicio.Before(lista[j].Inicio) })
	return lista, nil
}
