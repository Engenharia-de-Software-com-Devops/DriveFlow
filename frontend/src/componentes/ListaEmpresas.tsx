import { useEffect, useState } from 'react';
import { ErroApi, listarEmpresas, type Empresa } from '../api';

interface Props {
  aoSelecionar: (empresa: Empresa) => void;
}

function formatarCnpj(cnpj: string) {
  const digitos = cnpj.replace(/\D/g, '');
  if (digitos.length !== 14) return cnpj;
  return digitos.replace(/^(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})$/, '$1.$2.$3/$4-$5');
}

export default function ListaEmpresas({ aoSelecionar }: Props) {
  const [empresas, setEmpresas] = useState<Empresa[] | null>(null);
  const [erro, setErro] = useState('');

  useEffect(() => {
    listarEmpresas()
      .then((lista) => {
        setErro('');
        setEmpresas(lista);
      })
      .catch((falha) => {
        setErro(falha instanceof ErroApi ? falha.message : 'Falha inesperada.');
        setEmpresas([]);
      });
  }, []);

  return (
    <section className="cartao cartao-lista-empresas">
      <h2>Empresas cadastradas</h2>
      <p className="ajuda">Clique em uma empresa para entrar nela.</p>

      <div className="login-lista-corpo">
        {empresas === null && <p className="ajuda">Carregando...</p>}

        {empresas && empresas.length === 0 && !erro && (
          <p className="ajuda">Nenhuma empresa cadastrada</p>
        )}

        {empresas && empresas.length > 0 && (
          <ul className="lista-empresas">
            {empresas.map((empresa) => (
              <li key={empresa.id}>
                <button
                  type="button"
                  className="item-empresa"
                  onClick={() => aoSelecionar(empresa)}
                >
                  <span className="item-empresa-nome">{empresa.nome}</span>
                  <span className="item-empresa-cnpj">CNPJ {formatarCnpj(empresa.cnpj)}</span>
                </button>
              </li>
            ))}
          </ul>
        )}

        {erro && <p role="alert" className="erro">{erro}</p>}
      </div>
    </section>
  );
}
