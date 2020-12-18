package order

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
)

// postOrder is the rest handler to create a new order
func (s *Service) postOrder(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleAPIError("Failed to read request body", err, w)
		return
	}

	response, err := s.prepareOrder(body)
	if err != nil {
		s.handleAPIError("Failed to create order", err, w)
		return
	}

	w.Write(response)
}

// getOrder is the rest handler to return a single order by its id
func (s *Service) getOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.Logger.Info("Received request to fetch order", "order", id)

	order, err := s.Storage.FindOrder(id)
	if err != nil {
		s.handleAPIError("Failed to find order", err, w)
		return
	}

	body, err := json.Marshal(order)
	if err != nil {
		s.handleAPIError("Failed to marshal response", err, w)
		return
	}

	w.Write(body)
}

// getAllOrders is the rest handler to return a list of all orders
func (s *Service) getAllOrders(w http.ResponseWriter, r *http.Request) {
	s.Logger.Info("Received request to fetch all orders")

	orders, err := s.Storage.AllOrders()
	if err != nil {
		s.handleAPIError("Failed to fetch orders", err, w)
		return
	}

	body, err := json.Marshal(orders)
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
