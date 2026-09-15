import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import Page from './Page';

describe('Page', () => {
  it('renderiza titulo e descricao', () => {
    render(
      <Page titulo="Frota" descricao="Gerencie os veiculos da empresa.">
        <p>Conteudo da pagina</p>
      </Page>,
    );

    expect(screen.getByRole('heading', { name: 'Frota' })).toBeInTheDocument();
    expect(screen.getByText('Gerencie os veiculos da empresa.')).toBeInTheDocument();
    expect(screen.getByText('Conteudo da pagina')).toBeInTheDocument();
  });
});
