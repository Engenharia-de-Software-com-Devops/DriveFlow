package httpdelivery

import (
	"net/http"
	"time"
)

type rentalRequest struct {
	VehicleID   string    `json:"veiculo_id"`
	Customer    string    `json:"cliente"`
	Start       time.Time `json:"inicio"`
	ExpectedEnd time.Time `json:"fim_previsto"`
}

func (s *Server) createRental(w http.ResponseWriter, r *http.Request) {
	var request rentalRequest
	if !readJSON(w, r, &request) {
		return
	}

	rental, err := s.rentals.Reserve(
		r.Context(),
		r.PathValue("empresaID"),
		request.VehicleID,
		request.Customer,
		request.Start,
		request.ExpectedEnd,
	)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rental)
}

type returnRequest struct {
	ReturnedAt time.Time `json:"devolvido_em"`
}

func (s *Server) returnRental(w http.ResponseWriter, r *http.Request) {
	var request returnRequest
	if r.ContentLength > 0 && !readJSON(w, r, &request) {
		return
	}

	rental, err := s.rentals.Return(
		r.Context(),
		r.PathValue("empresaID"),
		r.PathValue("locacaoID"),
		request.ReturnedAt,
	)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rental)
}

func (s *Server) listRentals(w http.ResponseWriter, r *http.Request) {
	rentals, err := s.rentals.List(r.Context(), r.PathValue("empresaID"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rentals)
}
