package armazenamento

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
)

//go:embed migracoes/*.sql
var arquivosMigracao embed.FS

// Migracao e um arquivo .up.sql versionado, aplicado uma unica vez.
type Migracao struct {
	Versao string
	Nome   string
	SQL    string
}

// CarregarMigracoes le as migrations embutidas no binario, em ordem de versao.
// Manter os arquivos versionados e embutidos elimina o passo manual de aplicar
// .sql soltos no banco, apontado como Gargalo 1 no diagnostico do E1.
func CarregarMigracoes() ([]Migracao, error) {
	entradas, err := fs.ReadDir(arquivosMigracao, "migracoes")
	if err != nil {
		return nil, fmt.Errorf("ler diretorio de migracoes: %w", err)
	}

	var migracoes []Migracao
	for _, entrada := range entradas {
		nome := entrada.Name()
		if !strings.HasSuffix(nome, ".up.sql") {
			continue
		}

		conteudo, err := arquivosMigracao.ReadFile("migracoes/" + nome)
		if err != nil {
			return nil, fmt.Errorf("ler migracao %s: %w", nome, err)
		}

		versao, _, encontrado := strings.Cut(nome, "_")
		if !encontrado {
			return nil, fmt.Errorf("migracao %s nao segue o padrao NNNN_descricao.up.sql", nome)
		}

		migracoes = append(migracoes, Migracao{
			Versao: versao,
			Nome:   strings.TrimSuffix(nome, ".up.sql"),
			SQL:    string(conteudo),
		})
	}

	sort.Slice(migracoes, func(i, j int) bool { return migracoes[i].Versao < migracoes[j].Versao })
	return migracoes, nil
}

// AplicarMigracoes executa as migrations pendentes em ordem.
//
// Cada migration roda dentro da propria transacao junto com o registro em
// schema_migracoes: ou a mudanca de schema e o registro entram juntos, ou
// nenhum dos dois entra. Isso torna a aplicacao idempotente e segura para
// rodar na subida de todo container da api.
func AplicarMigracoes(db *sql.DB, log *slog.Logger) error {
	if log == nil {
		log = slog.Default()
	}

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migracoes (
			versao      TEXT PRIMARY KEY,
			nome        TEXT NOT NULL,
			aplicada_em TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("criar tabela de controle: %w", err)
	}

	aplicadas, err := versoesAplicadas(db)
	if err != nil {
		return err
	}

	migracoes, err := CarregarMigracoes()
	if err != nil {
		return err
	}

	pendentes := 0
	for _, m := range migracoes {
		if aplicadas[m.Versao] {
			continue
		}
		if err := aplicarUma(db, m); err != nil {
			return err
		}
		log.Info("migracao aplicada", "versao", m.Versao, "nome", m.Nome)
		pendentes++
	}

	if pendentes == 0 {
		log.Info("schema ja atualizado", "migracoes", len(migracoes))
	}
	return nil
}

func aplicarUma(db *sql.DB, m Migracao) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("iniciar transacao da migracao %s: %w", m.Versao, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(m.SQL); err != nil {
		return fmt.Errorf("aplicar migracao %s: %w", m.Nome, err)
	}
	if _, err := tx.Exec(`INSERT INTO schema_migracoes (versao, nome) VALUES ($1, $2)`, m.Versao, m.Nome); err != nil {
		return fmt.Errorf("registrar migracao %s: %w", m.Nome, err)
	}
	return tx.Commit()
}

func versoesAplicadas(db *sql.DB) (map[string]bool, error) {
	linhas, err := db.Query(`SELECT versao FROM schema_migracoes`)
	if err != nil {
		return nil, fmt.Errorf("consultar migracoes aplicadas: %w", err)
	}
	defer func() { _ = linhas.Close() }()

	aplicadas := make(map[string]bool)
	for linhas.Next() {
		var versao string
		if err := linhas.Scan(&versao); err != nil {
			return nil, fmt.Errorf("ler versao aplicada: %w", err)
		}
		aplicadas[versao] = true
	}
	return aplicadas, linhas.Err()
}
