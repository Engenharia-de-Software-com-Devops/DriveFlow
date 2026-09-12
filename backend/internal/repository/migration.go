package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// migrationsTable e a tabela de controle criada no proprio banco.
// O nome segue em portugues porque bancos ja existentes registram as versoes
// aplicadas nela: renomear faria a api tentar reaplicar todo o schema.
const migrationsTable = "schema_migracoes"

// Migration e um arquivo .up.sql versionado, aplicado uma unica vez.
type Migration struct {
	Version string
	Name    string
	SQL     string
}

// LoadMigrations le as migrations embutidas no binario, em ordem de versao.
// Manter os arquivos versionados e embutidos elimina o passo manual de aplicar
// .sql soltos no banco, apontado como Gargalo 1 no diagnostico do E1.
func LoadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("ler diretorio de migracoes: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return nil, fmt.Errorf("ler migracao %s: %w", name, err)
		}

		version, _, found := strings.Cut(name, "_")
		if !found {
			return nil, fmt.Errorf("migracao %s nao segue o padrao NNNN_descricao.up.sql", name)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    strings.TrimSuffix(name, ".up.sql"),
			SQL:     string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	return migrations, nil
}

// ApplyMigrations executa as migrations pendentes em ordem.
//
// Cada migration roda dentro da propria transacao junto com o registro em
// schema_migracoes: ou a mudanca de schema e o registro entram juntos, ou
// nenhum dos dois entra. Isso torna a aplicacao idempotente e segura para
// rodar na subida de todo container da api.
func ApplyMigrations(db *sql.DB, log *slog.Logger) error {
	if log == nil {
		log = slog.Default()
	}

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ` + migrationsTable + ` (
			versao      TEXT PRIMARY KEY,
			nome        TEXT NOT NULL,
			aplicada_em TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("criar tabela de controle: %w", err)
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}

	migrations, err := LoadMigrations()
	if err != nil {
		return err
	}

	pending := 0
	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}
		if err := applyOne(db, m); err != nil {
			return err
		}
		log.Info("migracao aplicada", "versao", m.Version, "nome", m.Name)
		pending++
	}

	if pending == 0 {
		log.Info("schema ja atualizado", "migracoes", len(migrations))
	}
	return nil
}

func applyOne(db *sql.DB, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("iniciar transacao da migracao %s: %w", m.Version, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(m.SQL); err != nil {
		return fmt.Errorf("aplicar migracao %s: %w", m.Name, err)
	}
	if _, err := tx.Exec(`INSERT INTO `+migrationsTable+` (versao, nome) VALUES ($1, $2)`, m.Version, m.Name); err != nil {
		return fmt.Errorf("registrar migracao %s: %w", m.Name, err)
	}
	return tx.Commit()
}

func appliedVersions(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT versao FROM ` + migrationsTable)
	if err != nil {
		return nil, fmt.Errorf("consultar migracoes aplicadas: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("ler versao aplicada: %w", err)
		}
		applied[version] = true
	}
	return applied, rows.Err()
}
