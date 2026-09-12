package usecases_test

import (
	"context"
	"errors"
	"testing"

	"driveflow/backend/internal/entities"
)

func TestRegisterVehicleValidatesPlate(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	company, err := s.Companies.Register(ctx, "Locadora Alfa", "12345678000190")
	if err != nil {
		t.Fatalf("Register empresa: %v", err)
	}

	if _, err := s.Vehicles.Register(ctx, company.ID, "ABC12", "Onix", "economico", 15000); !errors.Is(err, entities.ErrInvalidData) {
		t.Fatalf("placa invalida deveria falhar, obtido %v", err)
	}

	// Placa no padrao Mercosul, informada com hifen e em minusculas.
	vehicle, err := s.Vehicles.Register(ctx, company.ID, "abc-1d23", "Onix", "economico", 15000)
	if err != nil {
		t.Fatalf("Register veiculo: %v", err)
	}
	if vehicle.Plate != "ABC1D23" {
		t.Errorf("placa = %q, esperado ABC1D23", vehicle.Plate)
	}
	if vehicle.Status != entities.VehicleAvailable {
		t.Errorf("status = %q, esperado disponivel", vehicle.Status)
	}
}

func TestRegisterVehicleRejectsDuplicatePlate(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	company, _ := companyWithVehicle(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")
	if _, err := s.Vehicles.Register(ctx, company.ID, "ABC1D23", "Onix", "economico", 15000); !errors.Is(err, entities.ErrDuplicatePlate) {
		t.Fatalf("esperado ErrDuplicatePlate, obtido %v", err)
	}
}

func TestListFleetIsolatesByCompany(t *testing.T) {
	s := newServices(t)
	ctx := context.Background()

	alfa, _ := companyWithVehicle(t, s, "Locadora Alfa", "12345678000190", "ABC1D23")
	beta, _ := companyWithVehicle(t, s, "Locadora Beta", "98765432000121", "XYZ9A88")

	fleetAlfa, err := s.Vehicles.ListFleet(ctx, alfa.ID)
	if err != nil {
		t.Fatalf("ListFleet: %v", err)
	}
	if len(fleetAlfa) != 1 || fleetAlfa[0].Plate != "ABC1D23" {
		t.Fatalf("frota da Alfa = %+v, esperado apenas ABC1D23", fleetAlfa)
	}

	fleetBeta, err := s.Vehicles.ListFleet(ctx, beta.ID)
	if err != nil {
		t.Fatalf("ListFleet: %v", err)
	}
	if len(fleetBeta) != 1 || fleetBeta[0].Plate != "XYZ9A88" {
		t.Fatalf("frota da Beta = %+v, esperado apenas XYZ9A88", fleetBeta)
	}
}
