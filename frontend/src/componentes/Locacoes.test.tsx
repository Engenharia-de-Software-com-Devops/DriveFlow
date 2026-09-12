import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState, type ComponentProps } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Locacoes from './Locacoes';
import Page, { type ControleAcaoPagina } from './Page';
import * as api from '../api';
import type { Locacao, Veiculo } from '../api';

const veiculos: Veiculo[] = [
  {
    id: 'vei-1',
    placa: 'ABC1D23',
    modelo: 'Onix 1.0',
    categoria: 'economico',
    tarifa_diaria: 15000,
    status: 'disponivel',
  },
];

const locacoes: Locacao[] = [];

function LocacoesComPagina(props: ComponentProps<typeof Locacoes>) {
  const [controle, setControle] = useState<ControleAcaoPagina | null>(null);

  return (
    <Page
      titulo="Reservas"
      descricao="Teste"
      rightContent={
        <button
          type="button"
          onClick={() => controle?.abrirModal()}
          disabled={!controle || controle.acaoDesabilitada}
        >
          Criar reserva
        </button>
      }
    >
      <Locacoes {...props} onControleAcao={setControle} />
    </Page>
  );
}

async function abrirModalReserva(usuario: ReturnType<typeof userEvent.setup>) {
  await usuario.click(screen.getByRole('button', { name: 'Criar reserva' }));
  return within(await screen.findByRole('dialog'));
}

describe('Locacoes', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('atualiza o valor estimado ao selecionar veiculo e datas', async () => {
    const usuario = userEvent.setup();
    render(
      <LocacoesComPagina
        empresaId="emp-1"
        veiculos={veiculos}
        locacoes={locacoes}
        aoAtualizar={async () => {}}
      />,
    );

    const modal = await abrirModalReserva(usuario);
    await usuario.selectOptions(modal.getByLabelText('Veiculo'), 'vei-1');
    await usuario.type(modal.getByLabelText('Inicio'), '2026-03-10T10:00');
    await usuario.type(modal.getByLabelText('Fim previsto'), '2026-03-13T10:00');

    expect(modal.getByText(/Valor estimado:/)).toHaveTextContent('R$');
    expect(modal.getByText(/Valor estimado:/)).toHaveTextContent('450,00');
  });

  it('cria reserva e recarrega a lista', async () => {
    const usuario = userEvent.setup();
    const criar = vi.spyOn(api, 'criarLocacao').mockResolvedValue({
      id: 'loc-1',
      empresa_id: 'emp-1',
      veiculo_id: 'vei-1',
      cliente: 'Cliente A',
      inicio: '2026-03-10T13:00:00.000Z',
      fim_previsto: '2026-03-13T13:00:00.000Z',
      valor_previsto: 45000,
      status: 'aberta',
    });
    const aoAtualizar = vi.fn().mockResolvedValue(undefined);

    render(
      <LocacoesComPagina
        empresaId="emp-1"
        veiculos={veiculos}
        locacoes={locacoes}
        aoAtualizar={aoAtualizar}
      />,
    );

    const modal = await abrirModalReserva(usuario);
    await usuario.selectOptions(modal.getByLabelText('Veiculo'), 'vei-1');
    await usuario.type(modal.getByLabelText('Cliente'), 'Cliente A');
    await usuario.type(modal.getByLabelText('Inicio'), '2026-03-10T10:00');
    await usuario.type(modal.getByLabelText('Fim previsto'), '2026-03-13T10:00');
    await usuario.click(modal.getByRole('button', { name: 'Criar reserva' }));

    await waitFor(() => {
      expect(criar).toHaveBeenCalledWith('emp-1', {
        veiculo_id: 'vei-1',
        cliente: 'Cliente A',
        inicio: expect.any(String),
        fim_previsto: expect.any(String),
      });
    });
    expect(aoAtualizar).toHaveBeenCalled();
  });

  it('exibe mensagem de conflito quando a api retorna 409', async () => {
    const usuario = userEvent.setup();
    vi.spyOn(api, 'criarLocacao').mockRejectedValue(
      new api.ErroApi('conflito com a locacao loc-2', 409),
    );

    render(
      <LocacoesComPagina
        empresaId="emp-1"
        veiculos={veiculos}
        locacoes={locacoes}
        aoAtualizar={async () => {}}
      />,
    );

    const modal = await abrirModalReserva(usuario);
    await usuario.selectOptions(modal.getByLabelText('Veiculo'), 'vei-1');
    await usuario.type(modal.getByLabelText('Cliente'), 'Cliente A');
    await usuario.type(modal.getByLabelText('Inicio'), '2026-03-10T10:00');
    await usuario.type(modal.getByLabelText('Fim previsto'), '2026-03-13T10:00');
    await usuario.click(modal.getByRole('button', { name: 'Criar reserva' }));

    expect(await modal.findByRole('alert')).toHaveTextContent('conflito com a locacao loc-2');
  });
});
