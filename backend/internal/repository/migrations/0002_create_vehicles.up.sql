-- Frota de cada empresa. A placa e unica por empresa, nao globalmente:
-- empresas diferentes podem, em tese, cadastrar a mesma placa.
CREATE TABLE IF NOT EXISTS veiculos (
    id            TEXT PRIMARY KEY,
    empresa_id    TEXT NOT NULL REFERENCES empresas (id) ON DELETE CASCADE,
    placa         VARCHAR(7) NOT NULL CHECK (placa ~ '^[A-Z]{3}[0-9][0-9A-Z][0-9]{2}$'),
    modelo        TEXT NOT NULL CHECK (length(trim(modelo)) > 0),
    categoria     TEXT NOT NULL CHECK (length(trim(categoria)) > 0),
    tarifa_diaria BIGINT NOT NULL CHECK (tarifa_diaria > 0), -- centavos
    status        TEXT NOT NULL DEFAULT 'disponivel'
                  CHECK (status IN ('disponivel', 'manutencao')),
    CONSTRAINT veiculos_placa_por_empresa UNIQUE (empresa_id, placa)
);

CREATE INDEX IF NOT EXISTS veiculos_empresa_idx ON veiculos (empresa_id);
