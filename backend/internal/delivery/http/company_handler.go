package httpdelivery

import "net/http"

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
