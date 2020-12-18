package customer

import (
	"encoding/json"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// Service uses composition to expand the service library
type Service struct {
	*service.Service
}

// New launches a new custom service based on the service library in /pkg/service
func New(config *service.Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	var err error

	// initialize a new service instance based on the config
	customerService := &Service{}
	customerService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	// close the rabbitmq session because the customer service is not using the rabbitmq server for communication
	customerService.RbmqSession.Close()

	// initialize the database
	err = customerService.InitStorage()
	if err != nil {
		return nil, err
	}

	// initialize a chi router and its handler functions
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	router.Post("/", customerService.postCustomer)
	router.Get("/", customerService.getAllCustomers)
	router.Get("/{id}", customerService.getCustomer)

	// launch the api router in a new thread
	go customerService.InitAPI(router)

	return customerService, nil
}

// prepareCustomer takes a http request body and parses it so that it can be stored in the database
func (s *Service) prepareCustomer(body []byte) ([]byte, error) {
	s.Logger.Info("Received request to create customer")

	customer := entities.Customer{
		Created: time.Now().UTC(),
	}

	// decode the body
	err := json.Unmarshal(body, &customer)
	if err != nil {
		return nil, err
	}

	// add customer to the database
	customer.ObjectID, err = s.Storage.CreateCustomer(customer)
	if err != nil {
		return nil, err
	}

	s.Logger.Infow("Successfully created customer", "customer", customer.ObjectID)

	// encode the http response body
	responseBody, err := json.Marshal(customer)
	if err != nil {
		return nil, err
	}

	return responseBody, err
}
