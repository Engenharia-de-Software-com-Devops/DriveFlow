package httpdelivery

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"driveflow/backend/internal/entities"
)

type errorResponse struct {
	Error string `json:"erro"`
}

// writeError traduz os erros de dominio em codigos HTTP.
func (s *Server) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, entities.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, entities.ErrBookingConflict),
		errors.Is(err, entities.ErrDuplicatePlate),
		errors.Is(err, entities.ErrVehicleUnavailable),
		errors.Is(err, entities.ErrRentalClosed):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, entities.ErrInvalidData):
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
	default:
		s.log.Error("falha inesperada", "erro", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "erro interno"})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("falha ao escrever resposta", "erro", err)
	}
}

// readJSON decodifica o corpo da requisicao e ja responde 400 em caso de erro.
func readJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "corpo json invalido"})
		return false
	}
	return true
}
