import type { ReactNode } from 'react';

export interface ControleAcaoPagina {
  abrirModal: () => void;
  acaoDesabilitada?: boolean;
}

interface Props {
  titulo: string;
  descricao: string;
  rightContent?: ReactNode;
  children: ReactNode;
}

export default function Page({ titulo, descricao, rightContent, children }: Props) {
  return (
    <div className="pagina">
      <header className="pagina-cabecalho">
        <div className="pagina-intro">
          <h1>{titulo}</h1>
          <p className="ajuda">{descricao}</p>
        </div>
        {rightContent && <div className="pagina-direita">{rightContent}</div>}
      </header>
      <div className="pagina-conteudo">{children}</div>
    </div>
  );
}
