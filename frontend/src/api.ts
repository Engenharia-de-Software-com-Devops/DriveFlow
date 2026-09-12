// Cliente HTTP da API de locacao.
//
// A URL base vem de VITE_API_URL, fixada no momento do build.
//
// No docker compose ela fica vazia: o nginx do container web serve a
// aplicacao e repassa /api e /health para o container api, entao as chamadas
// ficam relativas e nao precisam de CORS. Fora do compose, o padrao e a api
// rodando na maquina local.
const URL_BASE = (import.meta.env.VITE_API_URL ?? 'http://localhost:8080').replace(/\/$/, '');

// ErroApi carrega a mensagem devolvida pela api junto com o status HTTP,
// para a interface distinguir conflito de reserva (409) de dado invalido (422).
export class ErroApi extends Error {
  status: number;

  constructor(mensagem: string, status: number) {
    super(mensagem);
    this.name = 'ErroApi';
    this.status = status;
  }
}

async function requisitar<T>(caminho: string, opcoes: RequestInit = {}): Promise<T> {
  let resposta: Response;
  try {
    resposta = await fetch(`${URL_BASE}${caminho}`, {
      headers: { 'Content-Type': 'application/json' },
      ...opcoes,
    });
  } catch {
    throw new ErroApi('Nao foi possivel falar com a API. Ela esta no ar?', 0);
  }

  const texto = await resposta.text();
  const corpo = texto ? JSON.parse(texto) : null;

  if (!resposta.ok) {
    const mensagem = corpo?.erro || `Falha na requisicao (HTTP ${resposta.status})`;
    throw new ErroApi(mensagem, resposta.status);
  }
  return corpo;
}

export interface Empresa {
  id: string;
  nome: string;
  cnpj: string;
}

export interface Veiculo {
  id: string;
  placa: string;
  modelo: string;
  categoria: string;
  tarifa_diaria: number;
  status: string;
}

export function verificarSaude() {
  return requisitar('/health');
}

export function criarEmpresa(nome: string, cnpj: string) {
  return requisitar<Empresa>('/api/empresas', {
    method: 'POST',
    body: JSON.stringify({ nome, cnpj }),
  });
}

export function listarFrota(empresaId: string) {
  return requisitar<Veiculo[]>(`/api/empresas/${empresaId}/veiculos`);
}

export function cadastrarVeiculo(
  empresaId: string,
  veiculo: { placa: string; modelo: string; categoria: string; tarifa_diaria: number },
) {
  return requisitar<Veiculo>(`/api/empresas/${empresaId}/veiculos`, {
    method: 'POST',
    body: JSON.stringify(veiculo),
  });
}

// Os valores trafegam em centavos para evitar erro de arredondamento.
export function formatarMoeda(centavos: number | null | undefined) {
  if (centavos === null || centavos === undefined) return '-';
  return (centavos / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}
