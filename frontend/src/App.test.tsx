import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import * as api from './api'

describe('App', () => {
  beforeEach(() => {
    window.localStorage.clear()
    vi.spyOn(api, 'verificarSaude').mockResolvedValue({ status: 'ok' })
  })

  it('pede o cadastro da empresa quando nenhuma esta selecionada', async () => {
    render(
      <MemoryRouter>
        <App />
      </MemoryRouter>,
    )

    expect(await screen.findByRole('heading', { name: 'Cadastrar empresa' })).toBeInTheDocument()
  })
})
