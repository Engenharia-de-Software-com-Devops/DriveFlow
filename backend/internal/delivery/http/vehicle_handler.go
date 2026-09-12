package httpdelivery

import "net/http"

type vehicleRequest struct {
	Plate     string `json:"placa"`
	Model     string `json:"modelo"`
	Category  string `json:"categoria"`
	DailyRate int64  `json:"tarifa_diaria"`
}

func (s *Server) createVehicle(w http.ResponseWriter, r *http.Request) {
	var request vehicleRequest
	if !readJSON(w, r, &request) {
		return
	}

	vehicle, err := s.vehicles.Register(
		r.Context(),
		r.PathValue("empresaID"),
		request.Plate,
		request.Model,
		request.Category,
		request.DailyRate,
	)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, vehicle)
}

func (s *Server) listFleet(w http.ResponseWriter, r *http.Request) {
	fleet, err := s.vehicles.ListFleet(r.Context(), r.PathValue("empresaID"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, fleet)
}
