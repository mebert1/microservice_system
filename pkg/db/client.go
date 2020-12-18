package db

import (
	"fmt"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/db/mongo"
	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
)

const (
	mongoDriver    = "mongo"
	postgresDriver = "postgres"
	memoryDriver   = "memory"
)

// Client is a database storage interface
// the interface allows the services to theoretically make use of differnt database
// backend without the need to implement new CRUD methods
type Client interface {
	// customer_crud
	CreateCustomer(entities.Customer) (string, error)
	FindCustomer(string) (entities.Customer, error)
	AllCustomers() ([]entities.Customer, error)

	// part_crud
	FindSupplier(string) (entities.Supplier, error)
	UpdatePart(entities.Part) error
	FindPart(int) (entities.Part, error)
	InitPartDatabase() error

	// order_crud
	CreateOrder(entities.Order) (string, error)
	UpdateOrderStatus(entities.Order) error
	FindOrder(string) (entities.Order, error)
	AllOrders() ([]entities.Order, error)

	// factory_crud
	CreateOrderFactory(entities.Order) (string, error)
	UpdateOrderStatusFactory(entities.Order) error
	UpdateOrderCosts(entities.Order) error
	AggregateKPI() ([]entities.KPI, error)

	// delegation_crud
	GetFactoryStatus(string) (entities.FactoryStatus, error)
	UpdateFactoryStatus(entities.FactoryStatus) error
	InitDelegationDatabase() error

	// ticker_crud
	CreateTicket(entities.Ticket) (string, error)
	UpdateTicket(entities.Ticket) error
	FindTicket(string) (entities.Ticket, error)
	AllTickets() ([]entities.Ticket, error)

	// kpi_crud
	CreateKPI(entities.KPI) (string, error)
	FindKPI(string) ([]entities.KPI, error)
	FindLastNKPI(string, int64) ([]entities.KPI, error)

	// model_crud
	FindModel(int) (entities.Model, error)
	AllModels() ([]entities.Model, error)
	UpdateModelPart(entities.Part) error
	InitModelDatabase() error

	Close() error
}

// Config is the database config
type Config struct {
	Driver   string
	User     string
	Password string
	Host     string
}

// New returns a new database connection
func New(config Config) (Client, error) {
	switch config.Driver {
	case mongoDriver:
		return mongo.New(config.User, config.Password, config.Host)
	case postgresDriver:
		return nil, nil
	case memoryDriver:
		return nil, nil
	default:
		return nil, fmt.Errorf("Unknown database driver %s", config.Driver)
	}
}
