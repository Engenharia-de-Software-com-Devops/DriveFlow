package usecases_test

import (
	"context"
	"errors"
	"testing"

	"driveflow/backend/internal/entities"
)

// Gargalo 2 do diagnostico: reserva duplicada do mesmo veiculo.
func TestReserveRejectsOverlappingPeriod(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	company, vehicle := companyWithVehicle(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	if _, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente A", day(10), day(14)); err != nil {
		t.Fatalf("primeira reserva deveria ser aceita: %v", err)
	}

	_, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente B", day(12), day(16))
	if !errors.Is(err, entities.ErrBookingConflict) {
		t.Fatalf("esperado ErrBookingConflict, obtido %v", err)
	}
}

func TestReserveAcceptsFollowingPeriod(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	company, vehicle := companyWithVehicle(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	if _, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente A", day(10), day(14)); err != nil {
		t.Fatalf("primeira reserva: %v", err)
	}
	if _, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente B", day(14), day(16)); err != nil {
		t.Fatalf("reserva no periodo seguinte deveria ser aceita: %v", err)
	}
}

// Gargalo 2 do diagnostico: vazamento de frota entre empresas (multi-tenant).
func TestCompanyCannotReserveVehicleFromAnotherCompany(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	_, vehicleAlfa := companyWithVehicle(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")
	beta, _ := companyWithVehicle(t, s, "Locadora Beta", "98765432000121", "XYZ9A88")

	_, err := s.rentals.Reserve(ctx, beta.ID, vehicleAlfa.ID, "Cliente B", day(10), day(14))
	if !errors.Is(err, entities.ErrNotFound) {
		t.Fatalf("esperado ErrNotFound, obtido %v", err)
	}
}

func TestReserveRejectsInvalidPeriod(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	company, vehicle := companyWithVehicle(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")
	if _, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente A", day(14), day(10)); !errors.Is(err, entities.ErrInvalidData) {
		t.Fatalf("esperado ErrInvalidData, obtido %v", err)
	}
}

func TestReturnCalculatesFinalTotalWithLateFee(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	company, vehicle := companyWithVehicle(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	rental, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente A", day(10), day(13))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if rental.EstimatedTotal != 45000 {
		t.Fatalf("valor previsto = %d, esperado 45000", rental.EstimatedTotal)
	}

	// Devolucao um dia apos o previsto: 45000 + 15000 + 30% de multa.
	closed, err := s.rentals.Return(ctx, company.ID, rental.ID, day(14))
	if err != nil {
		t.Fatalf("Return: %v", err)
	}
	if closed.Status != entities.RentalClosed {
		t.Errorf("status = %q, esperado encerrada", closed.Status)
	}
	if closed.FinalTotal == nil || *closed.FinalTotal != 64500 {
		t.Errorf("valor final = %v, esperado 64500", closed.FinalTotal)
	}
}

func TestReturnReleasesVehicleForNewReservation(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	company, vehicle := companyWithVehicle(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	rental, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente A", day(10), day(14))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if _, err := s.rentals.Return(ctx, company.ID, rental.ID, day(12)); err != nil {
		t.Fatalf("Return: %v", err)
	}
	if _, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente B", day(11), day(13)); err != nil {
		t.Fatalf("apos a devolucao o veiculo deveria estar livre: %v", err)
	}
}

func TestReturnTwiceFails(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	company, vehicle := companyWithVehicle(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")

	rental, err := s.rentals.Reserve(ctx, company.ID, vehicle.ID, "Cliente A", day(10), day(14))
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if _, err := s.rentals.Return(ctx, company.ID, rental.ID, day(14)); err != nil {
		t.Fatalf("Return: %v", err)
	}
	if _, err := s.rentals.Return(ctx, company.ID, rental.ID, day(15)); !errors.Is(err, entities.ErrRentalClosed) {
		t.Fatalf("esperado ErrRentalClosed, obtido %v", err)
	}
}
