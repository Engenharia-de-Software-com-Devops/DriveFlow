import { useState } from 'react';
import type { Veiculo } from '../api';
import Frota from '../componentes/Frota';
import Page, { type ControleAcaoPagina } from '../componentes/Page';

interface Props {
  empresaId: string;
  veiculos: Veiculo[];
  aoAtualizar: () => Promise<void>;
}

export default function PaginaFrota({ empresaId, veiculos, aoAtualizar }: Props) {
  const [controle, setControle] = useState<ControleAcaoPagina | null>(null);

  return (
    <Page
      titulo="Frota"
      descricao="Gerencie os veiculos da empresa e cadastre novos modelos para locacao."
      rightContent={
        <button
          type="button"
          onClick={() => controle?.abrirModal()}
          disabled={!controle || controle.acaoDesabilitada}
        >
          Adicionar veiculo
        </button>
      }
    >
      <Frota
        empresaId={empresaId}
        veiculos={veiculos}
        aoAtualizar={aoAtualizar}
        onControleAcao={setControle}
      />
    </Page>
  );
}
