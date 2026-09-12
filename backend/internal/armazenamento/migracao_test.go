package armazenamento_test

import (
	"strings"
	"testing"

	"driveflow/backend/internal/armazenamento"
)

// As migrations precisam estar versionadas, em ordem e sem versoes repetidas:
// foi a aplicacao de .sql fora de ordem que o diagnostico do E1 apontou como
// risco de derrubar a plataforma inteira.
func TestMigracoesEstaoVersionadasEmOrdem(t *testing.T) {
	migracoes, err := armazenamento.CarregarMigracoes()
	if err != nil {
		t.Fatalf("CarregarMigracoes: %v", err)
	}
	if len(migracoes) == 0 {
		t.Fatal("nenhuma migracao carregada")
	}

	vistas := make(map[string]string)
	anterior := ""
	for _, m := range migracoes {
		if m.Versao <= anterior {
			t.Errorf("migracao %s fora de ordem (anterior: %s)", m.Nome, anterior)
		}
		if duplicada, existe := vistas[m.Versao]; existe {
			t.Errorf("versao %s duplicada entre %s e %s", m.Versao, duplicada, m.Nome)
		}
		if strings.TrimSpace(m.SQL) == "" {
			t.Errorf("migracao %s esta vazia", m.Nome)
		}
		vistas[m.Versao] = m.Nome
		anterior = m.Versao
	}
}

func TestMigracaoDeLocacoesBloqueiaSobreposicao(t *testing.T) {
	migracoes, err := armazenamento.CarregarMigracoes()
	if err != nil {
		t.Fatalf("CarregarMigracoes: %v", err)
	}

	for _, m := range migracoes {
		if strings.Contains(m.Nome, "locacoes") {
			if !strings.Contains(m.SQL, "locacoes_sem_sobreposicao") {
				t.Error("a migracao de locacoes precisa manter a constraint de exclusao de periodos")
			}
			return
		}
	}
	t.Fatal("migracao de locacoes nao encontrada")
}
