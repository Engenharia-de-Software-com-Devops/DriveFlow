import { useState } from 'react';
import { devolver, formatarData, formatarMoeda, reservar } from '../api';

const RESERVA_VAZIA = { veiculoId: '', cliente: '', inicio: '', fim: '' };

// Locacoes cria reservas e encerra contratos.
//
// O conflito de reserva apontado no diagnostico do E1 chega aqui como um 409
// da api e e mostrado ao atendente antes de virar um incidente com o cliente.
export default function Locacoes({ empresaId, veiculos, locacoes, aoAtualizar }) {
  const [form, setForm] = useState(RESERVA_VAZIA);
  const [erro, setErro] = useState('');
  const [aviso, setAviso] = useState('');
  const [enviando, setEnviando] = useState(false);

  function alterar(campo, valor) {
    setForm((atual) => ({ ...atual, [campo]: valor }));
  }

  async function enviarReserva(evento) {
    evento.preventDefault();
    setErro('');
    setAviso('');
    setEnviando(true);
    try {
      await reservar(empresaId, {
        veiculo_id: form.veiculoId,
        cliente: form.cliente,
        inicio: new Date(form.inicio).toISOString(),
        fim_previsto: new Date(form.fim).toISOString(),
      });
      setForm(RESERVA_VAZIA);
      setAviso('Reserva criada.');
      await aoAtualizar();
    } catch (falha) {
      setErro(
        falha.status === 409
          ? `Reserva recusada: ${falha.message}`
          : falha.message
      );
    } finally {
      setEnviando(false);
    }
  }

  async function encerrar(locacaoId) {
    setErro('');
    setAviso('');
    try {
      const contrato = await devolver(empresaId, locacaoId, new Date().toISOString());
      setAviso(`Devolucao registrada. Valor final: ${formatarMoeda(contrato.valor_final)}.`);
      await aoAtualizar();
    } catch (falha) {
      setErro(falha.message);
    }
  }

  const placaPorId = Object.fromEntries(veiculos.map((v) => [v.id, v.placa]));

  return (
    <section className="cartao">
      <h2>Locacoes</h2>

      {locacoes.length === 0 ? (
        <p className="ajuda">Nenhuma locacao registrada ainda.</p>
      ) : (
        <table>
          <caption className="sr-only">Contratos de locacao</caption>
          <thead>
            <tr>
              <th scope="col">Veiculo</th>
              <th scope="col">Cliente</th>
              <th scope="col">Inicio</th>
              <th scope="col">Fim previsto</th>
              <th scope="col">Valor</th>
              <th scope="col">Status</th>
              <th scope="col">Acao</th>
            </tr>
          </thead>
          <tbody>
            {locacoes.map((contrato) => (
              <tr key={contrato.id}>
                <td>{placaPorId[contrato.veiculo_id] || contrato.veiculo_id}</td>
                <td>{contrato.cliente}</td>
                <td>{formatarData(contrato.inicio)}</td>
                <td>{formatarData(contrato.fim_previsto)}</td>
                <td>
                  {contrato.status === 'encerrada'
                    ? formatarMoeda(contrato.valor_final)
                    : `${formatarMoeda(contrato.valor_previsto)} (previsto)`}
                </td>
                <td>{contrato.status}</td>
                <td>
                  {contrato.status === 'aberta' && (
                    <button type="button" onClick={() => encerrar(contrato.id)}>
                      Registrar devolucao
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <form onSubmit={enviarReserva}>
        <h3>Nova reserva</h3>

        <label htmlFor="reserva-veiculo">Veiculo</label>
        <select
          id="reserva-veiculo"
          value={form.veiculoId}
          onChange={(e) => alterar('veiculoId', e.target.value)}
          required
        >
          <option value="">Selecione</option>
          {veiculos.map((veiculo) => (
            <option key={veiculo.id} value={veiculo.id}>
              {veiculo.placa} - {veiculo.modelo}
            </option>
          ))}
        </select>

        <label htmlFor="reserva-cliente">Cliente</label>
        <input
          id="reserva-cliente"
          value={form.cliente}
          onChange={(e) => alterar('cliente', e.target.value)}
          placeholder="Nome do cliente"
          required
        />

        <label htmlFor="reserva-inicio">Inicio</label>
        <input
          id="reserva-inicio"
          type="datetime-local"
          value={form.inicio}
          onChange={(e) => alterar('inicio', e.target.value)}
          required
        />

        <label htmlFor="reserva-fim">Fim previsto</label>
        <input
          id="reserva-fim"
          type="datetime-local"
          value={form.fim}
          onChange={(e) => alterar('fim', e.target.value)}
          required
        />

        <button type="submit" disabled={enviando}>
          {enviando ? 'Reservando...' : 'Reservar'}
        </button>
      </form>

      {erro && <p role="alert" className="erro">{erro}</p>}
      {aviso && <p role="status" className="aviso">{aviso}</p>}
    </section>
  );
}
