import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import * as api from './api'

describe('App', () => {
  beforeEach(() => {
    window.localStorage.clear()
    vi.spyOn(api, 'verificarSaude').mockResolvedValue({ status: 'ok', versao: '0.1.0' })
  })

  it('pede o cadastro da empresa quando nenhuma esta selecionada', async () => {
    render(
      <MemoryRouter>
        <App />
      </MemoryRouter>,
    )

    expect(await screen.findByRole('heading', { name: 'Cadastrar empresa' })).toBeInTheDocument()
  })

  describe('com empresa salva', () => {
    beforeEach(() => {
      window.localStorage.setItem(
        'driveflow:empresa',
        JSON.stringify({ id: 'emp-1', nome: 'Locadora', cnpj: '12345678000199' }),
      )
    })

    it('carrega a frota da empresa ao abrir', async () => {
      const listarFrota = vi.spyOn(api, 'listarFrota').mockResolvedValue([
        {
          id: 'vei-1',
          placa: 'ABC1D23',
          modelo: 'Onix 1.0',
          categoria: 'economico',
          tarifa_diaria: 15000,
          status: 'disponivel',
        },
      ])
      vi.spyOn(api, 'listarLocacoes').mockResolvedValue([])

      render(
        <MemoryRouter initialEntries={['/frota']}>
          <App />
        </MemoryRouter>,
      )

      expect(await screen.findByText('ABC1D23')).toBeInTheDocument()
      expect(listarFrota).toHaveBeenCalledWith('emp-1')
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })

    it('mostra o erro quando o carregamento falha', async () => {
      vi.spyOn(api, 'listarFrota').mockRejectedValue(new Error('falha ao listar frota'))
      vi.spyOn(api, 'listarLocacoes').mockResolvedValue([])

      render(
        <MemoryRouter initialEntries={['/frota']}>
          <App />
        </MemoryRouter>,
      )

      expect(await screen.findByRole('alert')).toHaveTextContent('falha ao listar frota')
    })
  })
})
