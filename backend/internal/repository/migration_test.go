package repository_test

import (
	"strings"
	"testing"

	"driveflow/backend/internal/repository"
)

// As migrations precisam estar versionadas, em ordem e sem versoes repetidas:
// foi a aplicacao de .sql fora de ordem que o diagnostico do E1 apontou como
// risco de derrubar a plataforma inteira.
func TestMigrationsAreVersionedInOrder(t *testing.T) {
	migrations, err := repository.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("nenhuma migracao carregada")
	}

	seen := make(map[string]string)
	previous := ""
	for _, m := range migrations {
		if m.Version <= previous {
			t.Errorf("migracao %s fora de ordem (anterior: %s)", m.Name, previous)
		}
		if duplicated, exists := seen[m.Version]; exists {
			t.Errorf("versao %s duplicada entre %s e %s", m.Version, duplicated, m.Name)
		}
		if strings.TrimSpace(m.SQL) == "" {
			t.Errorf("migracao %s esta vazia", m.Name)
		}
		seen[m.Version] = m.Name
		previous = m.Version
	}
}

func TestRentalsMigrationBlocksOverlap(t *testing.T) {
	migrations, err := repository.LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations: %v", err)
	}

	for _, m := range migrations {
		if strings.Contains(m.Name, "rentals") {
			if !strings.Contains(m.SQL, "locacoes_sem_sobreposicao") {
				t.Error("a migracao de locacoes precisa manter a constraint de exclusao de periodos")
			}
			return
		}
	}
	t.Fatal("migracao de locacoes nao encontrada")
}
