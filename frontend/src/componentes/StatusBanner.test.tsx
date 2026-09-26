import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import StatusBanner from './StatusBanner';

describe('StatusBanner', () => {
  it('renderiza estado de carregamento com aria-live adequado', () => {
    render(<StatusBanner tipo="loading" mensagem="Carregando empresas..." />);

    expect(screen.getByRole('status')).toHaveTextContent('Carregando empresas...');
  });

  it('renderiza estado de sucesso com mensagem clara', () => {
    render(<StatusBanner tipo="success" mensagem="Empresa cadastrada com sucesso." />);

    expect(screen.getByRole('status')).toHaveTextContent('Empresa cadastrada com sucesso.');
  });

  it('renderiza erro como alerta', () => {
    render(<StatusBanner tipo="error" mensagem="Falha ao salvar empresa." />);

    expect(screen.getByRole('alert')).toHaveTextContent('Falha ao salvar empresa.');
  });
});
