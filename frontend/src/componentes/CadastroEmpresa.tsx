import { useState, type FormEvent } from 'react';
import { criarEmpresa, ErroApi, type Empresa } from '../api';
import StatusBanner from './StatusBanner';

interface Props {
  aoCadastrar: (empresa: Empresa) => void;
}

// CadastroEmpresa registra o tenant e devolve o id para o restante da tela.
export default function CadastroEmpresa({ aoCadastrar }: Props) {
  const [nome, setNome] = useState('');
  const [cnpj, setCnpj] = useState('');
  const [erro, setErro] = useState('');
  const [enviando, setEnviando] = useState(false);

  async function enviar(evento: FormEvent) {
    evento.preventDefault();
    setErro('');
    setEnviando(true);
    try {
      aoCadastrar(await criarEmpresa(nome, cnpj));
    } catch (falha) {
      setErro(falha instanceof ErroApi ? falha.message : 'Falha inesperada.');
    } finally {
      setEnviando(false);
    }
  }

  return (
    <section className="cartao">
      <h2>Cadastrar empresa</h2>
      <p className="ajuda">
        Cada empresa e um tenant: enxerga apenas a propria frota e as proprias locacoes.
      </p>
      <form onSubmit={enviar}>
        <label htmlFor="empresa-nome">Nome da empresa</label>
        <input
          id="empresa-nome"
          value={nome}
          onChange={(e) => setNome(e.target.value)}
          placeholder="Locadora Alfa"
          required
        />

        <label htmlFor="empresa-cnpj">CNPJ</label>
        <input
          id="empresa-cnpj"
          value={cnpj}
          onChange={(e) => setCnpj(e.target.value)}
          placeholder="12345678000190"
          required
        />

        <button type="submit" disabled={enviando}>
          {enviando ? 'Cadastrando...' : 'Cadastrar empresa'}
        </button>
      </form>
      {erro && <StatusBanner tipo="error" mensagem={erro} />}
    </section>
  );
}
