package support

import (
	"encoding/json"
	"time"

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
	supportService := &Service{}
	supportService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	// add additional producers to send messages to the headquarter
	for _, location := range additionalProducers {
		producer, err := supportService.Service.RbmqSession.NewProducer(location, config.Rbmq.ExchangeType)
		if err != nil {
			return nil, err
		}

		supportService.Producer[location] = producer
	}

	// launch a new thread to handle incoming rabbitmq messages
	go supportService.handleRbmqMessage(messages)

	return supportService, nil
}

// handleRbmqMessage handles incoming messages
func (s *Service) handleRbmqMessage(messages <-chan rbmq.Message) {
	// infinite loop iterating over an unbuffered channel that blocks until a new message is received
	for msg := range messages {
		ticketMsg := rbmq.TicketMessage{}

		// decode the msg body
		err := json.Unmarshal(msg.Body, &ticketMsg)
		if err != nil {
			s.Logger.Errorw("Failed to parse message", "err", err, "msg", string(msg.Body))
			continue
		}

		s.Logger.Infow("Received ticket", "ticket", ticketMsg.TicketID)

		// resolve the ticket
		err = s.resolve(ticketMsg.TicketID)
		if err != nil {
			s.Logger.Errorw("Failed to resolve ticket", "err", err)
		}

		s.Logger.Infow("Resolved ticket", "ticket", ticketMsg.TicketID)

	}
}

// resolve responds to a new ticket with a dummy response text
func (s *Service) resolve(id string) error {
	time.Sleep(5 * time.Second)

	// initialize the response message
	msg := rbmq.TicketMessage{
		Timestamp: time.Now().UTC(),
		MsgType:   "resolve",
		TicketID:  id,
		Response: `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras vulputate cursus tincidunt. Integer tincidunt purus metus, vel finibus ex commodo sed
		Quisque ornare dignissim sapien euismod malesuada. Cras ullamcorper mattis tempor. Pellentesque diam odio, posuere sed efficitur pharetra, sagittis sollicitudin
		neque. Quisque vitae placerat urna, non porttitor dui. Suspendisse egestas nunc diam, et consequat diam interdum ut. Etiam turpis nunc, varius lobortis porttitor
		in, blandit ac tortor.`,
	}

	// encode the message
	msgBody, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// publish the message
	return s.Producer["london"].Publish(msgBody, "ticket")
}
