-- Empresas sao os tenants da plataforma. Todo o restante do schema referencia
-- esta tabela, o que torna o isolamento entre empresas uma garantia do banco.
CREATE TABLE IF NOT EXISTS empresas (
    id        TEXT PRIMARY KEY,
    nome      TEXT NOT NULL CHECK (length(trim(nome)) > 0),
    cnpj      CHAR(14) NOT NULL UNIQUE CHECK (cnpj ~ '^[0-9]{14}$'),
    criada_em TIMESTAMPTZ NOT NULL DEFAULT now()
);
