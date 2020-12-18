package shipping

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	//"net/http"
	//"io/ioutil"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/service"

	"go.uber.org/zap"
)

const (
	customerServiceURL = "http://customer-service:8080"
	orderServiceURL    = "http://order-service:8080"
)

// Service is the instance wrapper
type Service struct {
	*service.Service
}

// New launches a new custom service based on the service library in /pkg/service
func New(config *service.Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	var err error

	shippingService := &Service{}
	shippingService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	go shippingService.handleRbmqMessage(messages)

	return shippingService, nil
}

/* handleRbmqMessage receives assembled order from factory service */
/* uses REST interface to get customer information and sends ack msg to factory service afetr shipping */
func (s *Service) handleRbmqMessage(messages <-chan rbmq.Message) {
	for msg := range messages {

		// unmarshal rmbq message ([]byte) into struct
		recMsg := rbmq.OrderMessage{}
		err := json.Unmarshal(msg.Body, &recMsg)
		if err != nil {
			s.Logger.Errorw("Failed to parse message", "err", err)
			continue
		}

		customerID := recMsg.Customer

		if recMsg.Customer == "" {
			// HTTP GET request for order to find customer connected to order
			order, err := getOrder(recMsg.OrderID)
			if err != nil {
				s.Logger.Errorw("Failed to get order via HTTP GET", "err", err, "orderID", recMsg.OrderID)
				continue
			}
			customerID = order.Customer
		}

		// HTTP GET request for customer to get shipping address/information
		customer, err := getCustomer(customerID)
		if err != nil {
			s.Logger.Errorw("Failed to get customer via HTTP GET", "err", err)
			continue
		}

		s.Logger.Infow("Received shipping request", "order", recMsg.OrderID)

		// Notify order service that order has been shipped
		response, err := shipmentAck(recMsg, customer)
		if err != nil {
			s.Logger.Errorw("Failed to ship message", "err", err)
			return
		}
		s.Producer[s.Config.Location].Publish(response, "factory")

		s.Logger.Infow("Order shipped", "order", recMsg.OrderID, "customer", customer.ObjectID)
	}
}

func getOrder(id string) (entities.Order, error) {
	var order entities.Order
	var err error

	// HTTP Rest Get request to get customer from customer service
	httpResponse, err := http.Get(fmt.Sprintf("%s/%s", orderServiceURL, id))
	if err != nil {
		return order, err
	}

	byteResponse, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return order, err
	}
	// unmarshal []byte into entities.Customer struct
	err = json.Unmarshal(byteResponse, &order)

	return order, err
}

func getCustomer(id string) (entities.Customer, error) {
	var customer entities.Customer
	var err error

	// HTTP Rest Get request to get customer from customer service
	httpResponse, err := http.Get(fmt.Sprintf("%s/%s", customerServiceURL, id))
	if err != nil {
		return customer, err
	}

	byteResponse, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return customer, err
	}
	// unmarshal []byte into entities.Customer struct
	err = json.Unmarshal(byteResponse, &customer)

	return customer, err
}

func shipmentAck(recMsg rbmq.OrderMessage, customer entities.Customer) ([]byte, error) {

	response := rbmq.OrderMessage{
		Timestamp: time.Now().UTC(),
		MsgType:   "orderupdate",
		Status:    "shipped",
		OrderID:   recMsg.OrderID,
	}

	// marshal message struct back into rbmq message ([]byte)
	responseBody, err := json.Marshal(response)

	return responseBody, err
}
