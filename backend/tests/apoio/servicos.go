// Package apoio reune a montagem que os testes de unidade e de integracao
// compartilham. Fica fora de arquivo _test.go de proposito: pacote de teste nao
// enxerga o _test.go de outro pacote, e sem isto cada nivel manteria a sua
// propria copia dos tres casos de uso.
package apoio

import (
	"fmt"
	"time"

	"driveflow/backend/internal/usecases"
)

// Servicos reune os tres casos de uso montados sobre o mesmo repositorio, como
// acontece na composicao feita em cmd/app.
type Servicos struct {
	Companies *usecases.CompanyService
	Vehicles  *usecases.VehicleService
	Rentals   *usecases.RentalService
}

// Deterministico monta os casos de uso com id sequencial e relogio parado, para
// o teste poder afirmar id e data exatos.
func Deterministico(repo usecases.Repository) Servicos {
	counter := 0
	newID := func() string {
		counter++
		return fmt.Sprintf("id-%03d", counter)
	}
	now := func() time.Time { return time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC) }

	return Servicos{
		Companies: usecases.NewCompanyService(repo, newID, now),
		Vehicles:  usecases.NewVehicleService(repo, newID),
		Rentals:   usecases.NewRentalService(repo, newID, now),
	}
}

// Padrao monta os casos de uso com o id e o relogio de producao (nil deixa cada
// servico usar o seu proprio default). E o que a integracao com o banco precisa:
// cada teste grava linhas com id unico e data real.
func Padrao(repo usecases.Repository) Servicos {
	return Servicos{
		Companies: usecases.NewCompanyService(repo, nil, nil),
		Vehicles:  usecases.NewVehicleService(repo, nil),
		Rentals:   usecases.NewRentalService(repo, nil, nil),
	}
}
