package repository_test

import (
	"context"
	"testing"
	"time"

	"driveflow/backend/internal/entities"
	"driveflow/backend/internal/repository"
)

func TestMemoryListCompaniesSortedByName(t *testing.T) {
	repo := repository.NewMemoryRepository()
	ctx := context.Background()
	now := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)

	empty, err := repo.ListCompanies(ctx)
	if err != nil {
		t.Fatalf("ListCompanies vazia: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("lista inicial = %+v, esperada vazia", empty)
	}

	if _, err := repo.CreateCompany(ctx, entities.Company{
		ID: "emp-2", Name: "Locadora Zeta", TaxID: "98765432000121", CreatedAt: now,
	}); err != nil {
		t.Fatalf("CreateCompany zeta: %v", err)
	}
	if _, err := repo.CreateCompany(ctx, entities.Company{
		ID: "emp-1", Name: "Locadora Alfa", TaxID: "12345678000190", CreatedAt: now,
	}); err != nil {
		t.Fatalf("CreateCompany alfa: %v", err)
	}

	listed, err := repo.ListCompanies(ctx)
	if err != nil {
		t.Fatalf("ListCompanies: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("len = %d, esperado 2", len(listed))
	}
	if listed[0].Name != "Locadora Alfa" || listed[1].Name != "Locadora Zeta" {
		t.Errorf("ordem = %q, %q; esperado Alfa, Zeta", listed[0].Name, listed[1].Name)
	}
}
