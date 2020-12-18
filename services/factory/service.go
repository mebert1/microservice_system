package factory

import (
	"encoding/json"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/service"
	"go.uber.org/zap"
)

var additionalProducers = [...]string{"london"}

// Service uses composition to expand the service library
type Service struct {
	*service.Service
}

// New launches a new custom service based on the service library in /pkg/service
func New(config *service.Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	var err error

	// initialize a new service instance based on the config
	factoryService := &Service{}
	factoryService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	// add additional producers to send messages to the headquarter
	for _, location := range additionalProducers {
		producer, err := factoryService.Service.RbmqSession.NewProducer(location, config.Rbmq.ExchangeType)
		if err != nil {
			return nil, err
		}

		factoryService.Producer[location] = producer
	}

	// initialize the database
	err = factoryService.InitStorage()
	if err != nil {
		return nil, err
	}

	// launch a new thread to handle incoming rabbitmq messages
	go factoryService.handleRbmqMessage(messages)

	return factoryService, nil
}

// handleRbmqMessage handles incoming messages
func (s *Service) handleRbmqMessage(messages <-chan rbmq.Message) {
	// infinite loop iterating over an unbuffered channel that blocks until a new message is received
	for msg := range messages {

		// decode the msg body
		orderMsg := rbmq.OrderMessage{}
		err := json.Unmarshal(msg.Body, &orderMsg)
		if err != nil {
			s.Logger.Errorw("Failed to parse message", "err", err, "msg", string(msg.Body))
			continue
		}

		// read the message type to decide whether its a new order, an update to an existing order or a request for new kpi
		switch orderMsg.MsgType {
		case "neworder":
			s.Logger.Infow("Received new order", "order", orderMsg.OrderID)

			// add order to the database and receive the message to be forwarded to the part service
			orderMsg, err := s.insertOrder(orderMsg)

			if err != nil {
				s.Logger.Errorw("Failed to insert into database", "err", err)
			}

			body, err := json.Marshal(orderMsg)
			if err != nil {
				s.Logger.Errorw("Failed to marshal message", "err", err)
			}

			// send a part order to the part service
			err = s.Producer[s.Config.Location].Publish(body, "part")
			if err != nil {
				s.Logger.Errorw("Failed to send message to part service", "err", err)
			}

			orderMsg.Timestamp = time.Now().UTC()
			orderMsg.Status = "production"
			orderMsg.MsgType = "orderupdate"

			body, err = json.Marshal(orderMsg)
			if err != nil {
				s.Logger.Errorw("Failed to marshal message", "err", err)
			}

			err = s.Producer["london"].Publish(body, "order")
			if err != nil {
				s.Logger.Errorw("Failed to send message to order service", "err", err)
			}

		case "orderupdate":
			s.Logger.Infow("Received order update", "order", orderMsg.OrderID, "status", orderMsg.Status)

			targetService := ""
			targetLocation := s.Config.Location

			// check the orders status to update the status in the database accordingly and notify the headquarter if an order is complete
			if orderMsg.Status == "partsdelivered" {
				orderMsg.Timestamp = time.Now().UTC()
				targetService = "assembly"
				s.updateCosts(orderMsg)
			} else if orderMsg.Status == "complete" {
				orderMsg.Timestamp = time.Now().UTC()
				targetService = "shipping"
			} else if orderMsg.Status == "shipped" {
				s.notifyLondon(orderMsg)
			} else {
				s.Logger.Errorw("Unknown order status", "status", orderMsg.Status)
				continue
			}

			// update the database entry for an order
			err := s.updateFactoryOrder(orderMsg)
			if err != nil {
				s.Logger.Errorw("Failed to update order", "id", orderMsg.OrderID, "err", err)
			}

			// encode the message body
			event, err := json.Marshal(orderMsg)
			if err != nil {
				s.Logger.Errorw("Failed to parse message", "err", err, "msg", string(msg.Body))
				continue
			}

			// publish the message to the next service
			err = s.Producer[targetLocation].Publish(event, targetService)
			if err != nil {
				s.Logger.Errorw("Failed to send message to assembly", "err", err)
			}

		case "requestkpi":
			s.Logger.Info("Received kpi request")

			// get new kpis from the kpi request handler function
			kpi, err := s.handleKPIRequest()
			if err != nil {
				s.Logger.Errorw("Failed to fetch kpis", "err", err)
				continue
			}

			// send the new kpis to the headquarter
			err = s.Producer["london"].Publish(kpi, "kpi")
			if err != nil {
				s.Logger.Errorw("Failed to send message to assembly", "err", err)
			}

			s.Logger.Info("Sent aggregated kpis to kpi service")
		default:
			s.Logger.Errorw("Unhandled message type", "type", orderMsg.MsgType)
		}
	}
}

