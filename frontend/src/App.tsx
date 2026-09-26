import { useCallback, useEffect, useState } from 'react';
import { Navigate, Route, Routes, useNavigate } from 'react-router-dom';
import './App.css';
import { listarFrota, listarLocacoes, verificarSaude, type Empresa, type Locacao, type Veiculo } from './api';
import CadastroEmpresa from './componentes/CadastroEmpresa';
import ListaEmpresas from './componentes/ListaEmpresas';
import Navbar from './componentes/Navbar';
import PaginaFrota from './paginas/PaginaFrota';
import PaginaReservas from './paginas/PaginaReservas';

const CHAVE_EMPRESA = 'driveflow:empresa';

function empresaSalva(): Empresa | null {
  try {
    const bruto = window.localStorage.getItem(CHAVE_EMPRESA);
    return bruto ? JSON.parse(bruto) : null;
  } catch {
    return null;
  }
}

export default function App() {
  const navigate = useNavigate();
  const [empresa, setEmpresa] = useState<Empresa | null>(empresaSalva);
  const [veiculos, setVeiculos] = useState<Veiculo[]>([]);
  const [locacoes, setLocacoes] = useState<Locacao[]>([]);
  const [apiNoAr, setApiNoAr] = useState<boolean | null>(null);
  const [versao, setVersao] = useState<string | null>(null);
  const [erro, setErro] = useState('');

  useEffect(() => {
    verificarSaude()
      .then((saude) => {
        setApiNoAr(true);
        setVersao(saude.versao);
      })
      .catch(() => setApiNoAr(false));
  }, []);

  // State is only updated inside the promise callbacks, so calling this from an
  // effect never triggers a synchronous re-render.
  const carregar = useCallback(async () => {
    if (!empresa) return;
    await Promise.all([listarFrota(empresa.id), listarLocacoes(empresa.id)])
      .then(([frota, reservas]) => {
        setErro('');
        setVeiculos(frota);
        setLocacoes(reservas);
      })
      .catch((falha) => {
        setErro(falha instanceof Error ? falha.message : 'Falha inesperada.');
      });
  }, [empresa]);

  useEffect(() => {
    carregar();
  }, [carregar]);

  function selecionarEmpresa(nova: Empresa) {
    window.localStorage.setItem(CHAVE_EMPRESA, JSON.stringify(nova));
    setEmpresa(nova);
    navigate('/frota');
  }

  function trocarEmpresa() {
    window.localStorage.removeItem(CHAVE_EMPRESA);
    setEmpresa(null);
    setVeiculos([]);
    setLocacoes([]);
    navigate('/');
  }

  return (
    <div className="app-shell">
      <Navbar empresa={empresa} versao={versao} onSair={trocarEmpresa} />

      <div className="app-conteudo">
        {apiNoAr === false && (
          <p role="alert" className="erro aviso-api">
            API indisponivel. Suba o backend antes de usar a aplicacao.
          </p>
        )}

        <main>
          <Routes>
            <Route
              path="/"
              element={
                empresa ? (
                  <Navigate to="/frota" replace />
                ) : (
                  <div className="login-palco">
                    <div className="login-cadastro">
                      <CadastroEmpresa aoCadastrar={selecionarEmpresa} />
                    </div>
                    <div className="login-lista">
                      <ListaEmpresas aoSelecionar={selecionarEmpresa} />
                    </div>
                  </div>
                )
              }
            />
            <Route
              path="/frota"
              element={
                empresa ? (
                  <>
                    {erro && <p role="alert" className="erro">{erro}</p>}
                    <PaginaFrota
                      empresaId={empresa.id}
                      veiculos={veiculos}
                      aoAtualizar={carregar}
                    />
                  </>
                ) : (
                  <Navigate to="/" replace />
                )
              }
            />
            <Route
              path="/reservas"
              element={
                empresa ? (
                  <>
                    {erro && <p role="alert" className="erro">{erro}</p>}
                    <PaginaReservas
                      empresaId={empresa.id}
                      veiculos={veiculos}
                      locacoes={locacoes}
                      aoAtualizar={carregar}
                    />
                  </>
                ) : (
                  <Navigate to="/" replace />
                )
              }
            />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>
      </div>
    </div>
  );
}
