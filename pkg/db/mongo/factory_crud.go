package mongo

import (
	"context"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateOrderFactory creates an order that is written into factory database
// Service is required because factory service and order service have two seperate databases
func (c *Client) CreateOrderFactory(order entities.Order) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := c.mongoClient.Database(factoryDB).Collection(factoryCol).InsertOne(ctx, order)

	return result.InsertedID.(primitive.ObjectID).Hex(), err
}

// UpdateOrderStatusFactory updates the status of an order in the factory database
// Order status can be "partsdelivered" and "complete" before it is sent to shipping service
func (c *Client) UpdateOrderStatusFactory(order entities.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.mongoClient.Database(factoryDB).Collection(factoryCol).UpdateOne(
		ctx,
		bson.M{"orderID": order.OrderID},
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

// UpdateOrderCosts updates the status of an order in the factory database
func (c *Client) UpdateOrderCosts(order entities.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.mongoClient.Database(factoryDB).Collection(factoryCol).UpdateOne(
		ctx,
		bson.M{"orderID": order.OrderID},
		bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "costsOfParts", Value: order.CostsOfParts}},
			},
		},
	)
	return err
}

// AggregateKPI notifies KPI service of current factory load
// Function is called cyclic with a timer to simulate multiple KPI exchanges per day
func (c *Client) AggregateKPI() ([]entities.KPI, error) {
	var kpis []entities.KPI

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This object is used as a pipeline stage in mongo aggregations
	// it groups all entries by a single _id and sums up the complete and incomplete
	// orders aswell as their part costs
	group := bson.D{
		primitive.E{
			Key: "$group",
			Value: bson.D{
				primitive.E{Key: "_id", Value: "kpi"},
				primitive.E{Key: "total", Value: bson.M{
					"$sum": 1,
				}},
				primitive.E{Key: "completedOrders", Value: bson.M{
					"$sum": bson.M{
						"$switch": bson.M{
							"branches": bson.A{
								bson.M{
									"case": bson.M{
										"$eq": bson.A{
											"$status",
											"shipped",
										},
									},
									"then": 1,
								},
							},
							"default": 0,
						},
					},
				}},
				primitive.E{Key: "costsOfParts", Value: bson.M{"$sum": "$costsOfParts"}},
			},
		},
	}

	cursor, err := c.mongoClient.Database(factoryDB).Collection(factoryCol).Aggregate(ctx, mongo.Pipeline{group})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &kpis)
	return kpis, err
}
