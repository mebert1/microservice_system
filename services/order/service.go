package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

const (
	customerServiceURL = "http://customer-service:8080"
	modelServiceURL    = "http://model-service:8080"
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
	orderService := &Service{}
	orderService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	// add additional producers to send messages to the factories
	for _, location := range additionalProducers {
		producer, err := orderService.Service.RbmqSession.NewProducer(location, config.Rbmq.ExchangeType)
		if err != nil {
			return nil, err
		}

		orderService.Producer[location] = producer
	}

	// add additional producers to send messages to the factories
	err = orderService.InitStorage()
	if err != nil {
		return nil, err
	}

	// launch a new thread to handle incoming rabbitmq messages
	go orderService.handleRbmqMessage(messages)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	// initialize a chi router and its handler functions
	router.Post("/", orderService.postOrder)
	router.Get("/", orderService.getAllOrders)
	router.Get("/{id}", orderService.getOrder)

	go orderService.InitAPI(router)

	return orderService, nil
}

// handleRbmqMessage handles incoming messages, this service only expects order updates to be send to it by rabbitmq
func (s *Service) handleRbmqMessage(messages <-chan rbmq.Message) {
	// infinite loop iterating over an unbuffered channel that blocks until a new message is received
	for msg := range messages {
		orderMsg := rbmq.OrderMessage{}

		// decode the message
		err := json.Unmarshal(msg.Body, &orderMsg)
		if err != nil {
			s.Logger.Errorw("Failed to parse message", "err", err, "msg", string(msg.Body))
			continue
		}

		// reject if wrong message type
		if orderMsg.MsgType != "orderupdate" {
			s.Logger.Errorw("Unhandled message type", "type", orderMsg.MsgType)
			continue
		}

		s.Logger.Infow("Order update received", "order", orderMsg.OrderID, "status", orderMsg.Status)

		// update the order status
		s.updateOrder(orderMsg)
	}
}

// prepareOrder prepares and creates an order based on a http request body
func (s *Service) prepareOrder(body []byte) ([]byte, error) {
	// initialize the entity
	order := entities.Order{
		Created: time.Now().UTC(),
		Status:  "processing",
	}

	// decode the body into the entity
	err := json.Unmarshal(body, &order)
	if err != nil {
		return nil, err
	}

	// check if the customer exists
	if !s.customerExists(order) {
		return nil, errors.New("Invalid customer: " + order.Customer)
	}

	s.Logger.Info("Received request to create new order", "customer", order.Customer)

	// create a new database entry
	order.ObjectID, err = s.Storage.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	// fetch model and part ids
	items, err := s.fetchModelAndParts(order.Items)
	if err != nil {
		return nil, err
	}

	// prepare a message for the delegation service
	orderMsg := rbmq.OrderMessage{
		Timestamp: time.Now().UTC(),
		OrderID:   order.ObjectID,
		Customer:  order.Customer,
		MsgType:   "delegate",
		Items:     items,
	}

	s.Logger.Infow("Created new order", "customer", order.Customer, "order", order.ObjectID)

	// delegate the order
	err = s.delegateOrder(orderMsg)
	if err != nil {
		return nil, err
	}

	// return a response for the http request
	responseBody, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	return responseBody, err
}

// delegateOrder forwards an order to the delegation service
func (s *Service) delegateOrder(orderMsg rbmq.OrderMessage) error {

	// encode the message
	msgBody, err := json.Marshal(orderMsg)
	if err != nil {
		return err
	}

	// publish the message to ther service
	err = s.Producer["london"].Publish(msgBody, "delegation")
	if err != nil {
		return err
	}

	s.Logger.Infow("Forwarded order to delegation service", "order", orderMsg.OrderID)

	return nil
}

// updateOrder updates the status of a single order
func (s *Service) updateOrder(msg rbmq.OrderMessage) {
	// initialize an entity and fill it with the updated information
	order := entities.Order{
		ObjectID:   msg.OrderID,
		LastUpdate: time.Now().UTC(),
		Status:     msg.Status,
	}

	// write the updates to the database
	err := s.Storage.UpdateOrderStatus(order)
	if err != nil {
		s.Logger.Errorw("Failed to update order", "id", msg.OrderID, "err", err)
	}
}

// customerExists sends a http request to the customer service's rest api to check if a customer exists
func (s *Service) customerExists(order entities.Order) bool {
	// send a get request to the service
	resp, err := http.Get(fmt.Sprintf("%s/%s", customerServiceURL, order.Customer))
	if err != nil {
		s.Logger.Errorw("Failed to fetch customer", "err", err)
		return false
	}

	// read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Errorw("Failed to read response body", "err", err)
		return false
	}

	var customer entities.Customer

	// decode the response
	err = json.Unmarshal(body, &customer)
	if err != nil {
		s.Logger.Errorw("Failed to parse body to json", "err", err)
		return false
	}

	// return true if the customer exists
	return customer.ObjectID != ""
}

// dummy function to simulate the model service
func (s *Service) fetchModelAndParts(items []int) ([]rbmq.Item, error) {
	var rbmqItems []rbmq.Item
	for _, item := range items {
		resp, err := http.Get(fmt.Sprintf("%s/%v", modelServiceURL, item))
		if err != nil {
			return rbmqItems, err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return rbmqItems, err
		}

		model := entities.Model{}
		json.Unmarshal(body, &model)

		var parts []int

		for _, part := range model.Parts {
			parts = append(parts, part.ID)
		}

		rbmqItem := rbmq.Item{
			ItemID:       model.ID,
			Parts:        parts,
			AssemblyTime: model.AssemblyTime,
		}

		rbmqItems = append(rbmqItems, rbmqItem)
	}
	return rbmqItems, nil
}
