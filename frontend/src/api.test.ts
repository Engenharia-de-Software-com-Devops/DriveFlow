import { afterEach, describe, expect, it, vi } from 'vitest';
import { criarLocacao, ErroApi, listarEmpresas } from './api';

describe('criarLocacao', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('envia veiculo, cliente e datas para a rota de locacoes', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      text: async () =>
        JSON.stringify({
          id: 'loc-1',
          empresa_id: 'emp-1',
          veiculo_id: 'vei-1',
          cliente: 'Cliente A',
          inicio: '2026-03-10T10:00:00.000Z',
          fim_previsto: '2026-03-13T10:00:00.000Z',
          valor_previsto: 45000,
          status: 'aberta',
        }),
    } as Response);

    await criarLocacao('emp-1', {
      veiculo_id: 'vei-1',
      cliente: 'Cliente A',
      inicio: '2026-03-10T10:00:00.000Z',
      fim_previsto: '2026-03-13T10:00:00.000Z',
    });

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/api/empresas/emp-1/locacoes'),
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          veiculo_id: 'vei-1',
          cliente: 'Cliente A',
          inicio: '2026-03-10T10:00:00.000Z',
          fim_previsto: '2026-03-13T10:00:00.000Z',
        }),
      }),
    );
  });

  it('propaga ErroApi em conflito de reserva', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: false,
      status: 409,
      text: async () => JSON.stringify({ erro: 'conflito com a locacao loc-2' }),
    } as Response);

    await expect(
      criarLocacao('emp-1', {
        veiculo_id: 'vei-1',
        cliente: 'Cliente A',
        inicio: '2026-03-10T10:00:00.000Z',
        fim_previsto: '2026-03-13T10:00:00.000Z',
      }),
    ).rejects.toEqual(new ErroApi('conflito com a locacao loc-2', 409));
  });
});

describe('listarEmpresas', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('chama GET /api/empresas', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
      ok: true,
      text: async () =>
        JSON.stringify([
          { id: 'emp-1', nome: 'Locadora Alfa', cnpj: '12345678000190' },
        ]),
    } as Response);

    const empresas = await listarEmpresas();

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/api/empresas'),
      expect.objectContaining({
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    expect(empresas).toEqual([
      { id: 'emp-1', nome: 'Locadora Alfa', cnpj: '12345678000190' },
    ]);
  });
});
