import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

const EMPRESA = { id: 'e1', nome: 'Locadora Alfa', cnpj: '12345678000190' };
const VEICULO = {
  id: 'v1',
  empresa_id: 'e1',
  placa: 'ABC1D23',
  modelo: 'Onix 1.0',
  categoria: 'economico',
  tarifa_diaria: 15000,
  status: 'disponivel',
};

// apiFalsa responde por rota, simulando o backend sem precisar dele no ar.
function apiFalsa(rotas) {
  return jest.fn(async (url, opcoes = {}) => {
    const metodo = opcoes.method || 'GET';
    const rota = Object.keys(rotas).find((chave) => {
      const [metodoRota, caminho] = chave.split(' ');
      return metodoRota === metodo && url.includes(caminho);
    });

    if (!rota) {
      throw new Error(`rota nao simulada: ${metodo} ${url}`);
    }

    const { status = 200, corpo } = rotas[rota];
    return {
      ok: status >= 200 && status < 300,
      status,
      text: async () => JSON.stringify(corpo),
    };
  });
}

beforeEach(() => {
  window.localStorage.clear();
});

afterEach(() => {
  jest.resetAllMocks();
});

it('mostra o formulario de empresa quando nenhuma esta selecionada', async () => {
  global.fetch = apiFalsa({ 'GET /health': { corpo: { status: 'ok' } } });

  render(<App />);

  expect(await screen.findByRole('heading', { name: /cadastrar empresa/i })).toBeInTheDocument();
});

it('carrega frota e locacoes da empresa ja selecionada', async () => {
  window.localStorage.setItem('driveflow:empresa', JSON.stringify(EMPRESA));
  global.fetch = apiFalsa({
    'GET /health': { corpo: { status: 'ok' } },
    'GET /veiculos': { corpo: [VEICULO] },
    'GET /locacoes': {
      corpo: [
        {
          id: 'l1',
          veiculo_id: 'v1',
          cliente: 'Cliente A',
          inicio: '2026-03-10T10:00:00Z',
          fim_previsto: '2026-03-13T10:00:00Z',
          valor_previsto: 45000,
          status: 'aberta',
        },
      ],
    },
  });

  render(<App />);

  const tabelaFrota = await screen.findByRole('table', { name: /veiculos cadastrados/i });
  expect(within(tabelaFrota).getByText('ABC1D23')).toBeInTheDocument();
  expect(within(tabelaFrota).getByText('Onix 1.0')).toBeInTheDocument();

  const tabelaLocacoes = screen.getByRole('table', { name: /contratos de locacao/i });
  expect(within(tabelaLocacoes).getByText('Cliente A')).toBeInTheDocument();
  // A locacao mostra a placa do veiculo, e nao o id interno.
  expect(within(tabelaLocacoes).getByText('ABC1D23')).toBeInTheDocument();
  expect(within(tabelaLocacoes).getByText(/450,00.*previsto/)).toBeInTheDocument();
});

// Regressao do Gargalo 2 do E1: o conflito precisa aparecer para o atendente.
it('mostra o motivo quando a api recusa uma reserva conflitante', async () => {
  window.localStorage.setItem('driveflow:empresa', JSON.stringify(EMPRESA));
  global.fetch = apiFalsa({
    'GET /health': { corpo: { status: 'ok' } },
    'GET /veiculos': { corpo: [VEICULO] },
    'GET /locacoes': { corpo: [] },
    'POST /locacoes': {
      status: 409,
      corpo: { erro: 'veiculo ja reservado no periodo informado' },
    },
  });

  render(<App />);
  const usuario = userEvent.setup();

  await screen.findByRole('table', { name: /veiculos cadastrados/i });

  await usuario.selectOptions(screen.getByLabelText(/veiculo/i), 'v1');
  await usuario.type(screen.getByLabelText(/cliente/i), 'Cliente B');
  await usuario.type(screen.getByLabelText(/^inicio$/i), '2026-03-12T10:00');
  await usuario.type(screen.getByLabelText(/fim previsto/i), '2026-03-16T10:00');
  await usuario.click(screen.getByRole('button', { name: /reservar/i }));

  await waitFor(() => {
    expect(screen.getByRole('alert')).toHaveTextContent(/ja reservado no periodo/i);
  });
});

it('avisa quando a api esta fora do ar', async () => {
  global.fetch = jest.fn().mockRejectedValue(new TypeError('Failed to fetch'));

  render(<App />);

  expect(await screen.findByRole('alert')).toHaveTextContent(/api indisponivel/i);
});
