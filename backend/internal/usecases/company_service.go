package usecases

import (
	"context"
	"fmt"
	"strings"

	"driveflow/backend/internal/entities"
)

// CompanyService cuida do cadastro dos tenants da plataforma.
type CompanyService struct {
	repo  Repository
	newID IDGenerator
	now   Clock
}

// NewCompanyService monta o servico. newID e now podem ser nil, caso em que
// assumem as implementacoes padrao.
func NewCompanyService(repo Repository, newID IDGenerator, now Clock) *CompanyService {
	return &CompanyService{
		repo:  repo,
		newID: generatorOrDefault(newID),
		now:   clockOrDefault(now),
	}
}

// Register registra uma nova empresa na plataforma.
func (s *CompanyService) Register(ctx context.Context, name, taxID string) (entities.Company, error) {
	name = strings.TrimSpace(name)
	taxID = onlyDigits(taxID)

	if name == "" {
		return entities.Company{}, fmt.Errorf("%w: nome da empresa e obrigatorio", entities.ErrInvalidData)
	}
	if len(taxID) != 14 {
		return entities.Company{}, fmt.Errorf("%w: cnpj deve conter 14 digitos", entities.ErrInvalidData)
	}

	return s.repo.CreateCompany(ctx, entities.Company{
		ID:        s.newID(),
		Name:      name,
		TaxID:     taxID,
		CreatedAt: s.now().UTC(),
	})
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
