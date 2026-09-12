// Package repository reune as implementacoes da porta usecases.Repository.
//
// E uma camada externa: conhece entities e o banco, mas nenhum caso de uso
// depende de um tipo daqui.
package repository

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"driveflow/backend/internal/entities"
)

// MemoryRepository guarda os dados em estruturas na propria aplicacao.
// E a implementacao usada nos testes e no modo de desenvolvimento sem banco,
// o que mantem `go test ./...` reproduzivel sem depender de containers.
type MemoryRepository struct {
	mu        sync.RWMutex
	companies map[string]entities.Company
	vehicles  map[string]entities.Vehicle
	rentals   map[string]entities.Rental
}

// NewMemoryRepository devolve um repositorio em memoria vazio.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		companies: make(map[string]entities.Company),
		vehicles:  make(map[string]entities.Vehicle),
		rentals:   make(map[string]entities.Rental),
	}
}

func (m *MemoryRepository) CreateCompany(_ context.Context, c entities.Company) (entities.Company, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.companies {
		if existing.TaxID == c.TaxID {
			return entities.Company{}, fmt.Errorf("%w: cnpj ja cadastrado", entities.ErrInvalidData)
		}
	}
	m.companies[c.ID] = c
	return c, nil
}

func (m *MemoryRepository) FindCompany(_ context.Context, companyID string) (entities.Company, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.companies[companyID]
	if !ok {
		return entities.Company{}, fmt.Errorf("%w: empresa %s", entities.ErrNotFound, companyID)
	}
	return c, nil
}

func (m *MemoryRepository) CreateVehicle(_ context.Context, v entities.Vehicle) (entities.Vehicle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.vehicles {
		if existing.CompanyID == v.CompanyID && existing.Plate == v.Plate {
			return entities.Vehicle{}, entities.ErrDuplicatePlate
		}
	}
	m.vehicles[v.ID] = v
	return v, nil
}

// FindVehicle so encontra o veiculo se ele pertencer a empresa informada.
// E o ponto que garante o isolamento entre tenants.
func (m *MemoryRepository) FindVehicle(_ context.Context, companyID, vehicleID string) (entities.Vehicle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.vehicles[vehicleID]
	if !ok || v.CompanyID != companyID {
		return entities.Vehicle{}, fmt.Errorf("%w: veiculo %s", entities.ErrNotFound, vehicleID)
	}
	return v, nil
}

func (m *MemoryRepository) ListVehicles(_ context.Context, companyID string) ([]entities.Vehicle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fleet := make([]entities.Vehicle, 0)
	for _, v := range m.vehicles {
		if v.CompanyID == companyID {
			fleet = append(fleet, v)
		}
	}
	sort.Slice(fleet, func(i, j int) bool { return fleet[i].Plate < fleet[j].Plate })
	return fleet, nil
}

func (m *MemoryRepository) CreateRental(_ context.Context, r entities.Rental) (entities.Rental, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rentals[r.ID] = r
	return r, nil
}

func (m *MemoryRepository) UpdateRental(_ context.Context, r entities.Rental) (entities.Rental, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, ok := m.rentals[r.ID]
	if !ok || current.CompanyID != r.CompanyID {
		return entities.Rental{}, fmt.Errorf("%w: locacao %s", entities.ErrNotFound, r.ID)
	}
	m.rentals[r.ID] = r
	return r, nil
}

func (m *MemoryRepository) FindRental(_ context.Context, companyID, rentalID string) (entities.Rental, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.rentals[rentalID]
	if !ok || r.CompanyID != companyID {
		return entities.Rental{}, fmt.Errorf("%w: locacao %s", entities.ErrNotFound, rentalID)
	}
	return r, nil
}

func (m *MemoryRepository) ListRentals(_ context.Context, companyID string) ([]entities.Rental, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]entities.Rental, 0)
	for _, r := range m.rentals {
		if r.CompanyID == companyID {
			list = append(list, r)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Start.Before(list[j].Start) })
	return list, nil
}

func (m *MemoryRepository) ActiveRentalsForVehicle(_ context.Context, companyID, vehicleID string) ([]entities.Rental, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]entities.Rental, 0)
	for _, r := range m.rentals {
		if r.CompanyID == companyID && r.VehicleID == vehicleID && r.Status == entities.RentalOpen {
			list = append(list, r)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Start.Before(list[j].Start) })
	return list, nil
}
