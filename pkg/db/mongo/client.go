package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	customerDB  = "customer"
	customerCol = "data"

	orderDB  = "order"
	orderCol = "data"

	factoryDB  = "factory"
	factoryCol = "data"

	delegationDB  = "delegation"
	delegationCol = "status"

	ticketDB  = "ticket"
	ticketCol = "data"

	kpiDB  = "kpi"
	kpiCol = "data"

	modelDB  = "model"
	modelCol = "data"

	partDB      = "parts"
	partCol     = "data"
	supplierCol = "data"
)

// Client is a wrapper for a database connection
type Client struct {
	mongoClient *mongo.Client
	database    string
}

// New returns a new database connection
func New(user, password, host string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	url := fmt.Sprintf("mongodb://%s:%s@%s:27017", user, password, host)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	return &Client{
		mongoClient: client,
	}, nil
}

// Close ends current database connection
func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.mongoClient.Disconnect(ctx)
}
