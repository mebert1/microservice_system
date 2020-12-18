package delegation

import (
	"encoding/json"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/service"
	"go.uber.org/zap"
)

// Service uses composition to expand the service library
type Service struct {
	*service.Service
}

var additionalProducers = [...]string{"china", "usa"}

// New launches a new custom service based on the service library in /pkg/service
func New(config *service.Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	var err error

	// initialize a new service instance based on the config
	delegationService := &Service{}
	delegationService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	// add additional producers to send messages to the factories
	for _, location := range additionalProducers {
		producer, err := delegationService.Service.RbmqSession.NewProducer(location, config.Rbmq.ExchangeType)
		if err != nil {
			return nil, err
		}

		delegationService.Producer[location] = producer
	}

	// initialize the database
	err = delegationService.InitStorage()
	if err != nil {
		return nil, err
	}

	logger.Info("Initializing database")
	err = delegationService.Storage.InitDelegationDatabase()
	if err != nil {
		return nil, err
	}

	// launch a new thread to handle incoming rabbitmq messages
	go delegationService.handleRbmqMessage(messages)

	return delegationService, nil
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

		// read the message type to decide whether its a new order that needs to be delegated
		// or an existing one that has been updated and thus decreasing the load
		if orderMsg.MsgType == "delegate" {
			err = s.delegateOrder(orderMsg)
			if err != nil {
				s.Logger.Errorw("Failed to delegate message", "err", err)
			}
		} else if orderMsg.MsgType == "orderupdate" {
			err = s.updateFactoryStatus(orderMsg)
			if err != nil {
				s.Logger.Errorw("Failed to update factory status", "location", orderMsg.Location, "err", err)
			}
		} else {
			s.Logger.Debugw("Unhandled message type", "type", orderMsg.MsgType)
		}
	}
}

// delegateOrder determines the location a new order is being sent to
func (s *Service) delegateOrder(orderMsg rbmq.OrderMessage) error {
	targetLocation := ""
	var relativeLoadUSA float32 = 0
	var relativeLoadChina float32 = 0

	s.Logger.Infow("Received order", "order", orderMsg.OrderID)

	// get the current status of each factory
	status, err := s.getFactoryStatus()
	if err != nil {
		return err
	}

	// calculate the relative load of each factory
	relativeLoadUSA = float32(status["usa"].CurrentLoad) / float32(status["usa"].MaxConcurrentOrders)
	relativeLoadChina = float32(status["china"].CurrentLoad) / float32(status["china"].MaxConcurrentOrders)

	// send the order to the factory with less load, or the one with the higher capacity if both have the same relative load
	if relativeLoadUSA == relativeLoadChina {
		if status["usa"].MaxConcurrentOrders > status["china"].MaxConcurrentOrders {
			targetLocation = "usa"
		} else {
			targetLocation = "china"
		}
	} else if relativeLoadUSA > relativeLoadChina {
		targetLocation = "china"
	} else {
		targetLocation = "usa"
	}

	s.Logger.Infow("Calculating relative load", "china", relativeLoadChina, "usa", relativeLoadUSA)

	return s.delegateTo(status[targetLocation], orderMsg)
}

// getFactoryStatus accumulates the status of each factory
func (s *Service) getFactoryStatus() (map[string]entities.FactoryStatus, error) {
	var err error
	statusMap := make(map[string]entities.FactoryStatus)

	statusMap["usa"], err = s.Storage.GetFactoryStatus("usa")
	if err != nil {
		return statusMap, err
	}

	statusMap["china"], err = s.Storage.GetFactoryStatus("china")
	if err != nil {
		return statusMap, err
	}

	return statusMap, err
}

// delegateTo forwards an order to a specific location
func (s *Service) delegateTo(status entities.FactoryStatus, orderMsg rbmq.OrderMessage) error {
	// update the messages timestamp and status
	orderMsg.Timestamp = time.Now().UTC()
	orderMsg.MsgType = "neworder"

	// encode the message to json
	msgBody, err := json.Marshal(orderMsg)
	if err != nil {
		return err
	}

	// publish the message to the location
	s.Producer[status.Location].Publish(msgBody, "factory")
	if err != nil {
		return err
	}

	s.Logger.Infow("Delegating order to factory", "order", orderMsg.OrderID, "location", status.Location)

	// update the current load of the location
	status.CurrentLoad = status.CurrentLoad + 1

	return s.Storage.UpdateFactoryStatus(status)
}

// updateFactoryStatus changes the status of a factory
// It is usually called when a factory completed an order and thus decreases its load
func (s *Service) updateFactoryStatus(orderMsg rbmq.OrderMessage) error {
	s.Logger.Infow("Received order update", "order", orderMsg.OrderID)

	status, err := s.Storage.GetFactoryStatus(orderMsg.Location)
	if err != nil {
		return err
	}

	status.CurrentLoad = status.CurrentLoad - 1

	s.Logger.Infow("Updating load", "location", orderMsg.Location, "load", status.CurrentLoad)

	return s.Storage.UpdateFactoryStatus(status)
}