func orderFromMessage(msg rbmq.OrderMessage) entities.Order {
	return entities.Order{
		OrderID:      msg.OrderID,
		LastUpdate:   msg.Timestamp,
		Status:       msg.Status,
		CostsOfParts: msg.CostsOfParts,
	}
}

func (s *Service) updateCosts(orderMsg rbmq.OrderMessage) {
	order := orderFromMessage(orderMsg)
	s.Storage.UpdateOrderCosts(order)
}

// updateFactoryOrder uses an order message to update an order's fields
func (s *Service) updateFactoryOrder(msg rbmq.OrderMessage) error {
	order := orderFromMessage(msg)
	return s.Storage.UpdateOrderStatusFactory(order)
}

// deprecated
func (s *Service) updateStatus(status string, msg rbmq.Message) []byte {
	recMsg := rbmq.OrderMessage{}
	err := json.Unmarshal(msg.Body, &recMsg)
	if err != nil {
		s.Logger.Errorw("Failed to parse message", "err", err, "msg", string(msg.Body))
	}
	recMsg.Status = status
	response, err := json.Marshal(recMsg)

	if err != nil {
		s.Logger.Errorw("Failed to marshal", "err", err)
	}
	return response
}

// insertOrder adds a new order to the factories database
func (s *Service) insertOrder(orderMsg rbmq.OrderMessage) (rbmq.OrderMessage, error) {
	var items []int
	var err error

	// aggregate all items in a list of strings
	for _, item := range orderMsg.Items {
		items = append(items, item.ItemID)
	}

	// create a new entity object
	order := entities.Order{
		Created:  time.Now().UTC(),
		Status:   "waitingForParts",
		Customer: orderMsg.Customer,
		Items:    items,
		OrderID:  orderMsg.OrderID,
	}

	// store the object in the database
	_, err = s.Storage.CreateOrderFactory(order)
	if err != nil {
		return orderMsg, err
	}

	// prepare the message to the part service
	orderMsg.MsgType = "orderpart"
	orderMsg.Timestamp = time.Now().UTC()

	return orderMsg, nil
}

// handleKPIRequest aggregates new kpi entries
func (s *Service) handleKPIRequest() ([]byte, error) {
	// fetch new kpi from the database
	kpis, err := s.Storage.AggregateKPI()
	if err != nil {
		return nil, err
	}

	// parse the kpi to an update message
	msg := rbmq.KPIMessage{
		Timestamp: time.Now().UTC(),
		MsgType:   "kpiupdate",
		Location:  s.Config.Location,
	}

	// fill the fields
	if len(kpis) == 1 {
		kpi := kpis[0]

		msg.IncompleteOrders = kpi.Total - kpi.CompletedOrders
		msg.CompletedOrders = kpi.CompletedOrders
		msg.Total = kpi.Total
		msg.CostsOfParts = kpi.CostsOfParts
	}

	// return the encoded message
	return json.Marshal(msg)
}

// notifyLondon sends an update to london when an order is complete
func (s *Service) notifyLondon(orderMsg rbmq.OrderMessage) {
	orderMsg.Timestamp = time.Now().UTC()
	orderMsg.Status = "complete"
	orderMsg.Location = s.Config.Location

	event, err := json.Marshal(orderMsg)
	if err != nil {
		s.Logger.Errorw("Failed to parse message", "err", err)
	}

	err = s.Producer["london"].Publish(event, "delegation")
	if err != nil {
		s.Logger.Errorw("Failed to send message to assembly", "err", err)
	}

	err = s.Producer["london"].Publish(event, "order")
	if err != nil {
		s.Logger.Errorw("Failed to send message to assembly", "err", err)
	}
}
