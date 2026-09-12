import { useCallback, useEffect, useState } from 'react';
import './App.css';
import { listarFrota, listarLocacoes, verificarSaude } from './api';
import CadastroEmpresa from './componentes/CadastroEmpresa';
import Frota from './componentes/Frota';
import Locacoes from './componentes/Locacoes';

const CHAVE_EMPRESA = 'driveflow:empresa';

function empresaSalva() {
  try {
    const bruto = window.localStorage.getItem(CHAVE_EMPRESA);
    return bruto ? JSON.parse(bruto) : null;
  } catch {
    return null;
  }
}

export default function App() {
  const [empresa, setEmpresa] = useState(empresaSalva);
  const [veiculos, setVeiculos] = useState([]);
  const [locacoes, setLocacoes] = useState([]);
  const [apiNoAr, setApiNoAr] = useState(null);
  const [erro, setErro] = useState('');

  useEffect(() => {
    verificarSaude()
      .then(() => setApiNoAr(true))
      .catch(() => setApiNoAr(false));
  }, []);

  const carregar = useCallback(async () => {
    if (!empresa) return;
    setErro('');
    try {
      const [frota, contratos] = await Promise.all([
        listarFrota(empresa.id),
        listarLocacoes(empresa.id),
      ]);
      setVeiculos(frota);
      setLocacoes(contratos);
    } catch (falha) {
      setErro(falha.message);
    }
  }, [empresa]);

  useEffect(() => {
    carregar();
  }, [carregar]);

  function selecionarEmpresa(nova) {
    window.localStorage.setItem(CHAVE_EMPRESA, JSON.stringify(nova));
    setEmpresa(nova);
  }

  function trocarEmpresa() {
    window.localStorage.removeItem(CHAVE_EMPRESA);
    setEmpresa(null);
    setVeiculos([]);
    setLocacoes([]);
  }

  return (
    <div className="app">
      <header>
        <h1>DriveFlow</h1>
        <p>Plataforma de locacao de veiculos para empresas</p>
        {apiNoAr === false && (
          <p role="alert" className="erro">
            API indisponivel. Suba o backend antes de usar a aplicacao.
          </p>
        )}
      </header>

      <main>
        {!empresa ? (
          <CadastroEmpresa aoCadastrar={selecionarEmpresa} />
        ) : (
          <>
            <section className="cartao empresa-ativa">
              <h2>{empresa.nome}</h2>
              <p className="ajuda">CNPJ {empresa.cnpj}</p>
              <button type="button" onClick={trocarEmpresa}>
                Trocar de empresa
              </button>
            </section>

            {erro && <p role="alert" className="erro">{erro}</p>}

            <Frota empresaId={empresa.id} veiculos={veiculos} aoAtualizar={carregar} />
            <Locacoes
              empresaId={empresa.id}
              veiculos={veiculos}
              locacoes={locacoes}
              aoAtualizar={carregar}
            />
          </>
        )}
      </main>
    </div>
  );
}
