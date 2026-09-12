// Cliente HTTP da API de locacao.
//
// A URL base vem de REACT_APP_API_URL, fixada no momento do build.
//
// No docker compose ela e definida como string vazia: o nginx do container web
// serve a aplicacao e repassa /api e /health para o container api, entao as
// chamadas ficam relativas e nao precisam de CORS. Fora do compose, o padrao e
// a api rodando na maquina local.
const URL_BASE = (process.env.REACT_APP_API_URL ?? 'http://localhost:8080').replace(/\/$/, '');

// ErroApi carrega a mensagem devolvida pela api junto com o status HTTP,
// para a interface distinguir conflito de reserva (409) de dado invalido (422).
export class ErroApi extends Error {
  constructor(mensagem, status) {
    super(mensagem);
    this.name = 'ErroApi';
    this.status = status;
  }
}

async function requisitar(caminho, opcoes = {}) {
  let resposta;
  try {
    resposta = await fetch(`${URL_BASE}${caminho}`, {
      headers: { 'Content-Type': 'application/json' },
      ...opcoes,
    });
  } catch (causa) {
    throw new ErroApi('Nao foi possivel falar com a API. Ela esta no ar?', 0);
  }

  const texto = await resposta.text();
  const corpo = texto ? JSON.parse(texto) : null;

  if (!resposta.ok) {
    const mensagem = (corpo && corpo.erro) || `Falha na requisicao (HTTP ${resposta.status})`;
    throw new ErroApi(mensagem, resposta.status);
  }
  return corpo;
}

export function verificarSaude() {
  return requisitar('/health');
}

export function criarEmpresa(nome, cnpj) {
  return requisitar('/api/empresas', {
    method: 'POST',
    body: JSON.stringify({ nome, cnpj }),
  });
}

export function listarFrota(empresaId) {
  return requisitar(`/api/empresas/${empresaId}/veiculos`);
}

export function cadastrarVeiculo(empresaId, veiculo) {
  return requisitar(`/api/empresas/${empresaId}/veiculos`, {
    method: 'POST',
    body: JSON.stringify(veiculo),
  });
}

export function listarLocacoes(empresaId) {
  return requisitar(`/api/empresas/${empresaId}/locacoes`);
}

export function reservar(empresaId, reserva) {
  return requisitar(`/api/empresas/${empresaId}/locacoes`, {
    method: 'POST',
    body: JSON.stringify(reserva),
  });
}

export function devolver(empresaId, locacaoId, devolvidoEm) {
  return requisitar(`/api/empresas/${empresaId}/locacoes/${locacaoId}/devolucao`, {
    method: 'POST',
    body: JSON.stringify({ devolvido_em: devolvidoEm }),
  });
}

// Os valores trafegam em centavos para evitar erro de arredondamento.
export function formatarMoeda(centavos) {
  if (centavos === null || centavos === undefined) return '-';
  return (centavos / 100).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
}

export function formatarData(iso) {
  if (!iso) return '-';
  return new Date(iso).toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' });
}
