import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as api from '../api';
import ListaEmpresas from './ListaEmpresas';

describe('ListaEmpresas', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('mostra estado vazio quando nao ha empresas', async () => {
    vi.spyOn(api, 'listarEmpresas').mockResolvedValue([]);

    render(<ListaEmpresas aoSelecionar={() => {}} />);

    expect(await screen.findByText('Nenhuma empresa cadastrada')).toBeInTheDocument();
  });

  it('chama aoSelecionar ao clicar numa empresa', async () => {
    const usuario = userEvent.setup();
    const aoSelecionar = vi.fn();
    const empresa = { id: 'emp-1', nome: 'Locadora Alfa', cnpj: '12345678000190' };
    vi.spyOn(api, 'listarEmpresas').mockResolvedValue([empresa]);

    render(<ListaEmpresas aoSelecionar={aoSelecionar} />);

    await usuario.click(await screen.findByRole('button', { name: /Locadora Alfa/i }));
    expect(aoSelecionar).toHaveBeenCalledWith(empresa);
  });
});
