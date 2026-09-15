package usecases

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"driveflow/backend/internal/entities"
)

// plateFormat aceita o padrao antigo (ABC1234) e o padrao Mercosul (ABC1D23).
var plateFormat = regexp.MustCompile(`^[A-Z]{3}[0-9][0-9A-Z][0-9]{2}$`)

// VehicleService cuida da frota de cada empresa.
type VehicleService struct {
	repo  Repository
	newID IDGenerator
}

// NewVehicleService monta o servico. newID pode ser nil, caso em que assume a
// implementacao padrao.
func NewVehicleService(repo Repository, newID IDGenerator) *VehicleService {
	return &VehicleService{repo: repo, newID: generatorOrDefault(newID)}
}

// Register adiciona um veiculo a frota de uma empresa.
func (s *VehicleService) Register(ctx context.Context, companyID, plate, model, category string, dailyRate int64) (entities.Vehicle, error) {
	if _, err := s.repo.FindCompany(ctx, companyID); err != nil {
		return entities.Vehicle{}, err
	}

	plate = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(plate), "-", ""))
	model = strings.TrimSpace(model)
	category = strings.TrimSpace(category)

	if !plateFormat.MatchString(plate) {
		return entities.Vehicle{}, fmt.Errorf("%w: placa %q fora do padrao brasileiro", entities.ErrInvalidData, plate)
	}
	if model == "" {
		return entities.Vehicle{}, fmt.Errorf("%w: modelo e obrigatorio", entities.ErrInvalidData)
	}
	if category == "" {
		return entities.Vehicle{}, fmt.Errorf("%w: categoria e obrigatoria", entities.ErrInvalidData)
	}
	if dailyRate <= 0 {
		return entities.Vehicle{}, fmt.Errorf("%w: tarifa diaria deve ser maior que zero", entities.ErrInvalidData)
	}

	return s.repo.CreateVehicle(ctx, entities.Vehicle{
		ID:        s.newID(),
		CompanyID: companyID,
		Plate:     plate,
		Model:     model,
		Category:  category,
		DailyRate: dailyRate,
		Status:    entities.VehicleAvailable,
	})
}

// ListFleet devolve os veiculos de uma empresa.
func (s *VehicleService) ListFleet(ctx context.Context, companyID string) ([]entities.Vehicle, error) {
	if _, err := s.repo.FindCompany(ctx, companyID); err != nil {
		return nil, err
	}
	return s.repo.ListVehicles(ctx, companyID)
}
