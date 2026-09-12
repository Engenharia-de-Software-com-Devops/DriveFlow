import { useState } from 'react';
import { cadastrarVeiculo, formatarMoeda } from '../api';

const VEICULO_VAZIO = { placa: '', modelo: '', categoria: 'economico', tarifa: '' };

// Frota lista os veiculos da empresa e permite adicionar novos.
export default function Frota({ empresaId, veiculos, aoAtualizar }) {
  const [form, setForm] = useState(VEICULO_VAZIO);
  const [erro, setErro] = useState('');
  const [enviando, setEnviando] = useState(false);

  function alterar(campo, valor) {
    setForm((atual) => ({ ...atual, [campo]: valor }));
  }

  async function enviar(evento) {
    evento.preventDefault();
    setErro('');
    setEnviando(true);
    try {
      await cadastrarVeiculo(empresaId, {
        placa: form.placa,
        modelo: form.modelo,
        categoria: form.categoria,
        // A tarifa e digitada em reais e enviada em centavos.
        tarifa_diaria: Math.round(Number(form.tarifa) * 100),
      });
      setForm(VEICULO_VAZIO);
      await aoAtualizar();
    } catch (falha) {
      setErro(falha.message);
    } finally {
      setEnviando(false);
    }
  }

  return (
    <section className="cartao">
      <h2>Frota</h2>

      {veiculos.length === 0 ? (
        <p className="ajuda">Nenhum veiculo cadastrado ainda.</p>
      ) : (
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
            {veiculos.map((veiculo) => (
              <tr key={veiculo.id}>
                <td>{veiculo.placa}</td>
                <td>{veiculo.modelo}</td>
                <td>{veiculo.categoria}</td>
                <td>{formatarMoeda(veiculo.tarifa_diaria)}</td>
                <td>{veiculo.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <form onSubmit={enviar}>
        <h3>Adicionar veiculo</h3>

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

        <button type="submit" disabled={enviando}>
          {enviando ? 'Salvando...' : 'Adicionar veiculo'}
        </button>
      </form>
      {erro && <p role="alert" className="erro">{erro}</p>}
    </section>
  );
}
