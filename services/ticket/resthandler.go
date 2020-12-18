package ticket

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
)

// postTicket handles new tickets
func (s *Service) postTicket(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleAPIError("Failed to read request body", err, w)
		return
	}

	response, err := s.prepareTicket(body)
	if err != nil {
		s.handleAPIError("Failed to create order", err, w)
		return
	}

	w.Write(response)
}

// getTicket returns a ticket by its id
func (s *Service) getTicket(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.Logger.Infow("Received request to fetch ticket", "ticket", id)

	ticket, err := s.Storage.FindTicket(id)
	if err != nil {
		s.handleAPIError("Failed to find ticket", err, w)
		return
	}

	body, err := json.Marshal(ticket)
	if err != nil {
		s.handleAPIError("Failed to marshal response", err, w)
		return
	}

	w.Write(body)
}

// getAllTickets returns a list of all tickets
func (s *Service) getAllTickets(w http.ResponseWriter, r *http.Request) {
	s.Logger.Info("Received request to fetch all tickets")

	tickets, err := s.Storage.AllTickets()
	if err != nil {
		s.handleAPIError("Failed to fetch tickets", err, w)
		return
	}

	body, err := json.Marshal(tickets)
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
