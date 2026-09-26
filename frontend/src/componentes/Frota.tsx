import { useCallback, useEffect, useState, type FormEvent } from 'react';
import { cadastrarVeiculo, formatarMoeda, ErroApi, type Veiculo } from '../api';
import Modal from './Modal';
import type { ControleAcaoPagina } from './Page';
import StatusBanner from './StatusBanner';

interface Props {
  empresaId: string;
  veiculos: Veiculo[];
  aoAtualizar: () => Promise<void>;
  onControleAcao?: (controle: ControleAcaoPagina) => void;
}

interface Form {
  placa: string;
  modelo: string;
  categoria: string;
  tarifa: string;
}

const VEICULO_VAZIO: Form = { placa: '', modelo: '', categoria: 'economico', tarifa: '' };

export default function Frota({ empresaId, veiculos, aoAtualizar, onControleAcao }: Props) {
  const [modalAberto, setModalAberto] = useState(false);
  const [form, setForm] = useState<Form>(VEICULO_VAZIO);
  const [erro, setErro] = useState('');
  const [enviando, setEnviando] = useState(false);

  const abrirModal = useCallback(() => {
    setErro('');
    setForm(VEICULO_VAZIO);
    setModalAberto(true);
  }, []);

  useEffect(() => {
    onControleAcao?.({ abrirModal });
  }, [onControleAcao, abrirModal]);

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
    setEnviando(true);
    try {
      await cadastrarVeiculo(empresaId, {
        placa: form.placa,
        modelo: form.modelo,
        categoria: form.categoria,
        tarifa_diaria: Math.round(Number(form.tarifa) * 100),
      });
      setModalAberto(false);
      setForm(VEICULO_VAZIO);
      await aoAtualizar();
    } catch (falha) {
      setErro(falha instanceof ErroApi ? falha.message : 'Falha inesperada.');
    } finally {
      setEnviando(false);
    }
  }

  return (
    <section className="cartao">
      <div className="tabela-scroll">
      <table>
        <caption className="sr-only">Veiculos cadastrados</caption>
        <thead>
          <tr>
            <th scope="col">Placa</th>
            <th scope="col">Modelo</th>
            <th scope="col">Categoria</th>
            <th scope="col">Tarifa diaria</th>
            <th scope="col">Status</th>
          </tr>
        </thead>
        <tbody>
          {veiculos.length === 0 ? (
            <tr>
              <td colSpan={5} className="ajuda">
                Nenhum veiculo cadastrado ainda.
              </td>
            </tr>
          ) : (
            veiculos.map((veiculo) => (
              <tr key={veiculo.id}>
                <td>{veiculo.placa}</td>
                <td>{veiculo.modelo}</td>
                <td>{veiculo.categoria}</td>
                <td>{formatarMoeda(veiculo.tarifa_diaria)}</td>
                <td>{veiculo.status}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
      </div>

      <Modal aberto={modalAberto} titulo="Adicionar veiculo" onFechar={fecharModal}>
        <form onSubmit={enviar} className="modal-form">
          <label htmlFor="veiculo-placa">Placa</label>
          <input
            id="veiculo-placa"
            value={form.placa}
            onChange={(e) => alterar('placa', e.target.value)}
            placeholder="ABC1D23"
            required
          />

          <label htmlFor="veiculo-modelo">Modelo</label>
          <input
            id="veiculo-modelo"
            value={form.modelo}
            onChange={(e) => alterar('modelo', e.target.value)}
            placeholder="Onix 1.0"
            required
          />

          <label htmlFor="veiculo-categoria">Categoria</label>
          <select
            id="veiculo-categoria"
            value={form.categoria}
            onChange={(e) => alterar('categoria', e.target.value)}
          >
            <option value="economico">Economico</option>
            <option value="intermediario">Intermediario</option>
            <option value="executivo">Executivo</option>
            <option value="utilitario">Utilitario</option>
          </select>

          <label htmlFor="veiculo-tarifa">Tarifa diaria (R$)</label>
          <input
            id="veiculo-tarifa"
            type="number"
            min="0.01"
            step="0.01"
            value={form.tarifa}
            onChange={(e) => alterar('tarifa', e.target.value)}
            placeholder="150.00"
            required
          />

          {erro && <StatusBanner tipo="error" mensagem={erro} />}

          <div className="modal-acoes">
            <button type="button" className="botao-secundario" onClick={fecharModal} disabled={enviando}>
              Cancelar
            </button>
            <button type="submit" disabled={enviando}>
              {enviando ? 'Salvando...' : 'Adicionar veiculo'}
            </button>
          </div>
        </form>
      </Modal>
    </section>
  );
}
