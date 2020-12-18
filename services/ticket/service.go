package ticket

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

var additionalProducers = [...]string{"india", "mexico"}

// New launches a new custom service based on the service library in /pkg/service
func New(config *service.Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	var err error

	// initialize a new service instance based on the config
	ticketService := &Service{}
	ticketService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	// add additional producers to send messages to the factories
	for _, location := range additionalProducers {
		producer, err := ticketService.Service.RbmqSession.NewProducer(location, config.Rbmq.ExchangeType)
		if err != nil {
			return nil, err
		}

		ticketService.Producer[location] = producer
	}

	// initialize the database
	err = ticketService.InitStorage()
	if err != nil {
		return nil, err
	}

	// launch a new thread to handle incoming rabbitmq messages
	go ticketService.handleRbmqMessage(messages)

	// initialize a chi router and its handler functions
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	router.Post("/", ticketService.postTicket)
	router.Get("/", ticketService.getAllTickets)
	router.Get("/{id}", ticketService.getTicket)

	go ticketService.InitAPI(router)

	return ticketService, nil
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

		// reject if the message type isn't resolve
		if ticketMsg.MsgType != "resolve" {
			s.Logger.Debugw("Unhandled message type", "type", ticketMsg.MsgType)
			continue
		}

		s.Logger.Infow("Received ticket update", "ticket", ticketMsg.TicketID)

		// check if ticket is open
		if isOpen, err := s.ticketIsOpen(ticketMsg.TicketID); !isOpen {
			if err != nil {
				s.Logger.Errorw("Cannot update ticket", "err", err)
				continue
			}
			s.Logger.Errorw("Cannot update ticket, ticket is already closed", "id", ticketMsg.TicketID)
			continue
		}

		// prepare a new ticket object
		ticket := entities.Ticket{
			ObjectID: ticketMsg.TicketID,
			Status:   "closed",
			Closed:   time.Now().UTC(),
			Response: ticketMsg.Response,
		}

		// update the ticket
		err = s.Storage.UpdateTicket(ticket)
		if err != nil {
			s.Logger.Errorw("Failed to update ticket", "err", err)
		}
	}
}

// prepareTicket creates and stores a new ticket
func (s *Service) prepareTicket(body []byte) ([]byte, error) {
	s.Logger.Info("Received request to create new ticket")
	// determine the target support location
	location, err := s.currentSupportLocation()
	if err != nil {
		return nil, err
	}

	ticket := entities.Ticket{
		Created:  time.Now().UTC(),
		Status:   "open",
		Location: location,
	}

	// decode the body
	err = json.Unmarshal(body, &ticket)
	if err != nil {
		return nil, err
	}

	// create a new entry in the database
	ticket.ObjectID, err = s.Storage.CreateTicket(ticket)
	if err != nil {
		return nil, err
	}

	s.Logger.Infow("Created new ticket", "ticket", ticket.ObjectID)

	err = s.forwardTicket(ticket)
	if err != nil {
		return nil, err
	}

	s.Logger.Infow("Forwarded ticket to support center", "ticket", ticket.ObjectID, "location", ticket.Location)

	// encode the response
	responseBody, err := json.Marshal(ticket)
	if err != nil {
		return nil, err
	}

	return responseBody, err
}

// forwardTicket sends the ticket to the support location
func (s *Service) forwardTicket(ticket entities.Ticket) error {
	msg := rbmq.TicketMessage{
		Timestamp: time.Now().UTC(),
		TicketID:  ticket.ObjectID,
	}

	msgBody, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return s.Producer[ticket.Location].Publish(msgBody, "support")
}

// currentSupportLocation determines the target location based on the current time of day in india
func (s *Service) currentSupportLocation() (string, error) {
	// declare time window 6am to 8pm ist
	limitLower, err := time.Parse(time.Kitchen, "12:30AM")
	if err != nil {
		return "", err
	}

	limitUpper, err := time.Parse(time.Kitchen, "02:30PM")
	if err != nil {
		return "", err
	}

	// get current time in india
	currentTime := time.Now().UTC()

	// check if the current time is within the time window
	if currentTime.After(limitLower) && currentTime.Before(limitUpper) {
		return "india", nil
	}

	return "mexico", nil
}

// ticketIsOpen returns true if a ticket is open
func (s *Service) ticketIsOpen(id string) (bool, error) {
	ticket, err := s.Storage.FindTicket(id)
	if err != nil {
		return false, err
	}

	return ticket.Status == "open", nil
}
