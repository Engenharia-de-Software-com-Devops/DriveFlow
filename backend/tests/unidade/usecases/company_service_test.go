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

func TestListCompaniesIsEmptyThenSortedByName(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	empty, err := s.Companies.List(ctx)
	if err != nil {
		t.Fatalf("List vazia: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("lista inicial = %+v, esperada vazia", empty)
	}

	if _, err := s.Companies.Register(ctx, "Locadora Zeta", "98765432000121"); err != nil {
		t.Fatalf("Register zeta: %v", err)
	}
	if _, err := s.Companies.Register(ctx, "Locadora Alfa", "12345678000190"); err != nil {
		t.Fatalf("Register alfa: %v", err)
	}

	listed, err := s.Companies.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("len = %d, esperado 2", len(listed))
	}
	if listed[0].Name != "Locadora Alfa" || listed[1].Name != "Locadora Zeta" {
		t.Errorf("ordem = %q, %q; esperado Alfa, Zeta", listed[0].Name, listed[1].Name)
	}
}
