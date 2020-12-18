package rbmq

import "time"

/*
This package is a collection of all message formats used for communication between services
*/

// OrderMessage contains all information relevant to an order
type OrderMessage struct {
	Timestamp    time.Time `json:"timestamp,omitempty"`
	MsgType      string    `json:"type,omitempty"`
	Customer     string    `json:"customer,omitempty"`
	OrderID      string    `json:"order,omitempty"`
	Status       string    `json:"status,omitempty"`
	Location     string    `json:"location,omitempty"`
	Items        []Item    `json:"items,omitempty"`
	CostsOfParts int       `json:"costsOfParts,omitempty"`
}

// PartMessage contains all information about a single part
type PartMessage struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	MsgType   string    `json:"type,omitempty"`
	Part      int       `json:"part,omitempty"`
	Price     int       `json:"price,omitempty"`
}

// Item contains all information about a single fridge
type Item struct {
	ItemID       int   `json:"item,omitempty"`
	Parts        []int `json:"parts,omitempty"`
	AssemblyTime int   `json:"assemblytime,omitempty"`
}

// TicketMessage contains all information required to resolve a support ticket
type TicketMessage struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	MsgType   string    `json:"type,omitempty"`
	TicketID  string    `json:"ticketID,omitempty"`
	Response  string    `json:"response,omitempty"`
}

// KPIMessage contains all information used to create new KPI entries
type KPIMessage struct {
	Timestamp        time.Time `json:"timestamp,omitempty"`
	MsgType          string    `json:"type,omitempty"`
	Location         string    `json:"location,omitempty"`
	IncompleteOrders int       `json:"incompleteOrders,omitempty"`
	CompletedOrders  int       `json:"completedOrders,omitempty"`
	Total            int       `json:"total"`
	CostsOfParts     int       `json:"costsOfParts,omitempty"`
}
