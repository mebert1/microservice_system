package kpi

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

const requestInterval = 90

var additionalProducers = [...]string{"china", "usa"}

// Service uses composition to expand the service library
type Service struct {
	*service.Service
}

// New launches a new custom service based on the service library in /pkg/service
func New(config *service.Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	var err error

	// initialize a new service instance based on the config
	kpiService := &Service{}
	kpiService.Service, err = service.New(config, messages, logger)
	if err != nil {
		return nil, err
	}

	// add additional producers to send messages to the factories
	for _, location := range additionalProducers {
		producer, err := kpiService.Service.RbmqSession.NewProducer(location, config.Rbmq.ExchangeType)
		if err != nil {
			return nil, err
		}

		kpiService.Producer[location] = producer
	}

	// initialize the database
	err = kpiService.InitStorage()
	if err != nil {
		return nil, err
	}

	// launch a new thread to handle incoming rabbitmq messages
	go kpiService.handleRbmqMessage(messages)

	// initialize a chi router and its handler functions
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	router.Get("/", kpiService.getKpi)
	router.Get("/{location}", kpiService.getKpi)
	router.Get("/{location}/{n}", kpiService.getKpis)

	// launch the api router in a new thread
	go kpiService.InitAPI(router)

	// launch a new thread that periodically requests new kpi from the factories
	go kpiService.requestKPIs()

	return kpiService, nil
}

// handleRbmqMessage handles incoming messages
func (s *Service) handleRbmqMessage(messages <-chan rbmq.Message) {
	// infinite loop iterating over an unbuffered channel that blocks until a new message is received
	for msg := range messages {
		var kpiMsg rbmq.KPIMessage

		// decode the msg body
		err := json.Unmarshal(msg.Body, &kpiMsg)
		if err != nil {
			s.Logger.Errorw("Failed to unmarshal message", "err", err)
			continue
		}

		// this service only expects messages to be of type kpiupdate, everything else is rejected
		if kpiMsg.MsgType != "kpiupdate" {
			s.Logger.Errorw("Unhandled message type", "type", kpiMsg.MsgType)
			continue
		}

		s.Logger.Infow("Received kpi update", "location", kpiMsg.Location)

		// create a new kpi entity based on the messsage
		kpi := entities.KPI{
			Created:          time.Now().UTC(),
			Location:         kpiMsg.Location,
			IncompleteOrders: kpiMsg.IncompleteOrders,
			CompletedOrders:  kpiMsg.CompletedOrders,
			Total:            kpiMsg.IncompleteOrders + kpiMsg.CompletedOrders,
			CostsOfParts:     kpiMsg.CostsOfParts,
		}

		// add the entity to the database
		_, err = s.Storage.CreateKPI(kpi)
		if err != nil {
			s.Logger.Errorw("Failed to add kpi entry", "err", err)
		}
	}
}

// requestKPIs that sends kpi requests to each factory
func (s *Service) requestKPIs() {
	for {
		// prepare a new message
		kpiMsg := rbmq.KPIMessage{
			Timestamp: time.Now().UTC(),
			MsgType:   "requestkpi",
		}

		// encode the message
		msg, err := json.Marshal(kpiMsg)
		if err != nil {
			s.Logger.Errorw("Failed to marshal message", "err", err)
			continue
		}

		// publish the message to china
		err = s.Producer["china"].Publish(msg, "factory")
		if err != nil {
			s.Logger.Errorw("Failed to publish message", "location", "china", "err", err)
		}

		// publish the message to usa
		err = s.Producer["usa"].Publish(msg, "factory")
		if err != nil {
			s.Logger.Errorw("Failed to publish message", "location", "usa", "err", err)
		}

		// block until the timer is over
		<-time.After(requestInterval * time.Second)
	}
}
