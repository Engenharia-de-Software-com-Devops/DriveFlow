package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"driveflow/backend/internal/api"
	"driveflow/backend/internal/armazenamento"
	"driveflow/backend/internal/locacao"
)

func novoServidor(t *testing.T) *api.Servidor {
	t.Helper()
	silencioso := slog.New(slog.NewTextHandler(io.Discard, nil))
	servico := locacao.NovoServico(armazenamento.NovaMemoria(), nil, nil)
	return api.NovoServidor(servico, silencioso)
}

// chamar executa uma requisicao contra o servidor e devolve status e corpo decodificado.
func chamar(t *testing.T, s *api.Servidor, metodo, rota string, corpo any) (int, map[string]any) {
	t.Helper()

	var leitor io.Reader
	if corpo != nil {
		bruto, err := json.Marshal(corpo)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		leitor = bytes.NewReader(bruto)
	}

	req := httptest.NewRequest(metodo, rota, leitor)
	resp := httptest.NewRecorder()
	s.ServeHTTP(resp, req)

	var decodificado map[string]any
	_ = json.Unmarshal(resp.Body.Bytes(), &decodificado)
	return resp.Code, decodificado
}

func TestHealth(t *testing.T) {
	status, corpo := chamar(t, novoServidor(t), http.MethodGet, "/health", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, esperado 200", status)
	}
	if corpo["status"] != "ok" {
		t.Errorf("status do corpo = %v, esperado ok", corpo["status"])
	}
}

// Fluxo completo do E1: cadastrar empresa, cadastrar frota, reservar e devolver.
func TestFluxoCompletoDeLocacao(t *testing.T) {
	s := novoServidor(t)

	status, empresa := chamar(t, s, http.MethodPost, "/api/empresas", map[string]any{
		"nome": "Locadora Alfa", "cnpj": "12345678000190",
	})
	if status != http.StatusCreated {
		t.Fatalf("criar empresa: status = %d, esperado 201", status)
	}
	empresaID, _ := empresa["id"].(string)

	status, veiculo := chamar(t, s, http.MethodPost, "/api/empresas/"+empresaID+"/veiculos", map[string]any{
		"placa": "ABC1D23", "modelo": "Onix 1.0", "categoria": "economico", "tarifa_diaria": 15000,
	})
	if status != http.StatusCreated {
		t.Fatalf("criar veiculo: status = %d, esperado 201", status)
	}
	veiculoID, _ := veiculo["id"].(string)

	status, contrato := chamar(t, s, http.MethodPost, "/api/empresas/"+empresaID+"/locacoes", map[string]any{
		"veiculo_id": veiculoID, "cliente": "Cliente A",
		"inicio": "2026-03-10T10:00:00Z", "fim_previsto": "2026-03-13T10:00:00Z",
	})
	if status != http.StatusCreated {
		t.Fatalf("criar locacao: status = %d, esperado 201", status)
	}
	if contrato["valor_previsto"] != float64(45000) {
		t.Errorf("valor previsto = %v, esperado 45000", contrato["valor_previsto"])
	}
	locacaoID, _ := contrato["id"].(string)

	status, encerrado := chamar(t, s, http.MethodPost,
		"/api/empresas/"+empresaID+"/locacoes/"+locacaoID+"/devolucao",
		map[string]any{"devolvido_em": "2026-03-14T10:00:00Z"})
	if status != http.StatusOK {
		t.Fatalf("devolucao: status = %d, esperado 200", status)
	}
	if encerrado["status"] != "encerrada" {
		t.Errorf("status da locacao = %v, esperado encerrada", encerrado["status"])
	}
	if encerrado["valor_final"] != float64(64500) {
		t.Errorf("valor final = %v, esperado 64500", encerrado["valor_final"])
	}
}

func TestReservaConflitanteRetorna409(t *testing.T) {
	s := novoServidor(t)

	_, empresa := chamar(t, s, http.MethodPost, "/api/empresas", map[string]any{
		"nome": "Locadora Alfa", "cnpj": "12345678000190",
	})
	empresaID, _ := empresa["id"].(string)

	_, veiculo := chamar(t, s, http.MethodPost, "/api/empresas/"+empresaID+"/veiculos", map[string]any{
		"placa": "ABC1D23", "modelo": "Onix 1.0", "categoria": "economico", "tarifa_diaria": 15000,
	})
	veiculoID, _ := veiculo["id"].(string)

	reserva := map[string]any{
		"veiculo_id": veiculoID, "cliente": "Cliente A",
		"inicio": "2026-03-10T10:00:00Z", "fim_previsto": "2026-03-14T10:00:00Z",
	}
	if status, _ := chamar(t, s, http.MethodPost, "/api/empresas/"+empresaID+"/locacoes", reserva); status != http.StatusCreated {
		t.Fatalf("primeira reserva: status = %d, esperado 201", status)
	}

	reserva["cliente"] = "Cliente B"
	reserva["inicio"] = "2026-03-12T10:00:00Z"
	reserva["fim_previsto"] = "2026-03-16T10:00:00Z"
	status, corpo := chamar(t, s, http.MethodPost, "/api/empresas/"+empresaID+"/locacoes", reserva)
	if status != http.StatusConflict {
		t.Fatalf("reserva conflitante: status = %d, esperado 409 (corpo: %v)", status, corpo)
	}
}

func TestFrotaDeEmpresaInexistenteRetorna404(t *testing.T) {
	status, _ := chamar(t, novoServidor(t), http.MethodGet, "/api/empresas/nao-existe/veiculos", nil)
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, esperado 404", status)
	}
}

func TestVeiculoComDadosInvalidosRetorna422(t *testing.T) {
	s := novoServidor(t)

	_, empresa := chamar(t, s, http.MethodPost, "/api/empresas", map[string]any{
		"nome": "Locadora Alfa", "cnpj": "12345678000190",
	})
	empresaID, _ := empresa["id"].(string)

	status, _ := chamar(t, s, http.MethodPost, "/api/empresas/"+empresaID+"/veiculos", map[string]any{
		"placa": "XX", "modelo": "Onix", "categoria": "economico", "tarifa_diaria": 15000,
	})
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, esperado 422", status)
	}
}
