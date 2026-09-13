package usecases_test

import (
	"context"
	"errors"
	"testing"

	"driveflow/backend/internal/entities"
)

func TestRegisterCompanyValidatesTaxID(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	if _, err := s.Companies.Register(ctx, "Locadora Alfa", "123"); !errors.Is(err, entities.ErrInvalidData) {
		t.Fatalf("esperado ErrInvalidData, obtido %v", err)
	}
	company, err := s.Companies.Register(ctx, "Locadora Alfa", "12.345.678/0001-90")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if company.TaxID != "12345678000190" {
		t.Errorf("cnpj = %q, esperado apenas digitos", company.TaxID)
	}
}
