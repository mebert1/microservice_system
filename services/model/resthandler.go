package model

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
)

// updatePrice updates the pricing based on the body of a post request
func (s *Service) updatePrice(w http.ResponseWriter, r *http.Request) {
	var part entities.Part
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleAPIError("Failed to read request body", err, w)
		return
	}

	err = json.Unmarshal(body, &part)
	if err != nil {
		s.handleAPIError("Failed to parse message", err, w)
		return
	}

	s.Logger.Infow("Received request to update part", "part", part.ID)

	response, err := s.updatePriceDB(w, part)
	if err != nil {
		s.handleAPIError("Failed to parse message", err, w)
		return
	}

	err = s.notifyPartService(part)
	if err != nil {
		s.handleAPIError("Failed to parse message", err, w)
		return
	}

	w.Write(response)
}

// getModel is the rest handler to return a model by its id
func (s *Service) getModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.Logger.Infow("Received request to fetch model", "model", id)

	modelID, err := strconv.Atoi(id)
	if err != nil {
		s.handleAPIError("Failed to find order", err, w)
		return
	}

	model, err := s.Storage.FindModel(modelID)
	if err != nil {
		s.handleAPIError("Failed to find order", err, w)
		return
	}

	body, err := json.Marshal(model)
	if err != nil {
		s.handleAPIError("Failed to marshal response", err, w)
		return
	}

	w.Write(body)
}

// getAllModels returns a list of all models in the database
func (s *Service) getAllModels(w http.ResponseWriter, r *http.Request) {
	s.Logger.Info("Received request to fetch all models")

	models, err := s.Storage.AllModels()
	if err != nil {
		s.handleAPIError("Failed to fetch orders", err, w)
		return
	}

	body, err := json.Marshal(models)
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
