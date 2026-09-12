// Package api expoe o dominio de locacao como uma API HTTP JSON.
package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"driveflow/backend/internal/locacao"
)

// Versao da aplicacao, reportada em GET /health.
var Versao = "0.1.0"

// Servidor conecta as rotas HTTP ao servico de dominio.
type Servidor struct {
	servico *locacao.Servico
	log     *slog.Logger
	mux     *http.ServeMux
}

// NovoServidor monta o roteador com todas as rotas registradas.
func NovoServidor(servico *locacao.Servico, log *slog.Logger) *Servidor {
	if log == nil {
		log = slog.Default()
	}
	s := &Servidor{servico: servico, log: log, mux: http.NewServeMux()}
	s.registrarRotas()
	return s
}

func (s *Servidor) registrarRotas() {
	s.mux.HandleFunc("GET /health", s.saude)
	s.mux.HandleFunc("POST /api/empresas", s.criarEmpresa)
	s.mux.HandleFunc("GET /api/empresas/{empresaID}/veiculos", s.listarFrota)
	s.mux.HandleFunc("POST /api/empresas/{empresaID}/veiculos", s.criarVeiculo)
	s.mux.HandleFunc("GET /api/empresas/{empresaID}/locacoes", s.listarLocacoes)
	s.mux.HandleFunc("POST /api/empresas/{empresaID}/locacoes", s.criarLocacao)
	s.mux.HandleFunc("POST /api/empresas/{empresaID}/locacoes/{locacaoID}/devolucao", s.devolver)
}

// ServeHTTP aplica CORS e registra cada requisicao antes de rotear.
func (s *Servidor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inicio := time.Now()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.mux.ServeHTTP(w, r)
	s.log.Info("requisicao", "metodo", r.Method, "rota", r.URL.Path, "duracao", time.Since(inicio).String())
}

type respostaSaude struct {
	Status string `json:"status"`
	Versao string `json:"versao"`
}

func (s *Servidor) saude(w http.ResponseWriter, _ *http.Request) {
	escreverJSON(w, http.StatusOK, respostaSaude{Status: "ok", Versao: Versao})
}

type pedidoEmpresa struct {
	Nome string `json:"nome"`
	CNPJ string `json:"cnpj"`
}

func (s *Servidor) criarEmpresa(w http.ResponseWriter, r *http.Request) {
	var pedido pedidoEmpresa
	if !lerJSON(w, r, &pedido) {
		return
	}

	empresa, err := s.servico.CadastrarEmpresa(r.Context(), pedido.Nome, pedido.CNPJ)
	if err != nil {
		s.escreverErro(w, err)
		return
	}
	escreverJSON(w, http.StatusCreated, empresa)
}

type pedidoVeiculo struct {
	Placa        string `json:"placa"`
	Modelo       string `json:"modelo"`
	Categoria    string `json:"categoria"`
	TarifaDiaria int64  `json:"tarifa_diaria"`
}

func (s *Servidor) criarVeiculo(w http.ResponseWriter, r *http.Request) {
	var pedido pedidoVeiculo
	if !lerJSON(w, r, &pedido) {
		return
	}

	veiculo, err := s.servico.CadastrarVeiculo(
		r.Context(),
		r.PathValue("empresaID"),
		pedido.Placa,
		pedido.Modelo,
		pedido.Categoria,
		pedido.TarifaDiaria,
	)
	if err != nil {
		s.escreverErro(w, err)
		return
	}
	escreverJSON(w, http.StatusCreated, veiculo)
}

func (s *Servidor) listarFrota(w http.ResponseWriter, r *http.Request) {
	frota, err := s.servico.ListarFrota(r.Context(), r.PathValue("empresaID"))
	if err != nil {
		s.escreverErro(w, err)
		return
	}
	escreverJSON(w, http.StatusOK, frota)
}

type pedidoLocacao struct {
	VeiculoID   string    `json:"veiculo_id"`
	Cliente     string    `json:"cliente"`
	Inicio      time.Time `json:"inicio"`
	FimPrevisto time.Time `json:"fim_previsto"`
}

func (s *Servidor) criarLocacao(w http.ResponseWriter, r *http.Request) {
	var pedido pedidoLocacao
	if !lerJSON(w, r, &pedido) {
		return
	}

	contrato, err := s.servico.ReservarVeiculo(
		r.Context(),
		r.PathValue("empresaID"),
		pedido.VeiculoID,
		pedido.Cliente,
		pedido.Inicio,
		pedido.FimPrevisto,
	)
	if err != nil {
		s.escreverErro(w, err)
		return
	}
	escreverJSON(w, http.StatusCreated, contrato)
}

type pedidoDevolucao struct {
	DevolvidoEm time.Time `json:"devolvido_em"`
}

func (s *Servidor) devolver(w http.ResponseWriter, r *http.Request) {
	var pedido pedidoDevolucao
	if r.ContentLength > 0 && !lerJSON(w, r, &pedido) {
		return
	}

	contrato, err := s.servico.DevolverVeiculo(
		r.Context(),
		r.PathValue("empresaID"),
		r.PathValue("locacaoID"),
		pedido.DevolvidoEm,
	)
	if err != nil {
		s.escreverErro(w, err)
		return
	}
	escreverJSON(w, http.StatusOK, contrato)
}

func (s *Servidor) listarLocacoes(w http.ResponseWriter, r *http.Request) {
	contratos, err := s.servico.ListarLocacoes(r.Context(), r.PathValue("empresaID"))
	if err != nil {
		s.escreverErro(w, err)
		return
	}
	escreverJSON(w, http.StatusOK, contratos)
}

type respostaErro struct {
	Erro string `json:"erro"`
}

// escreverErro traduz os erros de dominio em codigos HTTP.
func (s *Servidor) escreverErro(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, locacao.ErrNaoEncontrado):
		escreverJSON(w, http.StatusNotFound, respostaErro{Erro: err.Error()})
	case errors.Is(err, locacao.ErrConflitoReserva),
		errors.Is(err, locacao.ErrPlacaDuplicada),
		errors.Is(err, locacao.ErrVeiculoIndisponivel),
		errors.Is(err, locacao.ErrLocacaoEncerrada):
		escreverJSON(w, http.StatusConflict, respostaErro{Erro: err.Error()})
	case errors.Is(err, locacao.ErrDadosInvalidos):
		escreverJSON(w, http.StatusUnprocessableEntity, respostaErro{Erro: err.Error()})
	default:
		s.log.Error("falha inesperada", "erro", err)
		escreverJSON(w, http.StatusInternalServerError, respostaErro{Erro: "erro interno"})
	}
}

func escreverJSON(w http.ResponseWriter, status int, corpo any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(corpo); err != nil {
		slog.Error("falha ao escrever resposta", "erro", err)
	}
}

// lerJSON decodifica o corpo da requisicao e ja responde 400 em caso de erro.
func lerJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	if err := json.NewDecoder(r.Body).Decode(destino); err != nil {
		escreverJSON(w, http.StatusBadRequest, respostaErro{Erro: "corpo json invalido"})
		return false
	}
	return true
}
