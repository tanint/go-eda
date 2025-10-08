package models

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order represents an order in the system
type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	CustomerID string      `json:"customer_id" binding:"required"`
	Items      []OrderItem `json:"items" binding:"required,min=1,dive"`
}

// Validate validates the order item
func (oi *OrderItem) Validate() error {
	if oi.ProductID == "" {
		return ErrInvalidProductID
	}
	if oi.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if oi.Price < 0 {
		return ErrInvalidPrice
	}
	return nil
}

// NewOrder creates a new order from a create request
func NewOrder(req CreateOrderRequest) (*Order, error) {
	order := &Order{
		ID:         uuid.New().String(),
		CustomerID: req.CustomerID,
		Items:      req.Items,
		Status:     OrderStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Calculate total price
	var total float64
	for _, item := range order.Items {
		if err := item.Validate(); err != nil {
			return nil, err
		}
		total += item.Price * float64(item.Quantity)
	}
	order.TotalPrice = total

	return order, nil
}
