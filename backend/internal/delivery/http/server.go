// Package httpdelivery expoe os casos de uso como uma API HTTP JSON.
//
// E a camada mais externa: traduz requisicao HTTP em chamada de caso de uso e
// erro de dominio em codigo de status. Nenhuma regra de negocio mora aqui.
//
// O nome do pacote difere do diretorio (delivery/http) de proposito, para nao
// colidir com net/http em quem importa.
package httpdelivery

import (
	"log/slog"
	"net/http"
	"time"

	"driveflow/backend/internal/usecases"
)

// Version da aplicacao, reportada em GET /health.
var Version = "0.1.0"

// Server conecta as rotas HTTP aos servicos de caso de uso.
type Server struct {
	companies *usecases.CompanyService
	vehicles  *usecases.VehicleService
	rentals   *usecases.RentalService
	log       *slog.Logger
	mux       *http.ServeMux
}

// NewServer monta o roteador com todas as rotas registradas.
func NewServer(
	companies *usecases.CompanyService,
	vehicles *usecases.VehicleService,
	rentals *usecases.RentalService,
	log *slog.Logger,
) *Server {
	if log == nil {
		log = slog.Default()
	}
	s := &Server{
		companies: companies,
		vehicles:  vehicles,
		rentals:   rentals,
		log:       log,
		mux:       http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// registerRoutes mantem as rotas em um unico lugar. Os caminhos seguem em
// portugues porque sao o contrato publico ja consumido pelo frontend.
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("POST /api/empresas", s.createCompany)
	s.mux.HandleFunc("GET /api/empresas/{empresaID}/veiculos", s.listFleet)
	s.mux.HandleFunc("POST /api/empresas/{empresaID}/veiculos", s.createVehicle)
	s.mux.HandleFunc("GET /api/empresas/{empresaID}/locacoes", s.listRentals)
	s.mux.HandleFunc("POST /api/empresas/{empresaID}/locacoes", s.createRental)
	s.mux.HandleFunc("POST /api/empresas/{empresaID}/locacoes/{locacaoID}/devolucao", s.returnRental)
}

// ServeHTTP aplica CORS e registra cada requisicao antes de rotear.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.mux.ServeHTTP(w, r)
	s.log.Info("requisicao", "metodo", r.Method, "rota", r.URL.Path, "duracao", time.Since(start).String())
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"versao"`
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Version: Version})
}
