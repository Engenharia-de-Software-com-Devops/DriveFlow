import type { Veiculo } from '../api';

/** Same pattern as backend/internal/usecases/vehicle_service.go */
export const MERCOSUL_PLATE_REGEX = /^[A-Z]{3}[0-9][0-9A-Z][0-9]{2}$/;

const VEHICLE_CATEGORIES = ['economico', 'intermediario', 'executivo', 'utilitario'] as const;

const COMPANY_PREFIXES = ['Locadora', 'Rent', 'Auto', 'Fleet'] as const;

const FIRST_NAMES = ['Ana', 'Bruno', 'Carla', 'Diego', 'Elena', 'Felipe', 'Gabriela', 'Henrique'] as const;

const LAST_NAMES = ['Silva', 'Santos', 'Oliveira', 'Souza', 'Lima', 'Costa', 'Ferreira', 'Almeida'] as const;

const VEHICLE_MODELS = ['Onix 1.0', 'HB20 Comfort', 'Corolla XEi', 'Tracker LT', 'Mobi Like', 'Kwid Zen'] as const;

function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function pick<T>(items: readonly T[]): T {
  return items[randomInt(0, items.length - 1)];
}

function randomLetter(): string {
  return String.fromCharCode(randomInt(65, 90));
}

function randomDigitChar(): string {
  return String(randomInt(0, 9));
}

function randomAlphanumeric(): string {
  const chars = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  return chars[randomInt(0, chars.length - 1)];
}

function randomDigits(length: number): string {
  let out = '';
  for (let i = 0; i < length; i += 1) {
    out += randomDigitChar();
  }
  return out;
}

export function toDatetimeLocalValue(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export function randomMercosulPlate(): string {
  return (
    randomLetter() +
    randomLetter() +
    randomLetter() +
    randomDigitChar() +
    randomAlphanumeric() +
    randomDigitChar() +
    randomDigitChar()
  );
}

export function randomCompanyFields(): { nome: string; cnpj: string } {
  return {
    nome: `${pick(COMPANY_PREFIXES)} ${pick(LAST_NAMES)} ${randomInt(1, 999)}`,
    cnpj: randomDigits(14),
  };
}

export function randomVehicleFields(): {
  placa: string;
  modelo: string;
  categoria: string;
  tarifa: string;
} {
  const tarifaReais = randomInt(80, 350) + randomInt(0, 99) / 100;
  return {
    placa: randomMercosulPlate(),
    modelo: pick(VEHICLE_MODELS),
    categoria: pick(VEHICLE_CATEGORIES),
    tarifa: tarifaReais.toFixed(2),
  };
}

export function randomRentalFields(disponiveis: Veiculo[]): {
  veiculoId: string;
  cliente: string;
  inicio: string;
  fim: string;
} {
  if (disponiveis.length === 0) {
    return { veiculoId: '', cliente: '', inicio: '', fim: '' };
  }

  const inicio = new Date();
  inicio.setHours(inicio.getHours() + randomInt(1, 48));
  inicio.setMinutes(randomInt(0, 59), 0, 0);

  const fim = new Date(inicio);
  fim.setDate(fim.getDate() + randomInt(1, 5));

  return {
    veiculoId: pick(disponiveis).id,
    cliente: `${pick(FIRST_NAMES)} ${pick(LAST_NAMES)}`,
    inicio: toDatetimeLocalValue(inicio),
    fim: toDatetimeLocalValue(fim),
  };
}
