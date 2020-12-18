package mongo

import (
	"context"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateOrder reates order in order database
// Orders are received after a customer creates an order and send it to the order service through a REST POST command
func (c *Client) CreateOrder(order entities.Order) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := c.mongoClient.Database(orderDB).Collection(orderCol).InsertOne(ctx, order)

	return result.InsertedID.(primitive.ObjectID).Hex(), err
}

// UpdateOrderStatus status updates the status of a given order
// Status are updated after every manufactoring, assembling and shipping step
func (c *Client) UpdateOrderStatus(order entities.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(order.ObjectID)
	_, err := c.mongoClient.Database(orderDB).Collection(orderCol).UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "status", Value: order.Status}},
			},
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "lastUpdate", Value: order.LastUpdate}},
			},
		},
	)
	return err
}

// FindOrder returns the order with the given ID from the order database
func (c *Client) FindOrder(id string) (entities.Order, error) {
	order := entities.Order{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(id)

	result := c.mongoClient.Database(orderDB).Collection(orderCol).FindOne(ctx, bson.M{"_id": objectID})
	err := result.Decode(&order)

	return order, err
}

// AllOrders returns all orders within order database
func (c *Client) AllOrders() ([]entities.Order, error) {
	var orders []entities.Order

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := c.mongoClient.Database(orderDB).Collection(orderCol).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &orders)

	return orders, err
}
