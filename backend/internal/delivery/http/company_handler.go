package httpdelivery

import (
	"net/http"

	"driveflow/backend/internal/entities"
)

type companyRequest struct {
	Name  string `json:"nome"`
	TaxID string `json:"cnpj"`
}

func (s *Server) createCompany(w http.ResponseWriter, r *http.Request) {
	var request companyRequest
	if !readJSON(w, r, &request) {
		return
	}

	company, err := s.companies.Register(r.Context(), request.Name, request.TaxID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, company)
}

func (s *Server) listCompanies(w http.ResponseWriter, r *http.Request) {
	companies, err := s.companies.List(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	if companies == nil {
		companies = []entities.Company{}
	}
	writeJSON(w, http.StatusOK, companies)
}
