import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import * as api from './api'

describe('App', () => {
  beforeEach(() => {
    window.localStorage.clear()
    vi.spyOn(api, 'verificarSaude').mockResolvedValue({ status: 'ok' })
    vi.spyOn(api, 'listarEmpresas').mockResolvedValue([])
  })

  it('pede o cadastro da empresa quando nenhuma esta selecionada', async () => {
    render(
      <MemoryRouter>
        <App />
      </MemoryRouter>,
    )

    expect(await screen.findByRole('heading', { name: 'Cadastrar empresa' })).toBeInTheDocument()
    expect(await screen.findByRole('heading', { name: 'Empresas cadastradas' })).toBeInTheDocument()
  })

  it('entra na empresa ao clicar na lista', async () => {
    const usuario = userEvent.setup()
    vi.spyOn(api, 'listarEmpresas').mockResolvedValue([
      { id: 'emp-1', nome: 'Locadora Alfa', cnpj: '12345678000190' },
    ])
    vi.spyOn(api, 'listarFrota').mockResolvedValue([])
    vi.spyOn(api, 'listarLocacoes').mockResolvedValue([])

    render(
      <MemoryRouter>
        <App />
      </MemoryRouter>,
    )

    await usuario.click(await screen.findByRole('button', { name: /Locadora Alfa/i }))

    expect(screen.queryByRole('heading', { name: 'Cadastrar empresa' })).not.toBeInTheDocument()
    expect(await screen.findByRole('heading', { name: 'Frota' })).toBeInTheDocument()
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
