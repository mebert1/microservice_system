package mongo

import (
	"context"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetFactoryStatus returns the current status of a factory
func (c *Client) GetFactoryStatus(location string) (entities.FactoryStatus, error) {
	status := entities.FactoryStatus{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := c.mongoClient.Database(delegationDB).Collection(delegationCol).FindOne(ctx, bson.M{"location": location})
	err := result.Decode(&status)

	return status, err
}

// UpdateFactoryStatus can be used to set a factory to a new status
func (c *Client) UpdateFactoryStatus(status entities.FactoryStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.mongoClient.Database(delegationDB).Collection(delegationCol).UpdateOne(
		ctx,
		bson.M{"location": status.Location},
		bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "currentLoad", Value: status.CurrentLoad}},
			},
		},
	)
	return err
}

// InitDelegationDatabase initializes delegation database with values
// Function is called once after the storage is initialized
func (c *Client) InitDelegationDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.mongoClient.Database(delegationDB).Drop(ctx)
	if err != nil {
		return err
	}

	var status []interface{}

	status = append(status, entities.FactoryStatus{
		Location:            "usa",
		CurrentLoad:         0,
		MaxConcurrentOrders: 10,
	})

	status = append(status, entities.FactoryStatus{
		Location:            "china",
		CurrentLoad:         0,
		MaxConcurrentOrders: 20,
	})

	_, err = c.mongoClient.Database(delegationDB).Collection(delegationCol).InsertMany(ctx, status)
	return err
}
