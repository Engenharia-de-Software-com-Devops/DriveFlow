package entities

import "time"

// LateFeePercent e a multa aplicada sobre cada diaria devolvida em atraso.
const LateFeePercent = 30

// BillableDays converte um periodo em quantidade de diarias cobradas.
// Qualquer fracao de dia gera uma diaria cheia e todo contrato cobra no minimo uma.
func BillableDays(start, end time.Time) int64 {
	duration := end.Sub(start)
	if duration <= 0 {
		return 1
	}
	days := duration / (24 * time.Hour)
	if duration%(24*time.Hour) != 0 {
		days++
	}
	return int64(days)
}

// EstimatedTotal calcula o valor do contrato no momento da reserva, em centavos.
func EstimatedTotal(dailyRate int64, start, end time.Time) int64 {
	return dailyRate * BillableDays(start, end)
}

// FinalTotal calcula o valor cobrado na devolucao, em centavos.
// Diarias consumidas alem do fim previsto sao cobradas com multa de atraso;
// devolucao antecipada nao reduz o valor previsto no contrato.
func FinalTotal(dailyRate int64, start, expectedEnd, returnedAt time.Time) int64 {
	estimated := EstimatedTotal(dailyRate, start, expectedEnd)
	if !returnedAt.After(expectedEnd) {
		return estimated
	}

	lateDays := BillableDays(expectedEnd, returnedAt)
	lateAmount := dailyRate * lateDays
	fee := lateAmount * LateFeePercent / 100
	return estimated + lateAmount + fee
}

// PeriodsOverlap indica se dois intervalos [startA, endA) e [startB, endB)
// se cruzam. E a regra que impede a reserva dupla do mesmo veiculo.
func PeriodsOverlap(startA, endA, startB, endB time.Time) bool {
	return startA.Before(endB) && startB.Before(endA)
}
