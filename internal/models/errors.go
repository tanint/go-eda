package models

import "errors"

var (
	// Order errors
	ErrInvalidProductID = errors.New("invalid product ID")
	ErrInvalidQuantity  = errors.New("quantity must be greater than 0")
	ErrInvalidPrice     = errors.New("price cannot be negative")
	ErrOrderNotFound    = errors.New("order not found")

	// Inventory errors
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrProductNotFound   = errors.New("product not found")

	// Kafka errors
	ErrProducerNotInitialized = errors.New("kafka producer not initialized")
	ErrConsumerNotInitialized = errors.New("kafka consumer not initialized")
	ErrFailedToPublish        = errors.New("failed to publish message")
)
