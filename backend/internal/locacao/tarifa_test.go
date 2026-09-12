package locacao

import (
	"testing"
	"time"
)

func data(dia int, hora int) time.Time {
	return time.Date(2026, time.March, dia, hora, 0, 0, 0, time.UTC)
}

func TestDiarias(t *testing.T) {
	casos := []struct {
		nome     string
		inicio   time.Time
		fim      time.Time
		esperado int64
	}{
		{"periodo de um dia exato", data(1, 8), data(2, 8), 1},
		{"fracao de dia vira diaria cheia", data(1, 8), data(2, 9), 2},
		{"tres dias exatos", data(1, 0), data(4, 0), 3},
		{"periodo de poucas horas cobra uma diaria", data(1, 8), data(1, 11), 1},
		{"periodo invertido cobra o minimo", data(4, 0), data(1, 0), 1},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			if obtido := Diarias(caso.inicio, caso.fim); obtido != caso.esperado {
				t.Errorf("Diarias() = %d, esperado %d", obtido, caso.esperado)
			}
		})
	}
}

func TestValorPrevisto(t *testing.T) {
	// Tarifa de R$ 150,00 por 3 diarias = R$ 450,00.
	if obtido := ValorPrevisto(15000, data(1, 0), data(4, 0)); obtido != 45000 {
		t.Errorf("ValorPrevisto() = %d, esperado 45000", obtido)
	}
}

func TestValorFinalSemAtraso(t *testing.T) {
	// Devolucao no prazo cobra apenas o previsto.
	if obtido := ValorFinal(15000, data(1, 0), data(4, 0), data(4, 0)); obtido != 45000 {
		t.Errorf("ValorFinal() = %d, esperado 45000", obtido)
	}
}

func TestValorFinalDevolucaoAntecipadaNaoReduz(t *testing.T) {
	if obtido := ValorFinal(15000, data(1, 0), data(4, 0), data(2, 0)); obtido != 45000 {
		t.Errorf("ValorFinal() = %d, esperado 45000", obtido)
	}
}

func TestValorFinalComAtraso(t *testing.T) {
	// 3 diarias previstas (45000) + 1 diaria de atraso (15000) + multa de 30% (4500).
	if obtido := ValorFinal(15000, data(1, 0), data(4, 0), data(5, 0)); obtido != 64500 {
		t.Errorf("ValorFinal() = %d, esperado 64500", obtido)
	}
}

func TestPeriodosSobrepostos(t *testing.T) {
	casos := []struct {
		nome     string
		a1, a2   time.Time
		b1, b2   time.Time
		esperado bool
	}{
		{"periodos identicos", data(1, 0), data(4, 0), data(1, 0), data(4, 0), true},
		{"reserva contida na outra", data(1, 0), data(10, 0), data(3, 0), data(5, 0), true},
		{"sobreposicao parcial", data(1, 0), data(4, 0), data(3, 0), data(6, 0), true},
		{"limites encostados liberam o veiculo", data(1, 0), data(4, 0), data(4, 0), data(6, 0), false},
		{"periodos distantes", data(1, 0), data(2, 0), data(8, 0), data(9, 0), false},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			obtido := PeriodosSobrepostos(caso.a1, caso.a2, caso.b1, caso.b2)
			if obtido != caso.esperado {
				t.Errorf("PeriodosSobrepostos() = %v, esperado %v", obtido, caso.esperado)
			}
		})
	}
}
