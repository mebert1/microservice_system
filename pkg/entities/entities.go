package entities

import "time"

/*
This package is a collection of all data objects shared between the services as well as databases
*/

// Customer is the entity to save eFridge customers' data
type Customer struct {
	ObjectID  string    `json:"objectID,omitempty" bson:"_id,omitempty"`
	ID        int       `json:"id,omitempty" bson:"id,omitempty"`
	FirstName string    `json:"firstname" bson:"firstname"`
	LastName  string    `json:"lastname" bson:"lastname"`
	Address   Address   `json:"address" bson:"address"`
	Created   time.Time `json:"created" bson:"created"`
}

// Address is the entitiy to save addresses of customers and suppliers
type Address struct {
	Country string `json:"country" bson:"country"`
	City    string `json:"city" bson:"city"`
	ZIP     int    `json:"zipCode" bson:"zipCode"`
	Address string `json:"address" bson:"address"`
}

// Order is the entity used to control the order flow and constantly update with a new status
type Order struct {
	ObjectID     string    `json:"objectID,omitempty" bson:"_id,omitempty"`
	OrderID      string    `json:"orderID,omitemtpy" bson:"orderID,omitempty"`
	Created      time.Time `json:"created" bson:"created"`
	Customer     string    `json:"customer" bson:"customer"`
	Status       string    `json:"status" bson:"status"`
	Items        []int     `json:"items" bson:"items"`
	LastUpdate   time.Time `json:"lastUpdate" bson:"lastUpdate"`
	CostsOfParts int       `json:"costsOfParts,omitempty" bson:"costsOfParts,omitempty"`
}

// FactoryStatus is the entity that holds information about the load of a single factory
type FactoryStatus struct {
	ObjectID            string `json:"objectID,omitempty" bson:"_id,omitempty"`
	ID                  string `json:"id,omitempty" bson:"id,omitempty"`
	Location            string `json:"location" bson:"location"`
	CurrentLoad         int    `json:"currentLoad" bson:"currentLoad"`
	MaxConcurrentOrders int    `json:"maxConcurrentOrders" bson:"maxConcurrentOrders"`
}

// Ticket is the entity used to hold information of a support ticket
type Ticket struct {
	ObjectID string    `json:"objectID,omitempty" bson:"_id,omitempty"`
	ID       int       `json:"id,omitempty" bson:"id,omitempty"`
	Created  time.Time `json:"created" bson:"created"`
	Closed   time.Time `json:"closed" bson:"closed"`
	Status   string    `json:"status" bson:"status"`
	Text     string    `json:"text" bson:"text"`
	Response string    `json:"response" bson:"response"`
	Location string    `json:"location" bson:"location"`
}

// KPI is th entity that combines all relevant KPIs
type KPI struct {
	ObjectID         string    `json:"objectID,omitempty" bson:"_id,omitempty"`
	ID               int       `json:"id,omitempty" bson:"id,omitempty"`
	Created          time.Time `json:"created" bson:"created"`
	Location         string    `json:"location" bson:"location"`
	IncompleteOrders int       `json:"incompleteOrders" bson:"incompleteOrders"`
	CompletedOrders  int       `json:"completedOrders" bson:"completedOrders"`
	Total            int       `json:"total" bson:"total"`
	CostsOfParts     int       `json:"costsOfParts" bson:"costsOfParts"`
}

// Supplier is the supplier object
type Supplier struct {
	ObjectID string  `json:"objectID,omitempty" bson:"_id,omitempty"`
	ID       int     `json:"id,omitempty" bson:"id,omitempty"`
	Name     string  `json:"name" bson:"name"`
	Address  Address `json:"address" bson:"address"`
}

// Part is the part object
type Part struct {
	ObjectID string   `json:"objectID,omitempty" bson:"_id,omitempty"`
	ID       int      `json:"id,omitempty" bson:"id,omitempty"`
	Price    int      `json:"price,omitempty" bson:"price,omitempty"`
	Supplier Supplier `json:"supplier,omitempty" bson:"supplier,omitempty"`
}

// Model is the fridge object
type Model struct {
	ObjectID     string `json:"objectID,omitempty" bson:"_id,omitempty"`
	ID           int    `json:"id,omitempty" bson:"id,omitempty"`
	Name         string `json:"name,omitempty" bson:"name,omitempty"`
	AssemblyTime int    `json:"assemblytime,omitempty" bson:"assemblyTime,omitempty"`
	Parts        []Part `json:"parts" bson:"parts"`
}
