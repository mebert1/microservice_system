package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/db"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/rbmq"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/service"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/assembly"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/customer"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/delegation"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/factory"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/kpi"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/model"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/order"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/part"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/shipping"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/support"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/services/ticket"
	"go.uber.org/zap"
)

const (
	customerService   = "customer"
	orderService      = "order"
	delegationService = "delegation"
	modelService      = "model"
	kpiService        = "kpi"

	factoryService = "factory"
	partService    = "part"

	assemblyService = "assembly"
	shippingService = "shipping"

	ticketService  = "ticket"
	supportService = "support"
)

// Service is an interface used as a contract with the service library
type Service interface {
	Shutdown(ctx context.Context) []error
}

// getConfig fills the config structs with data read from environment variables
func getConfig() *service.Config {
	return &service.Config{
		Location: os.Getenv("SERVICE_LOCATION"),
		Db: db.Config{
			Driver:   os.Getenv("DB_DRIVER"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Host:     os.Getenv("DB_HOST"),
		},
		Rbmq: rbmq.Config{
			User:         os.Getenv("RBMQ_USER"),
			Password:     os.Getenv("RBMQ_PASSWORD"),
			URL:          os.Getenv("RBMQ_URL"),
			ExchangeName: os.Getenv("SERVICE_LOCATION"),
			ExchangeType: os.Getenv("RBMQ_EXCHANGE_TYPE"),
			BindingKey:   os.Getenv("RBMQ_BINDINGKEY"),
			ConsumerTag:  os.Getenv("RBMQ_CONSUMER_TAG"),
		},
	}
}

// printServices is a helper function to print the usage
func printServices() {
	fmt.Println("Invalid service name. Valid service names are:")
	fmt.Println("	[user, order, delegation, part, factory, assembly, model, shipping, kpi, ticket, support]")
}

func main() {
	var serviceInstance Service
	var serviceName string

	// initialize a new logger instance that is shared accross all objects
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		zap.L().Sugar().Fatalw("Failed to initialize logger", "err", err)
	}
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	// check if a service name is specified
	if len(os.Args) > 1 {
		serviceName = os.Args[1]
	} else {
		printServices()
		os.Exit(1)
	}

	// used to listen to ctrl+c signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	// channel for communication between the rabbitmq components and the services
	messages := make(chan rbmq.Message)

	// initialize service based on service name
	switch serviceName {
	case customerService:
		serviceInstance, err = customer.New(getConfig(), messages, logger)
	case orderService:
		serviceInstance, err = order.New(getConfig(), messages, logger)
	case delegationService:
		serviceInstance, err = delegation.New(getConfig(), messages, logger)
	case partService:
		serviceInstance, err = part.New(getConfig(), messages, logger)
	case factoryService:
		serviceInstance, err = factory.New(getConfig(), messages, logger)
	case assemblyService:
		serviceInstance, err = assembly.New(getConfig(), messages, logger)
	case modelService:
		serviceInstance, err = model.New(getConfig(), messages, logger)
	case shippingService:
		serviceInstance, err = shipping.New(getConfig(), messages, logger)
	case kpiService:
		serviceInstance, err = kpi.New(getConfig(), messages, logger)
	case ticketService:
		serviceInstance, err = ticket.New(getConfig(), messages, logger)
	case supportService:
		serviceInstance, err = support.New(getConfig(), messages, logger)
	default:
		printServices()
		os.Exit(1)
	}

	if err != nil {
		logger.Fatalw("Failed to start service", "service", serviceName, "err", err)
	} else {
		logger.Infof("Started service %s", serviceName)
	}

	// block until a shut down signal is received
	<-signalChan
	logger.Info("Received shutdown signal, attempting graceful shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	serviceInstance.Shutdown(ctx)
}
