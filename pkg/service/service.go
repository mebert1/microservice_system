package service

import (
	"context"
	"fmt"
	"net/http"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/db"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Service  contains all information required for a service to function
type Service struct {
	Config *Config

	Storage db.Client
	Api     *http.Server

	RbmqSession *rbmq.Session
	Consumer    *rbmq.Consumer
	Producer    map[string]*rbmq.Producer

	Logger *zap.SugaredLogger
}

// Config wraps the database and rabbitmq configuration structs together
type Config struct {
	Location string
	Db       db.Config
	Rbmq     rbmq.Config
}

// New initializes the service and all rabbitmq components required for it to function
func New(config *Config, messages chan rbmq.Message, logger *zap.SugaredLogger) (*Service, error) {
	rbmqSession, err := rbmq.NewSession(config.Rbmq, logger)
	if err != nil {
		return nil, err
	}

	consumer, err := rbmqSession.NewConsumer(messages)
	if err != nil {
		return nil, err
	}

	producers := make(map[string]*rbmq.Producer)

	defaultProducer, err := rbmqSession.NewProducer(config.Location, config.Rbmq.ExchangeType)
	if err != nil {
		return nil, err
	}

	producers[config.Location] = defaultProducer

	return &Service{
		Config:      config,
		RbmqSession: rbmqSession,
		Consumer:    consumer,
		Producer:    producers,
		Logger:      logger,
	}, nil
}

// InitStorage connects to the database specified in the config struct
func (s *Service) InitStorage() error {
	storage, err := db.New(s.Config.Db)
	if err != nil {
		return err
	}

	s.Storage = storage
	return nil
}

// InitAPI starts a http server on port 8080 that uses a chi router for routing
func (s *Service) InitAPI(router *chi.Mux) {
	s.Api = &http.Server{Addr: ":8080", Handler: router}
	err := s.Api.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		s.Logger.Fatalw("Failed to start API servier", "err", err)
	}
}

// Shutdown closes all connections and sessions
func (s *Service) Shutdown(ctx context.Context) []error {
	var err error
	var errs []error

	if s.Api != nil {
		err = s.Api.Shutdown(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("Failed to shut down api server: %e", err))
		}
	}

	err = s.RbmqSession.Close()
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed to close RabbitMQ connection: %e", err))
	}

	if s.Storage != nil {
		err = s.Storage.Close()
		if err != nil {
			errs = append(errs, fmt.Errorf("Failed to shut down database client: %e", err))
		}
	}

	return errs
}
