package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/tanint/go-eda/internal/config"
	kafkapkg "github.com/tanint/go-eda/internal/kafka"
	"github.com/tanint/go-eda/internal/logger"
	"github.com/tanint/go-eda/pkg/events"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Initialize(cfg.Logger); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Notification Service...")

	// Initialize Kafka consumer
	consumer, err := kafkapkg.NewConsumer(cfg.Kafka, "notification-service-group")
	if err != nil {
		logger.Fatal("Failed to create Kafka consumer", zap.Error(err))
	}
	defer consumer.Close()

	// Register message handlers
	inventoryReservedTopic := cfg.Kafka.Topics["inventory_reserved"]
	consumer.RegisterHandler(inventoryReservedTopic, handleInventoryReserved)

	// Subscribe to topics
	if err := consumer.Subscribe([]string{inventoryReservedTopic}); err != nil {
		logger.Fatal("Failed to subscribe to topics", zap.Error(err))
	}

	// Start consuming in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	logger.Info("Notification Service is running and consuming messages...")

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Shutting down Notification Service...")
		cancel()
	case err := <-errChan:
		logger.Error("Consumer error", zap.Error(err))
		cancel()
	}

	logger.Info("Notification Service stopped")
}

func handleInventoryReserved(ctx context.Context, msg *kafka.Message) error {
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

	var inventoryReserved events.InventoryReservedEvent
	if err := json.Unmarshal(eventDataJSON, &inventoryReserved); err != nil {
		logger.Error("Failed to unmarshal inventory reserved event",
			zap.Error(err),
		)
		return err
	}

	logger.Info("Processing inventory reserved event",
		zap.String("order_id", inventoryReserved.OrderID),
		zap.Int("items_count", len(inventoryReserved.Items)),
	)

	// Send notification (mock implementation)
	sendNotification(inventoryReserved.OrderID)

	return nil
}

func sendNotification(orderID string) {
	// This is a mock implementation
	// In production, you would integrate with email/SMS/push notification services
	logger.Info("Notification sent",
		zap.String("order_id", orderID),
		zap.String("type", "order_confirmed"),
		zap.String("message", "Your order has been confirmed and inventory has been reserved"),
	)
}
