import { useCallback, useEffect, useMemo, useState, type FormEvent } from 'react';
import {
  criarLocacao,
  devolverLocacao,
  formatarMoeda,
  ErroApi,
  type Locacao,
  type Veiculo,
} from '../api';
import { valorPrevisto } from '../tarifa';
import Modal from './Modal';
import type { ControleAcaoPagina } from './Page';

interface Props {
  empresaId: string;
  veiculos: Veiculo[];
  locacoes: Locacao[];
  aoAtualizar: () => Promise<void>;
  onControleAcao?: (controle: ControleAcaoPagina) => void;
}

interface Form {
  veiculoId: string;
  cliente: string;
  inicio: string;
  fim: string;
}

const FORM_VAZIO: Form = { veiculoId: '', cliente: '', inicio: '', fim: '' };

function formatarData(iso: string) {
  return new Date(iso).toLocaleString('pt-BR', { dateStyle: 'short', timeStyle: 'short' });
}

function placaDoVeiculo(veiculos: Veiculo[], veiculoId: string) {
  return veiculos.find((veiculo) => veiculo.id === veiculoId)?.placa ?? veiculoId;
}

export default function Locacoes({ empresaId, veiculos, locacoes, aoAtualizar, onControleAcao }: Props) {
  const [modalAberto, setModalAberto] = useState(false);
  const [form, setForm] = useState<Form>(FORM_VAZIO);
  const [erro, setErro] = useState('');
  const [enviando, setEnviando] = useState(false);
  const [devolvendo, setDevolvendo] = useState<string | null>(null);

  const disponiveis = useMemo(
    () => veiculos.filter((veiculo) => veiculo.status === 'disponivel'),
    [veiculos],
  );

  const veiculoSelecionado = useMemo(
    () => veiculos.find((veiculo) => veiculo.id === form.veiculoId),
    [veiculos, form.veiculoId],
  );

  const valorEstimado = useMemo(() => {
    if (!veiculoSelecionado || !form.inicio || !form.fim) return null;

    const inicio = new Date(form.inicio);
    const fim = new Date(form.fim);
    if (Number.isNaN(inicio.getTime()) || Number.isNaN(fim.getTime()) || !(fim > inicio)) {
      return null;
    }

    return valorPrevisto(veiculoSelecionado.tarifa_diaria, inicio, fim);
  }, [veiculoSelecionado, form.inicio, form.fim]);

  const abrirModal = useCallback(() => {
    setErro('');
    setForm(FORM_VAZIO);
    setModalAberto(true);
  }, []);

  useEffect(() => {
    onControleAcao?.({
      abrirModal,
      acaoDesabilitada: disponiveis.length === 0,
    });
  }, [onControleAcao, abrirModal, disponiveis.length]);

  function fecharModal() {
    if (enviando) return;
    setModalAberto(false);
    setErro('');
  }

  function alterar(campo: keyof Form, valor: string) {
    setForm((atual) => ({ ...atual, [campo]: valor }));
  }

  async function enviar(evento: FormEvent) {
    evento.preventDefault();
    setErro('');

    const cliente = form.cliente.trim();
    if (!form.veiculoId) {
      setErro('Selecione um veiculo.');
      return;
    }
    if (!cliente) {
      setErro('Informe o cliente.');
      return;
    }

    const inicio = new Date(form.inicio);
    const fim = new Date(form.fim);
    if (!(fim > inicio)) {
      setErro('A data de fim deve ser posterior ao inicio.');
      return;
    }

    setEnviando(true);
    try {
      await criarLocacao(empresaId, {
        veiculo_id: form.veiculoId,
        cliente,
        inicio: inicio.toISOString(),
        fim_previsto: fim.toISOString(),
      });
      setModalAberto(false);
      setForm(FORM_VAZIO);
      await aoAtualizar();
    } catch (falha) {
      setErro(falha instanceof ErroApi ? falha.message : 'Falha inesperada.');
    } finally {
      setEnviando(false);
    }
  }

  async function devolver(locacaoId: string) {
    setErro('');
    setDevolvendo(locacaoId);
    try {
      await devolverLocacao(empresaId, locacaoId);
      await aoAtualizar();
    } catch (falha) {
      setErro(falha instanceof ErroApi ? falha.message : 'Falha inesperada.');
    } finally {
      setDevolvendo(null);
    }
  }

  return (
    <section className="cartao">
      <div className="tabela-scroll">
      <table>
        <caption className="sr-only">Reservas cadastradas</caption>
        <thead>
          <tr>
            <th scope="col">Cliente</th>
            <th scope="col">Veiculo</th>
            <th scope="col">Inicio</th>
            <th scope="col">Fim</th>
            <th scope="col">Valor</th>
            <th scope="col">Status</th>
            <th scope="col">Acoes</th>
          </tr>
        </thead>
        <tbody>
          {locacoes.length === 0 ? (
            <tr>
              <td colSpan={7} className="ajuda">
                Nenhuma reserva cadastrada ainda.
              </td>
            </tr>
          ) : (
            locacoes.map((locacao) => (
              <tr key={locacao.id}>
                <td>{locacao.cliente}</td>
                <td>{placaDoVeiculo(veiculos, locacao.veiculo_id)}</td>
                <td>{formatarData(locacao.inicio)}</td>
                <td>{formatarData(locacao.fim_previsto)}</td>
                <td>
                  {locacao.status === 'encerrada' && locacao.valor_final != null
                    ? formatarMoeda(locacao.valor_final)
                    : formatarMoeda(locacao.valor_previsto)}
                </td>
                <td>{locacao.status}</td>
                <td>
                  {locacao.status === 'aberta' && (
                    <button
                      type="button"
                      className="botao-secundario"
                      disabled={devolvendo === locacao.id}
                      onClick={() => devolver(locacao.id)}
                    >
                      {devolvendo === locacao.id ? 'Devolvendo...' : 'Devolver'}
                    </button>
                  )}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
      </div>

      {disponiveis.length === 0 && (
        <p className="ajuda">Cadastre um veiculo disponivel para criar reservas.</p>
      )}

      {erro && !modalAberto && <p role="alert" className="erro">{erro}</p>}

      <Modal aberto={modalAberto} titulo="Nova reserva" onFechar={fecharModal}>
        <form onSubmit={enviar} className="modal-form">
          <label htmlFor="locacao-veiculo">Veiculo</label>
          <select
            id="locacao-veiculo"
            value={form.veiculoId}
            onChange={(e) => alterar('veiculoId', e.target.value)}
            required
          >
            <option value="">Selecione...</option>
            {disponiveis.map((veiculo) => (
              <option key={veiculo.id} value={veiculo.id}>
                {veiculo.placa} — {veiculo.modelo}
              </option>
            ))}
          </select>

          <label htmlFor="locacao-cliente">Cliente</label>
          <input
            id="locacao-cliente"
            value={form.cliente}
            onChange={(e) => alterar('cliente', e.target.value)}
            placeholder="Nome do cliente"
            required
          />

          <label htmlFor="locacao-inicio">Inicio</label>
          <input
            id="locacao-inicio"
            type="datetime-local"
            value={form.inicio}
            onChange={(e) => alterar('inicio', e.target.value)}
            required
          />

          <label htmlFor="locacao-fim">Fim previsto</label>
          <input
            id="locacao-fim"
            type="datetime-local"
            value={form.fim}
            onChange={(e) => alterar('fim', e.target.value)}
            required
          />

          <p className="ajuda">
            Valor estimado: {valorEstimado != null ? formatarMoeda(valorEstimado) : '-'}
          </p>

          {erro && <p role="alert" className="erro">{erro}</p>}

          <div className="modal-acoes">
            <button type="button" className="botao-secundario" onClick={fecharModal} disabled={enviando}>
              Cancelar
            </button>
            <button type="submit" disabled={enviando}>
              {enviando ? 'Salvando...' : 'Criar reserva'}
            </button>
          </div>
        </form>
      </Modal>
    </section>
  );
}
