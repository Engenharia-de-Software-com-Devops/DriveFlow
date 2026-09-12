-- btree_gist permite combinar igualdade (veiculo_id) com sobreposicao de
-- intervalo (&&) na mesma constraint de exclusao.
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS locacoes (
    id             TEXT PRIMARY KEY,
    empresa_id     TEXT NOT NULL REFERENCES empresas (id) ON DELETE CASCADE,
    veiculo_id     TEXT NOT NULL REFERENCES veiculos (id) ON DELETE RESTRICT,
    cliente        TEXT NOT NULL CHECK (length(trim(cliente)) > 0),
    inicio         TIMESTAMPTZ NOT NULL,
    fim_previsto   TIMESTAMPTZ NOT NULL,
    devolvido_em   TIMESTAMPTZ,
    valor_previsto BIGINT NOT NULL CHECK (valor_previsto >= 0), -- centavos
    valor_final    BIGINT CHECK (valor_final >= 0),             -- centavos
    status         TEXT NOT NULL DEFAULT 'aberta'
                   CHECK (status IN ('aberta', 'encerrada')),
    CONSTRAINT locacoes_periodo_valido CHECK (fim_previsto > inicio),
    CONSTRAINT locacoes_encerrada_tem_devolucao CHECK (
        (status = 'aberta'    AND devolvido_em IS NULL AND valor_final IS NULL) OR
        (status = 'encerrada' AND devolvido_em IS NOT NULL AND valor_final IS NOT NULL)
    )
);

-- Gargalo 2 do diagnostico do E1: o mesmo veiculo aparecia disponivel para
-- duas reservas ao mesmo tempo. A regra ja e aplicada na camada de dominio;
-- esta constraint a repete no banco, de modo que nem uma corrida entre duas
-- requisicoes simultaneas consegue gravar contratos sobrepostos.
ALTER TABLE locacoes
    ADD CONSTRAINT locacoes_sem_sobreposicao
    EXCLUDE USING gist (
        veiculo_id WITH =,
        tstzrange(inicio, fim_previsto) WITH &&
    ) WHERE (status = 'aberta');

CREATE INDEX IF NOT EXISTS locacoes_empresa_idx ON locacoes (empresa_id);
CREATE INDEX IF NOT EXISTS locacoes_veiculo_ativas_idx
    ON locacoes (veiculo_id) WHERE status = 'aberta';
