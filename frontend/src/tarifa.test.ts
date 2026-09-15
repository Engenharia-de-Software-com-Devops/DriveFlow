import { describe, expect, it } from 'vitest';
import { diariasCobradas, valorPrevisto } from './tarifa';

describe('diariasCobradas', () => {
  it('cobra tres diarias para periodo de tres dias exatos', () => {
    const inicio = new Date('2026-03-10T10:00:00Z');
    const fim = new Date('2026-03-13T10:00:00Z');
    expect(diariasCobradas(inicio, fim)).toBe(3);
  });

  it('arredonda fracao de dia para cima', () => {
    const inicio = new Date('2026-03-10T10:00:00Z');
    const fim = new Date('2026-03-11T11:00:00Z');
    expect(diariasCobradas(inicio, fim)).toBe(2);
  });

  it('cobra no minimo uma diaria quando fim nao e posterior ao inicio', () => {
    const inicio = new Date('2026-03-10T10:00:00Z');
    const fim = new Date('2026-03-10T10:00:00Z');
    expect(diariasCobradas(inicio, fim)).toBe(1);
  });
});

describe('valorPrevisto', () => {
  it('calcula valor em centavos com tarifa diaria', () => {
    const inicio = new Date('2026-03-10T10:00:00Z');
    const fim = new Date('2026-03-13T10:00:00Z');
    expect(valorPrevisto(15000, inicio, fim)).toBe(45000);
  });
});
