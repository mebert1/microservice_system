package mongo

import (
	"context"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdatePart updates price of a part
// The part service is responsible for the update of part prices
// Price updates are received from the model service as soon as is it notified of a price change by a supplier
func (c *Client) UpdatePart(part entities.Part) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//objectID, _ := primitive.ObjectIDFromHex(part.ID)
	_, err := c.mongoClient.Database(partDB).Collection(partCol).UpdateOne(
		ctx,
		bson.M{"id": part.ID},
		bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "price", Value: part.Price}},
			},
		},
	)
	return err
}

// FindPart finds and returns a part specified by its ID
func (c *Client) FindPart(id int) (entities.Part, error) {
	part := entities.Part{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//objectID, _ := primitive.ObjectIDFromHex(id)

	result := c.mongoClient.Database(partDB).Collection(partCol).FindOne(ctx, bson.M{"id": id})
	err := result.Decode(&part)

	return part, err
}

//FindSupplier finds and returns a supplier specified by its ID
func (c *Client) FindSupplier(id string) (entities.Supplier, error) {
	supplier := entities.Supplier{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(id)

	result := c.mongoClient.Database(partDB).Collection(supplierCol).FindOne(ctx, bson.M{"_id": objectID})
	err := result.Decode(&supplier)

	return supplier, err
}

// InitPartDatabase niitializes part database with init values
// These values are made up or were found in requirments document to simulate a full database
func (c *Client) InitPartDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.mongoClient.Database(modelDB).Drop(ctx)
	if err != nil {
		return err
	}

	var data []interface{}

	suppliers := []entities.Supplier{
		{
			Name: "electroStuff.com",
			Address: entities.Address{
				Country: "Germany",
				City:    "Reinheim",
				ZIP:     64354,
				Address: "Anne-Frank-Straße 23",
			},
		},
		{
			Name: "coolMechanics.com",
			Address: entities.Address{
				Country: "Denmark",
				City:    "Anaago",
				ZIP:     33674,
				Address: "Bjutsche 7",
			},
		},
	}

	parts := []entities.Part{
		{
			Price:    139,
			ID:       1,
			Supplier: suppliers[0],
		},
		{
			Price:    53,
			ID:       2,
			Supplier: suppliers[1],
		},
		{
			Price:    18,
			ID:       3,
			Supplier: suppliers[1],
		},
		{
			Price:    223,
			ID:       4,
			Supplier: suppliers[0],
		},
		{
			Price:    140,
			ID:       5,
			Supplier: suppliers[0],
		},
		{
			Price:    98,
			ID:       6,
			Supplier: suppliers[1],
		},
	}

	data = append(data, parts[0])

	data = append(data, parts[1])

	data = append(data, parts[2])

	data = append(data, parts[3])

	data = append(data, parts[4])

	data = append(data, parts[5])

	_, err = c.mongoClient.Database(partDB).Collection(partCol).InsertMany(ctx, data)
	return err
}
