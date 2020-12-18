package mongo

import (
	"context"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateKPI creates a new KPI entrance in KPI database
func (c *Client) CreateKPI(kpi entities.KPI) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := c.mongoClient.Database(kpiDB).Collection(kpiCol).InsertOne(ctx, kpi)

	return result.InsertedID.(primitive.ObjectID).Hex(), err
}

// FindKPI searches returns KPIs with given factory location in KPI database
func (c *Client) FindKPI(location string) ([]entities.KPI, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if location == "" {
		return c.aggregateKPIs(ctx)
	}
	return c.findKPIForLocation(ctx, location, 1)
}

// FindLastNKPI searches and returns the last n KPIs of given factory location
func (c *Client) FindLastNKPI(location string, n int64) ([]entities.KPI, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.findKPIForLocation(ctx, location, n)
}

// aggregateKPIs returns the a sum of all kpi entries
func (c *Client) aggregateKPIs(ctx context.Context) ([]entities.KPI, error) {
	var kpis []entities.KPI
	sort := bson.D{primitive.E{Key: "$sort", Value: bson.D{primitive.E{Key: "created", Value: -1}}}}
	group := bson.D{
		primitive.E{
			Key: "$group",
			Value: bson.D{
				primitive.E{Key: "_id", Value: "$location"},
				primitive.E{Key: "created", Value: bson.D{primitive.E{Key: "$first", Value: "$created"}}},
				primitive.E{Key: "location", Value: bson.D{primitive.E{Key: "$first", Value: "$location"}}},
				primitive.E{Key: "incompleteOrders", Value: bson.D{primitive.E{Key: "$first", Value: "$incompleteOrders"}}},
				primitive.E{Key: "completedOrders", Value: bson.D{primitive.E{Key: "$first", Value: "$completedOrders"}}},
				primitive.E{Key: "total", Value: bson.D{primitive.E{Key: "$first", Value: "$total"}}},
				primitive.E{Key: "costsOfParts", Value: bson.D{primitive.E{Key: "$first", Value: "$costsOfParts"}}},
			}}}

	cursor, err := c.mongoClient.Database(kpiDB).Collection(kpiCol).Aggregate(ctx, mongo.Pipeline{sort, group})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &kpis)
	return kpis, err
}

// findKPIForLocation returns the last n kpi reports for a specific factory
func (c *Client) findKPIForLocation(ctx context.Context, location string, limit int64) ([]entities.KPI, error) {
	var kpis []entities.KPI

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{primitive.E{Key: "created", Value: -1}})
	findOptions.SetLimit(limit)

	cursor, err := c.mongoClient.Database(kpiDB).Collection(kpiCol).Find(ctx, bson.M{"location": location}, findOptions)
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &kpis)

	return kpis, err
}
