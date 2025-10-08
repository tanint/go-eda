package events

import (
	"encoding/json"
	"time"

	"github.com/tanint/go-eda/internal/models"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeOrderCreated       EventType = "order.created"
	EventTypeOrderConfirmed     EventType = "order.confirmed"
	EventTypeInventoryReserved  EventType = "inventory.reserved"
	EventTypeInventoryReleased  EventType = "inventory.released"
	EventTypeNotificationSent   EventType = "notification.sent"
)

// Event represents a base event structure
type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// OrderCreatedEvent represents an order creation event
type OrderCreatedEvent struct {
	Order models.Order `json:"order"`
}

// OrderConfirmedEvent represents an order confirmation event
type OrderConfirmedEvent struct {
	OrderID    string    `json:"order_id"`
	CustomerID string    `json:"customer_id"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

// InventoryReservedEvent represents an inventory reservation event
type InventoryReservedEvent struct {
	OrderID    string                  `json:"order_id"`
	Items      []InventoryReservation  `json:"items"`
	ReservedAt time.Time               `json:"reserved_at"`
}

// InventoryReservation represents a single item reservation
type InventoryReservation struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// NewEvent creates a new event with the given type and data
func NewEvent(eventType EventType, data interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// Marshal serializes the event to JSON
func (e *Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalEvent deserializes JSON to an Event
func UnmarshalEvent(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
