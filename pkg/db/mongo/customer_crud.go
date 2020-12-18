package mongo

import (
	"context"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateCustomer creates a customer
// The customer entity is given from the outside
func (c *Client) CreateCustomer(customer entities.Customer) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := c.mongoClient.Database(customerDB).Collection(customerCol).InsertOne(ctx, customer)

	return result.InsertedID.(primitive.ObjectID).Hex(), err
}

// FindCustomer searches database for customer with given ID
// Can be used for all kinds of customer operations
func (c *Client) FindCustomer(id string) (entities.Customer, error) {
	customer := entities.Customer{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(id)

	result := c.mongoClient.Database(customerDB).Collection(customerCol).FindOne(ctx, bson.M{"_id": objectID})
	err := result.Decode(&customer)

	return customer, err
}

// AllCustomers returns all customers to caller
func (c *Client) AllCustomers() ([]entities.Customer, error) {
	var customers []entities.Customer

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := c.mongoClient.Database(customerDB).Collection(customerCol).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &customers)

	return customers, err
}
