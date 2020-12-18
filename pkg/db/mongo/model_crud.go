package mongo

import (
	"context"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"go.mongodb.org/mongo-driver/bson"
)

// FindModel finds a model specified by its ID
func (c *Client) FindModel(id int) (entities.Model, error) {
	model := entities.Model{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := c.mongoClient.Database(modelDB).Collection(modelCol).FindOne(ctx, bson.M{"id": id})
	err := result.Decode(&model)

	return model, err
}

// AllModels returns all models in model database
func (c *Client) AllModels() ([]entities.Model, error) {
	var models []entities.Model

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := c.mongoClient.Database(modelDB).Collection(modelCol).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &models)

	return models, err
}

// UpdateModelPart updates a part in all models it is being used in
func (c *Client) UpdateModelPart(part entities.Part) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.mongoClient.Database(modelDB).Collection(modelCol).UpdateMany(
		ctx,
		bson.M{"parts.id": part.ID},
		bson.M{
			"$set": bson.M{"parts.$.price": part.Price},
		},
	)
	return err
}

// InitModelDatabase initializes model database with init values
// These values are made up or were found in requirments document to simulate a full database
func (c *Client) InitModelDatabase() error {
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

	data = append(data, entities.Model{
		Name:         "UltraCool9000",
		ID:           1,
		AssemblyTime: 10,
		Parts:        []entities.Part{parts[0], parts[2], parts[5]},
	})

	data = append(data, entities.Model{
		Name:         "IcyX",
		ID:           2,
		AssemblyTime: 7,
		Parts:        []entities.Part{parts[0], parts[1], parts[4], parts[1], parts[3]},
	})

	data = append(data, entities.Model{
		Name:         "Chiller",
		ID:           3,
		AssemblyTime: 15,
		Parts:        []entities.Part{parts[0], parts[2], parts[2], parts[1], parts[3]},
	})

	data = append(data, entities.Model{
		Name:         "CoolBoy",
		ID:           4,
		AssemblyTime: 8,
		Parts:        []entities.Part{parts[0], parts[2], parts[4], parts[5]},
	})

	_, err = c.mongoClient.Database(modelDB).Collection(modelCol).InsertMany(ctx, data)
	return err
}
