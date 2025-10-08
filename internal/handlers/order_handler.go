package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tanint/go-eda/internal/kafka"
	"github.com/tanint/go-eda/internal/logger"
	"github.com/tanint/go-eda/internal/models"
	"github.com/tanint/go-eda/pkg/events"
	"go.uber.org/zap"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	producer *kafka.Producer
	topics   map[string]string
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(producer *kafka.Producer, topics map[string]string) *OrderHandler {
	return &OrderHandler{
		producer: producer,
		topics:   topics,
	}
}

// CreateOrder handles order creation requests
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Create order
	order, err := models.NewOrder(req)
	if err != nil {
		logger.Error("Failed to create order",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Publish order created event
	event := events.NewEvent(events.EventTypeOrderCreated, events.OrderCreatedEvent{
		Order: *order,
	})

	eventData, err := event.Marshal()
	if err != nil {
		logger.Error("Failed to marshal event",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process order",
		})
		return
	}

	topic := h.topics["order_created"]
	if err := h.producer.Publish(c.Request.Context(), topic, []byte(order.ID), eventData); err != nil {
		logger.Error("Failed to publish event",
			zap.Error(err),
			zap.String("topic", topic),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process order",
		})
		return
	}

	logger.Info("Order created successfully",
		zap.String("order_id", order.ID),
		zap.String("customer_id", order.CustomerID),
		zap.Float64("total_price", order.TotalPrice),
	)

	c.JSON(http.StatusCreated, order)
}

// GetOrderStatus handles order status requests (mock implementation)
func (h *OrderHandler) GetOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	// In a real application, you would fetch this from a database
	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"status":   "pending",
		"message":  "This is a mock response. In production, implement database lookup.",
	})
}

// HealthCheck returns the health status of the service
func (h *OrderHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "order-service",
	})
}

// HandleOrderCreated handles order created events (for inventory service)
func HandleOrderCreated(ctx context.Context, producer *kafka.Producer, topics map[string]string) func(context.Context, *kafka.Message) error {
	return func(ctx context.Context, msg *kafka.Message) error {
		var event events.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			logger.Error("Failed to unmarshal event",
				zap.Error(err),
			)
			return err
		}

		// Parse the event data
		eventDataJSON, err := json.Marshal(event.Data)
		if err != nil {
			logger.Error("Failed to marshal event data",
				zap.Error(err),
			)
			return err
		}

		var orderCreated events.OrderCreatedEvent
		if err := json.Unmarshal(eventDataJSON, &orderCreated); err != nil {
			logger.Error("Failed to unmarshal order created event",
				zap.Error(err),
			)
			return err
		}

		logger.Info("Processing order created event",
			zap.String("order_id", orderCreated.Order.ID),
			zap.String("customer_id", orderCreated.Order.CustomerID),
		)

		// Reserve inventory (mock logic)
		reservations := make([]events.InventoryReservation, len(orderCreated.Order.Items))
		for i, item := range orderCreated.Order.Items {
			reservations[i] = events.InventoryReservation{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			}
		}

		// Publish inventory reserved event
		inventoryEvent := events.NewEvent(events.EventTypeInventoryReserved, events.InventoryReservedEvent{
			OrderID: orderCreated.Order.ID,
			Items:   reservations,
		})

		inventoryData, err := inventoryEvent.Marshal()
		if err != nil {
			logger.Error("Failed to marshal inventory event",
				zap.Error(err),
			)
			return err
		}

		topic := topics["inventory_reserved"]
		if err := producer.Publish(ctx, topic, []byte(orderCreated.Order.ID), inventoryData); err != nil {
			logger.Error("Failed to publish inventory event",
				zap.Error(err),
			)
			return err
		}

		logger.Info("Inventory reserved successfully",
			zap.String("order_id", orderCreated.Order.ID),
		)

		return nil
	}
}
