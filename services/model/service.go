package model

import (
	"encoding/json"
	"time"

	"net/http"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/service"
	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var additionalProducers = [...]string{"china", "usa"}

// Service uses composition to expand the service library
type Service struct {
	*service.Service
}

// New launches a new custom service based on the service library in /pkg/service
func New(config *service.Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	var err error

	// initialize a new service instance based on the config
	modelService := &Service{}
	modelService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	// add additional producers to send messages to the factories
	for _, location := range additionalProducers {
		producer, err := modelService.Service.RbmqSession.NewProducer(location, config.Rbmq.ExchangeType)
		if err != nil {
			return nil, err
		}

		modelService.Producer[location] = producer
	}

	// initialize the database
	err = modelService.InitStorage()
	if err != nil {
		return nil, err
	}

	logger.Info("Initializing database")
	err = modelService.Storage.InitModelDatabase()
	if err != nil {
		return nil, err
	}

	// initialize a chi router and its handler functions
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	router.Post("/", modelService.updatePrice)
	router.Get("/{id}", modelService.getModel)
	router.Get("/", modelService.getAllModels)

	go modelService.InitAPI(router)

	return modelService, nil
}

// updatePriceDB updates the price of a single part in the database and terminates the http request
func (s *Service) updatePriceDB(w http.ResponseWriter, part entities.Part) ([]byte, error) {

	err := s.Storage.UpdateModelPart(part)
	if err != nil {
		s.Logger.Errorw("Failed to update part price", "err", err)
		return nil, err
	}

	response, err := json.Marshal(part)
	if err != nil {
		s.Logger.Errorw("Failed to create response message", "err", err)
		return response, err
	}

	return response, err
}

// notifyPartService sends a message to each factory to update their local pricing services
func (s *Service) notifyPartService(part entities.Part) error {

	partMsg := &rbmq.PartMessage{
		Timestamp: time.Now().UTC(),
		MsgType:   "updatepart",
		Part:      part.ID,
		Price:     part.Price,
	}

	notification, err := json.Marshal(partMsg)
	if err != nil {
		return err
	}

	err = s.Producer["usa"].Publish(notification, "part")
	if err != nil {
		return err
	}
	err = s.Producer["china"].Publish(notification, "part")
	if err != nil {
		return err
	}
	s.Logger.Infow("Sent part updates to factories", "part", part.ID)
	return nil
}
