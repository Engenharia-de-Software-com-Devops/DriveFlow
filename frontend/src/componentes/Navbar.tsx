import { useEffect, useRef, useState } from 'react';
import { NavLink } from 'react-router-dom';
import type { Empresa } from '../api';

interface Props {
  empresa: Empresa | null;
  versao: string | null;
  onSair: () => void;
}

function formatarCnpj(cnpj: string) {
  const digitos = cnpj.replace(/\D/g, '');
  if (digitos.length !== 14) return cnpj;
  return digitos.replace(/^(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})$/, '$1.$2.$3/$4-$5');
}

function classeLink({ isActive }: { isActive: boolean }) {
  return isActive ? 'navbar-link navbar-link-ativa' : 'navbar-link';
}

export default function Navbar({ empresa, versao, onSair }: Props) {
  const [aberto, setAberto] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!aberto) return;

    function aoClicarFora(evento: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(evento.target as Node)) {
        setAberto(false);
      }
    }

    function aoTeclar(evento: KeyboardEvent) {
      if (evento.key === 'Escape') setAberto(false);
    }

    document.addEventListener('mousedown', aoClicarFora);
    document.addEventListener('keydown', aoTeclar);
    return () => {
      document.removeEventListener('mousedown', aoClicarFora);
      document.removeEventListener('keydown', aoTeclar);
    };
  }, [aberto]);

  function sair() {
    setAberto(false);
    onSair();
  }

  return (
    <nav className="navbar">
      <div className="navbar-inner">
        <div className="navbar-esquerda">
          <div className="navbar-logo">
            <span className="navbar-marca">DriveFlow</span>
          </div>

          {empresa && (
            <div className="navbar-links">
              <NavLink to="/frota" className={classeLink}>
                Frota
              </NavLink>
              <NavLink to="/reservas" className={classeLink}>
                Reservas
              </NavLink>
            </div>
          )}
        </div>

        <div className="navbar-direita">
          <span className="navbar-versao">v{versao ?? '—'}</span>
          {empresa && (
            <div className="navbar-menu" ref={menuRef}>
              <button
                type="button"
                className="navbar-trigger"
                aria-expanded={aberto}
                aria-haspopup="true"
                onClick={() => setAberto((atual) => !atual)}
              >
                {empresa.nome}
                <span className="navbar-chevron" aria-hidden="true">
                  ▾
                </span>
              </button>

              {aberto && (
                <div className="navbar-dropdown" role="menu">
                  <p className="navbar-dropdown-cnpj">CNPJ {formatarCnpj(empresa.cnpj)}</p>
                  <button type="button" className="botao-secundario navbar-sair" onClick={sair}>
                    Sair
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </nav>
  );
}
