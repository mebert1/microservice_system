package kpi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// rest handler for simple get requests that returns only the last kpi per factory
func (s *Service) getKpi(w http.ResponseWriter, r *http.Request) {
	location := chi.URLParam(r, "location")
	s.Logger.Infow("Received request to fetch kpi", "location", location)

	kpis, err := s.Storage.FindKPI(location)
	if err != nil {
		s.handleAPIError("Failed to fetch kpis", err, w)
		return
	}

	body, err := json.Marshal(kpis)
	if err != nil {
		s.handleAPIError("Failed to marshal response", err, w)
		return
	}

	w.Write(body)
}

// returns a list of the last n kpis for each factory
func (s *Service) getKpis(w http.ResponseWriter, r *http.Request) {
	location := chi.URLParam(r, "location")
	n, err := strconv.Atoi(chi.URLParam(r, "n"))
	if err != nil {
		s.handleAPIError("Failed to parse to int", err, w)
		return
	}

	s.Logger.Infow("Received request to fetch kpi", "location", location, "amount", n)

	kpis, err := s.Storage.FindLastNKPI(location, int64(n))
	if err != nil {
		s.handleAPIError("Failed to fetch kpis", err, w)
		return
	}

	body, err := json.Marshal(kpis)
	if err != nil {
		s.handleAPIError("Failed to marshal response", err, w)
		return
	}

	w.Write(body)
}

func (s *Service) handleAPIError(msg string, err error, w http.ResponseWriter) {
	s.Logger.Errorw("msg", "err", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}
