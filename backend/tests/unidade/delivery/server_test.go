// Testes das rotas HTTP: status, formato do corpo e traducao de erro de dominio.
// Sobem o handler sobre o repositorio em memoria, entao rodam sem banco.
package httpdelivery_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	httpdelivery "driveflow/backend/internal/delivery/http"
	"driveflow/backend/internal/repository"
	"driveflow/backend/tests/apoio"
)

func newServer(t *testing.T) *httpdelivery.Server {
	t.Helper()
	quiet := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := apoio.Padrao(repository.NewMemoryRepository())
	return httpdelivery.NewServer(s.Companies, s.Vehicles, s.Rentals, quiet)
}

// call executa uma requisicao contra o servidor e devolve status e corpo decodificado.
func call(t *testing.T, s *httpdelivery.Server, method, route string, body any) (int, map[string]any) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, route, reader)
	resp := httptest.NewRecorder()
	s.ServeHTTP(resp, req)

	var decoded map[string]any
	_ = json.Unmarshal(resp.Body.Bytes(), &decoded)
	return resp.Code, decoded
}

func TestHealth(t *testing.T) {
	status, body := call(t, newServer(t), http.MethodGet, "/health", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, esperado 200", status)
	}
	if body["status"] != "ok" {
		t.Errorf("status do corpo = %v, esperado ok", body["status"])
	}
}

// Fluxo completo do E1: cadastrar empresa, cadastrar frota, reservar e devolver.
func TestFullRentalFlow(t *testing.T) {
	s := newServer(t)

	status, company := call(t, s, http.MethodPost, "/api/empresas", map[string]any{
		"nome": "Locadora Alfa", "cnpj": "12345678000190",
	})
	if status != http.StatusCreated {
		t.Fatalf("criar empresa: status = %d, esperado 201", status)
	}
	companyID, _ := company["id"].(string)

	status, vehicle := call(t, s, http.MethodPost, "/api/empresas/"+companyID+"/veiculos", map[string]any{
		"placa": "ABC1D23", "modelo": "Onix 1.0", "categoria": "economico", "tarifa_diaria": 15000,
	})
	if status != http.StatusCreated {
		t.Fatalf("criar veiculo: status = %d, esperado 201", status)
	}
	vehicleID, _ := vehicle["id"].(string)

	status, rental := call(t, s, http.MethodPost, "/api/empresas/"+companyID+"/locacoes", map[string]any{
		"veiculo_id": vehicleID, "cliente": "Cliente A",
		"inicio": "2026-03-10T10:00:00Z", "fim_previsto": "2026-03-13T10:00:00Z",
	})
	if status != http.StatusCreated {
		t.Fatalf("criar locacao: status = %d, esperado 201", status)
	}
	if rental["valor_previsto"] != float64(45000) {
		t.Errorf("valor previsto = %v, esperado 45000", rental["valor_previsto"])
	}
	rentalID, _ := rental["id"].(string)

	status, closed := call(t, s, http.MethodPost,
		"/api/empresas/"+companyID+"/locacoes/"+rentalID+"/devolucao",
		map[string]any{"devolvido_em": "2026-03-14T10:00:00Z"})
	if status != http.StatusOK {
		t.Fatalf("devolucao: status = %d, esperado 200", status)
	}
	if closed["status"] != "encerrada" {
		t.Errorf("status da locacao = %v, esperado encerrada", closed["status"])
	}
	if closed["valor_final"] != float64(64500) {
		t.Errorf("valor final = %v, esperado 64500", closed["valor_final"])
	}
}

func TestConflictingReservationReturns409(t *testing.T) {
	s := newServer(t)

	_, company := call(t, s, http.MethodPost, "/api/empresas", map[string]any{
		"nome": "Locadora Alfa", "cnpj": "12345678000190",
	})
	companyID, _ := company["id"].(string)

	_, vehicle := call(t, s, http.MethodPost, "/api/empresas/"+companyID+"/veiculos", map[string]any{
		"placa": "ABC1D23", "modelo": "Onix 1.0", "categoria": "economico", "tarifa_diaria": 15000,
	})
	vehicleID, _ := vehicle["id"].(string)

	reservation := map[string]any{
		"veiculo_id": vehicleID, "cliente": "Cliente A",
		"inicio": "2026-03-10T10:00:00Z", "fim_previsto": "2026-03-14T10:00:00Z",
	}
	if status, _ := call(t, s, http.MethodPost, "/api/empresas/"+companyID+"/locacoes", reservation); status != http.StatusCreated {
		t.Fatalf("primeira reserva: status = %d, esperado 201", status)
	}

	reservation["cliente"] = "Cliente B"
	reservation["inicio"] = "2026-03-12T10:00:00Z"
	reservation["fim_previsto"] = "2026-03-16T10:00:00Z"
	status, body := call(t, s, http.MethodPost, "/api/empresas/"+companyID+"/locacoes", reservation)
	if status != http.StatusConflict {
		t.Fatalf("reserva conflitante: status = %d, esperado 409 (corpo: %v)", status, body)
	}
}

func TestFleetOfUnknownCompanyReturns404(t *testing.T) {
	status, _ := call(t, newServer(t), http.MethodGet, "/api/empresas/nao-existe/veiculos", nil)
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, esperado 404", status)
	}
}

func TestVehicleWithInvalidDataReturns422(t *testing.T) {
	s := newServer(t)

	_, company := call(t, s, http.MethodPost, "/api/empresas", map[string]any{
		"nome": "Locadora Alfa", "cnpj": "12345678000190",
	})
	companyID, _ := company["id"].(string)

	status, _ := call(t, s, http.MethodPost, "/api/empresas/"+companyID+"/veiculos", map[string]any{
		"placa": "XX", "modelo": "Onix", "categoria": "economico", "tarifa_diaria": 15000,
	})
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, esperado 422", status)
	}
}
