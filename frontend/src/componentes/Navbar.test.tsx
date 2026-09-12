import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import Navbar from './Navbar';
import type { Empresa } from '../api';

const empresa: Empresa = {
  id: 'emp-1',
  nome: 'Locadora Alfa',
  cnpj: '12345678000190',
};

describe('Navbar', () => {
  it('exibe apenas a marca quando nao ha empresa', () => {
    render(
      <MemoryRouter>
        <Navbar empresa={null} onSair={() => {}} />
      </MemoryRouter>,
    );

    expect(screen.getByText('DriveFlow')).toBeInTheDocument();
    expect(screen.queryByRole('link')).not.toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('exibe links de navegacao e dropdown com CNPJ', async () => {
    const usuario = userEvent.setup();
    const onSair = vi.fn();

    render(
      <MemoryRouter initialEntries={['/frota']}>
        <Navbar empresa={empresa} onSair={onSair} />
      </MemoryRouter>,
    );

    expect(screen.getByRole('link', { name: 'Frota' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Reservas' })).toBeInTheDocument();

    await usuario.click(screen.getByRole('button', { name: /Locadora Alfa/i }));

    expect(screen.getByText('CNPJ 12.345.678/0001-90')).toBeInTheDocument();

    await usuario.click(screen.getByRole('button', { name: 'Sair' }));
    expect(onSair).toHaveBeenCalledOnce();
  });
});
