package locacao

import "time"

// PercentualMultaAtraso e a multa aplicada sobre cada diaria devolvida em atraso.
const PercentualMultaAtraso = 30

// Diarias converte um periodo em quantidade de diarias cobradas.
// Qualquer fracao de dia gera uma diaria cheia e todo contrato cobra no minimo uma.
func Diarias(inicio, fim time.Time) int64 {
	duracao := fim.Sub(inicio)
	if duracao <= 0 {
		return 1
	}
	dias := duracao / (24 * time.Hour)
	if duracao%(24*time.Hour) != 0 {
		dias++
	}
	return int64(dias)
}

// ValorPrevisto calcula o valor do contrato no momento da reserva, em centavos.
func ValorPrevisto(tarifaDiaria int64, inicio, fim time.Time) int64 {
	return tarifaDiaria * Diarias(inicio, fim)
}

// ValorFinal calcula o valor cobrado na devolucao, em centavos.
// Diarias consumidas alem do fim previsto sao cobradas com multa de atraso;
// devolucao antecipada nao reduz o valor previsto no contrato.
func ValorFinal(tarifaDiaria int64, inicio, fimPrevisto, devolucao time.Time) int64 {
	previsto := ValorPrevisto(tarifaDiaria, inicio, fimPrevisto)
	if !devolucao.After(fimPrevisto) {
		return previsto
	}

	diariasAtraso := Diarias(fimPrevisto, devolucao)
	valorAtraso := tarifaDiaria * diariasAtraso
	multa := valorAtraso * PercentualMultaAtraso / 100
	return previsto + valorAtraso + multa
}

// PeriodosSobrepostos indica se dois intervalos [inicioA, fimA) e [inicioB, fimB)
// se cruzam. E a regra que impede a reserva dupla do mesmo veiculo.
func PeriodosSobrepostos(inicioA, fimA, inicioB, fimB time.Time) bool {
	return inicioA.Before(fimB) && inicioB.Before(fimA)
}
