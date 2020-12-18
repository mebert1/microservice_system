package part

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/service"
	"go.uber.org/zap"
)

// Service is the instance wrapper
type Service struct {
	*service.Service
}

// New launches a new custom service based on the service library in /pkg/service
func New(config *service.Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	var err error

	// initialize a new service instance based on the config
	partsService := &Service{}
	partsService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	// initialize the database
	err = partsService.InitStorage()
	if err != nil {
		return nil, err
	}

	logger.Info("Initializing database")
	err = partsService.Storage.InitPartDatabase()
	if err != nil {
		return nil, err
	}

	// launch a new thread to handle incoming rabbitmq messages
	go partsService.handleRbmqMessage(messages)

	return partsService, nil
}

// handleRbmqMessage handles incoming messages
func (s *Service) handleRbmqMessage(messages <-chan rbmq.Message) {

	// this is a placeholder struct used to determine what to do with the message,
	// because the part service can receive two potential message objects
	type genericMessage struct {
		MsgType string `json:"type,omitempty"`
	}

	// infinite loop iterating over an unbuffered channel that blocks until a new message is received
	for msg := range messages {
		recMsg := genericMessage{}

		// decode the msg body
		err := json.Unmarshal(msg.Body, &recMsg)
		if err != nil {
			s.Logger.Errorw("Failed to parse message", "err", err, "msg", string(msg.Body))
			continue
		}

		switch recMsg.MsgType {
		// Message originates from model service with command to update price of certain part
		case "updatepart":
			err = s.handlePartUpdate(msg.Body)
			if err != nil {
				s.Logger.Errorw("Failed to update part", "err", err)
				continue
			}

		// Message originates from factory service with command to order parts
		case "orderpart":
			// Order all parts of all items within received order
			err = s.handlePartOrder(msg.Body)
			if err != nil {
				s.Logger.Errorw("Failed to order part", "err", err)
				continue
			}
			// returns acknowledgement message when all parts are delivered
		default:
			s.Logger.Debugw("Unhandled message type", "type", recMsg.MsgType)
		}
	}
}

func (s *Service) handlePartOrder(msg []byte) error {
	orderMsg := rbmq.OrderMessage{}
	err := json.Unmarshal(msg, &orderMsg)
	if err != nil {
		return err
	}
	return s.processOrder(orderMsg)
}

func (s *Service) handlePartUpdate(msg []byte) error {
	partMsg := rbmq.PartMessage{}

	err := json.Unmarshal(msg, &partMsg)
	if err != nil {
		return err
	}

	part := entities.Part{
		ID:    partMsg.Part,
		Price: partMsg.Price,
	}

	s.Logger.Infow("Received part update", "part", part.ID)

	err = s.Storage.UpdatePart(part)
	if err != nil {
		return fmt.Errorf("Failed to update part %v", part.ID)
	}

	return nil
}

func (s *Service) processOrder(order rbmq.OrderMessage) error {
	var combinedPrice int

	s.Logger.Infow("Received part order", "order", order.OrderID)

	for _, item := range order.Items {
		s.Logger.Infow("Ordering parts", "order", order.OrderID, "item", item.ItemID)
		for _, part := range item.Parts {

			dbPart, err := s.Storage.FindPart(part)
			if err != nil {
				return err
			}

			// sleep to simulate part delivery process
			orderPart()

			// Count up combined price of entire delivery
			combinedPrice = combinedPrice + dbPart.Price
		}
	}

	order.CostsOfParts = combinedPrice
	order.Timestamp = time.Now().UTC()
	order.MsgType = "orderupdate"
	order.Status = "partsdelivered"

	// marshal message struct back into rbmq message ([]byte)
	responseBody, err := json.Marshal(order)
	if err != nil {
		return err
	}

	s.Producer[s.Config.Location].Publish(responseBody, "factory")
	s.Logger.Infow("Ordered and received all parts", "order", order.OrderID, "price", combinedPrice)
	return nil
}

func orderPart() {
	waitTime := rand.Intn(500) + 500
	time.Sleep(time.Duration(waitTime) * time.Millisecond)
}
