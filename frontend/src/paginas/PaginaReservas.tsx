import { useState } from 'react';
import type { Locacao, Veiculo } from '../api';
import Locacoes from '../componentes/Locacoes';
import Page, { type ControleAcaoPagina } from '../componentes/Page';

interface Props {
  empresaId: string;
  veiculos: Veiculo[];
  locacoes: Locacao[];
  aoAtualizar: () => Promise<void>;
}

export default function PaginaReservas({ empresaId, veiculos, locacoes, aoAtualizar }: Props) {
  const [controle, setControle] = useState<ControleAcaoPagina | null>(null);

  return (
    <Page
      titulo="Reservas"
      descricao="Crie e acompanhe os contratos de locacao dos veiculos da frota."
      rightContent={
        <button
          type="button"
          onClick={() => controle?.abrirModal()}
          disabled={!controle || controle.acaoDesabilitada}
        >
          Criar reserva
        </button>
      }
    >
      <Locacoes
        empresaId={empresaId}
        veiculos={veiculos}
        locacoes={locacoes}
        aoAtualizar={aoAtualizar}
        onControleAcao={setControle}
      />
    </Page>
  );
}
