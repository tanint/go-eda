package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tanint/go-eda/internal/config"
	"github.com/tanint/go-eda/internal/handlers"
	"github.com/tanint/go-eda/internal/kafka"
	"github.com/tanint/go-eda/internal/logger"
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

	logger.Info("Starting Inventory Service...")

	// Initialize Kafka producer (for publishing events)
	producer, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		logger.Fatal("Failed to create Kafka producer", zap.Error(err))
	}
	defer producer.Close()

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Kafka, "inventory-service-group")
	if err != nil {
		logger.Fatal("Failed to create Kafka consumer", zap.Error(err))
	}
	defer consumer.Close()

	// Register message handlers
	orderCreatedTopic := cfg.Kafka.Topics["order_created"]
	consumer.RegisterHandler(orderCreatedTopic, handlers.HandleOrderCreated(context.Background(), producer, cfg.Kafka.Topics))

	// Subscribe to topics
	if err := consumer.Subscribe([]string{orderCreatedTopic}); err != nil {
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

	logger.Info("Inventory Service is running and consuming messages...")

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Shutting down Inventory Service...")
		cancel()
	case err := <-errChan:
		logger.Error("Consumer error", zap.Error(err))
		cancel()
	}

	logger.Info("Inventory Service stopped")
}
