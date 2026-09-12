import { ErroApi, cadastrarVeiculo, formatarMoeda, listarFrota, reservar } from './api';

describe('cliente da api', () => {
  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  function responder(status, corpo) {
    global.fetch.mockResolvedValue({
      ok: status >= 200 && status < 300,
      status,
      text: async () => JSON.stringify(corpo),
    });
  }

  it('devolve a frota da empresa', async () => {
    responder(200, [{ id: 'v1', placa: 'ABC1D23' }]);

    const frota = await listarFrota('e1');

    expect(frota).toHaveLength(1);
    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/empresas/e1/veiculos'),
      expect.objectContaining({ headers: { 'Content-Type': 'application/json' } })
    );
  });

  it('propaga o 409 de conflito de reserva com a mensagem da api', async () => {
    responder(409, { erro: 'veiculo ja reservado no periodo informado' });

    await expect(reservar('e1', { veiculo_id: 'v1' })).rejects.toMatchObject({
      status: 409,
      message: 'veiculo ja reservado no periodo informado',
    });
  });

  it('propaga o 422 de dados invalidos', async () => {
    responder(422, { erro: 'dados invalidos: placa fora do padrao' });

    await expect(cadastrarVeiculo('e1', { placa: 'XX' })).rejects.toBeInstanceOf(ErroApi);
  });

  it('avisa quando a api esta fora do ar', async () => {
    global.fetch.mockRejectedValue(new TypeError('Failed to fetch'));

    await expect(listarFrota('e1')).rejects.toMatchObject({ status: 0 });
  });
});

describe('formatarMoeda', () => {
  it('converte centavos em reais', () => {
    expect(formatarMoeda(45000).replace(/ /g, ' ')).toBe('R$ 450,00');
  });

  it('mostra um traco quando nao ha valor', () => {
    expect(formatarMoeda(null)).toBe('-');
  });
});
