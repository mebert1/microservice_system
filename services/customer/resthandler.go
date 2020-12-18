package customer

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
)

// postCustomer is the handler function used to create a new customer
func (s *Service) postCustomer(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleAPIError("Failed to read request body", err, w)
		return
	}

	response, err := s.prepareCustomer(body)
	if err != nil {
		s.handleAPIError("Failed to create customer", err, w)
		return
	}

	w.Write(response)
}

// getCustomer is the handler function used to get a single customer by id
func (s *Service) getCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.Logger.Infow("Received request to fetch customer", "customer", id)

	customer, err := s.Storage.FindCustomer(id)
	if err != nil {
		s.handleAPIError("Failed to find customer", err, w)
		return
	}

	body, err := json.Marshal(customer)
	if err != nil {
		s.handleAPIError("Failed to marshal response", err, w)
		return
	}

	w.Write(body)
}

// getAllCustomers is the handler function to get a list of all customers
func (s *Service) getAllCustomers(w http.ResponseWriter, r *http.Request) {
	s.Logger.Info("Received request to fetch all customers")

	customers, err := s.Storage.AllCustomers()
	if err != nil {
		s.handleAPIError("Failed to fetch customers", err, w)
		return
	}

	body, err := json.Marshal(customers)
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
