package mongo

import (
	"context"
	"time"

	"git.thm.de/verteilte-systeme-2020-efridge/gruppe-13/pkg/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateTicket creates a ticket after it was opened by a customer
func (c *Client) CreateTicket(ticket entities.Ticket) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := c.mongoClient.Database(ticketDB).Collection(ticketCol).InsertOne(ctx, ticket)

	return result.InsertedID.(primitive.ObjectID).Hex(), err
}

// UpdateTicket updates a given ticket in ticket database
func (c *Client) UpdateTicket(ticket entities.Ticket) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(ticket.ObjectID)
	_, err := c.mongoClient.Database(ticketDB).Collection(ticketCol).UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.D{
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "status", Value: ticket.Status}},
			},
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "closed", Value: ticket.Closed}},
			},
			primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "response", Value: ticket.Response}},
			},
		},
	)
	return err
}

// FindTicket finds and returns a ticket with given ID
func (c *Client) FindTicket(id string) (entities.Ticket, error) {
	ticket := entities.Ticket{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, _ := primitive.ObjectIDFromHex(id)

	result := c.mongoClient.Database(ticketDB).Collection(ticketCol).FindOne(ctx, bson.M{"_id": objectID})
	err := result.Decode(&ticket)

	return ticket, err
}

// AllTickets returns all tickets in ticket database
func (c *Client) AllTickets() ([]entities.Ticket, error) {
	var tickets []entities.Ticket

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := c.mongoClient.Database(ticketDB).Collection(ticketCol).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(ctx, &tickets)

	return tickets, err
}
