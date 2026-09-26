import { describe, expect, it } from 'vitest';
import type { Veiculo } from '../api';
import {
  MERCOSUL_PLATE_REGEX,
  randomCompanyFields,
  randomMercosulPlate,
  randomRentalFields,
  randomVehicleFields,
} from './preenchimentoAleatorio';

describe('preenchimentoAleatorio', () => {
  it('randomMercosulPlate matches backend plate format', () => {
    for (let i = 0; i < 50; i += 1) {
      expect(randomMercosulPlate()).toMatch(MERCOSUL_PLATE_REGEX);
    }
  });

  it('randomCompanyFields returns 14-digit CNPJ and non-empty name', () => {
    const fields = randomCompanyFields();
    expect(fields.nome.trim().length).toBeGreaterThan(0);
    expect(fields.cnpj).toMatch(/^\d{14}$/);
  });

  it('randomVehicleFields returns valid tariff and category', () => {
    const fields = randomVehicleFields();
    expect(fields.placa).toMatch(MERCOSUL_PLATE_REGEX);
    expect(fields.modelo.trim().length).toBeGreaterThan(0);
    expect(['economico', 'intermediario', 'executivo', 'utilitario']).toContain(fields.categoria);
    const tarifa = Number(fields.tarifa);
    expect(Number.isFinite(tarifa)).toBe(true);
    expect(tarifa).toBeGreaterThanOrEqual(0.01);
  });

  it('randomRentalFields returns empty ids when no vehicles', () => {
    expect(randomRentalFields([])).toEqual({
      veiculoId: '',
      cliente: '',
      inicio: '',
      fim: '',
    });
  });

  it('randomRentalFields returns fim after inicio', () => {
    const veiculos: Veiculo[] = [
      {
        id: 'vei-1',
        placa: 'ABC1D23',
        modelo: 'Onix',
        categoria: 'economico',
        tarifa_diaria: 10000,
        status: 'disponivel',
      },
    ];

    for (let i = 0; i < 20; i += 1) {
      const fields = randomRentalFields(veiculos);
      expect(fields.veiculoId).toBe('vei-1');
      expect(fields.cliente.trim().length).toBeGreaterThan(0);

      const inicio = new Date(fields.inicio);
      const fim = new Date(fields.fim);
      expect(Number.isNaN(inicio.getTime())).toBe(false);
      expect(Number.isNaN(fim.getTime())).toBe(false);
      expect(fim.getTime()).toBeGreaterThan(inicio.getTime());
    }
  });
});
