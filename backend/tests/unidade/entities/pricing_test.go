// Testes das regras de tarifa: aritmetica pura de diaria, multa e sobreposicao
// de periodo, sem repositorio nem banco.
package entities_test

import (
	"testing"
	"time"

	"driveflow/backend/internal/entities"
)

func at(day int, hour int) time.Time {
	return time.Date(2026, time.March, day, hour, 0, 0, 0, time.UTC)
}

func TestBillableDays(t *testing.T) {
	cases := []struct {
		name   string
		start  time.Time
		end    time.Time
		expect int64
	}{
		{"periodo de um dia exato", at(1, 8), at(2, 8), 1},
		{"fracao de dia vira diaria cheia", at(1, 8), at(2, 9), 2},
		{"tres dias exatos", at(1, 0), at(4, 0), 3},
		{"periodo de poucas horas cobra uma diaria", at(1, 8), at(1, 11), 1},
		{"periodo invertido cobra o minimo", at(4, 0), at(1, 0), 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := entities.BillableDays(c.start, c.end); got != c.expect {
				t.Errorf("entities.BillableDays() = %d, esperado %d", got, c.expect)
			}
		})
	}
}

func TestEstimatedTotal(t *testing.T) {
	// Tarifa de R$ 150,00 por 3 diarias = R$ 450,00.
	if got := entities.EstimatedTotal(15000, at(1, 0), at(4, 0)); got != 45000 {
		t.Errorf("entities.EstimatedTotal() = %d, esperado 45000", got)
	}
}

func TestFinalTotalSemAtraso(t *testing.T) {
	// Devolucao no prazo cobra apenas o previsto.
	if got := entities.FinalTotal(15000, at(1, 0), at(4, 0), at(4, 0)); got != 45000 {
		t.Errorf("entities.FinalTotal() = %d, esperado 45000", got)
	}
}

func TestFinalTotalDevolucaoAntecipadaNaoReduz(t *testing.T) {
	if got := entities.FinalTotal(15000, at(1, 0), at(4, 0), at(2, 0)); got != 45000 {
		t.Errorf("entities.FinalTotal() = %d, esperado 45000", got)
	}
}

func TestFinalTotalComAtraso(t *testing.T) {
	// 3 diarias previstas (45000) + 1 diaria de atraso (15000) + multa de 30% (4500).
	if got := entities.FinalTotal(15000, at(1, 0), at(4, 0), at(5, 0)); got != 64500 {
		t.Errorf("entities.FinalTotal() = %d, esperado 64500", got)
	}
}

func TestPeriodsOverlap(t *testing.T) {
	cases := []struct {
		name   string
		a1, a2 time.Time
		b1, b2 time.Time
		expect bool
	}{
		{"periodos identicos", at(1, 0), at(4, 0), at(1, 0), at(4, 0), true},
		{"reserva contida na outra", at(1, 0), at(10, 0), at(3, 0), at(5, 0), true},
		{"sobreposicao parcial", at(1, 0), at(4, 0), at(3, 0), at(6, 0), true},
		{"limites encostados liberam o veiculo", at(1, 0), at(4, 0), at(4, 0), at(6, 0), false},
		{"periodos distantes", at(1, 0), at(2, 0), at(8, 0), at(9, 0), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := entities.PeriodsOverlap(c.a1, c.a2, c.b1, c.b2); got != c.expect {
				t.Errorf("entities.PeriodsOverlap() = %v, esperado %v", got, c.expect)
			}
		})
	}
}
