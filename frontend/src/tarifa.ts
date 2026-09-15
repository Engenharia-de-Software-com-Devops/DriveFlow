const MS_POR_DIA = 24 * 60 * 60 * 1000;

// Espelha entities.BillableDays do backend: fracao de dia cobra diaria cheia.
export function diariasCobradas(inicio: Date, fim: Date): number {
  const duracao = fim.getTime() - inicio.getTime();
  if (duracao <= 0) return 1;

  const dias = Math.floor(duracao / MS_POR_DIA);
  const resto = duracao % MS_POR_DIA;
  return resto !== 0 ? dias + 1 : dias;
}

export function valorPrevisto(tarifaDiariaCentavos: number, inicio: Date, fim: Date): number {
  return tarifaDiariaCentavos * diariasCobradas(inicio, fim);
}
