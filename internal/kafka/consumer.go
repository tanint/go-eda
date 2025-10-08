package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/tanint/go-eda/internal/config"
	"github.com/tanint/go-eda/internal/logger"
	"go.uber.org/zap"
)

// MessageHandler is a function type for handling consumed messages
type MessageHandler func(ctx context.Context, msg *kafka.Message) error

// Consumer wraps Kafka consumer with additional functionality
type Consumer struct {
	consumer *kafka.Consumer
	config   config.KafkaConfig
	handlers map[string]MessageHandler
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg config.KafkaConfig, groupID string) (*Consumer, error) {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers":  cfg.Brokers,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
		"session.timeout.ms": 6000,
	}

	// Add security configuration if needed
	if cfg.SecurityProtocol != "PLAINTEXT" {
		configMap.SetKey("security.protocol", cfg.SecurityProtocol)
		configMap.SetKey("sasl.mechanism", cfg.SASLMechanism)
		configMap.SetKey("sasl.username", cfg.SASLUsername)
		configMap.SetKey("sasl.password", cfg.SASLPassword)
	}

	consumer, err := kafka.NewConsumer(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	logger.Info("Kafka consumer initialized successfully",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("group_id", groupID),
	)

	return &Consumer{
		consumer: consumer,
		config:   cfg,
		handlers: make(map[string]MessageHandler),
	}, nil
}

// Subscribe subscribes to topics with their handlers
func (c *Consumer) Subscribe(topics []string) error {
	err := c.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	logger.Info("Subscribed to topics",
		zap.Strings("topics", topics),
	)

	return nil
}

// RegisterHandler registers a message handler for a specific topic
func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
	logger.Info("Registered handler for topic",
		zap.String("topic", topic),
	)
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	logger.Info("Starting Kafka consumer...")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Consumer context cancelled, stopping...")
			return ctx.Err()
		default:
			msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Timeout is not an error, continue
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				logger.Error("Error reading message",
					zap.Error(err),
				)
				continue
			}

			if err := c.processMessage(ctx, msg); err != nil {
				logger.Error("Error processing message",
					zap.Error(err),
					zap.String("topic", *msg.TopicPartition.Topic),
					zap.Int32("partition", msg.TopicPartition.Partition),
					zap.String("offset", msg.TopicPartition.Offset.String()),
				)
				// Continue processing other messages even if one fails
				continue
			}

			// Commit the message offset after successful processing
			if _, err := c.consumer.CommitMessage(msg); err != nil {
				logger.Error("Error committing message",
					zap.Error(err),
					zap.String("topic", *msg.TopicPartition.Topic),
				)
			}
		}
	}
}

// processMessage processes a single message
func (c *Consumer) processMessage(ctx context.Context, msg *kafka.Message) error {
	topic := *msg.TopicPartition.Topic

	logger.Debug("Received message",
		zap.String("topic", topic),
		zap.Int32("partition", msg.TopicPartition.Partition),
		zap.String("offset", msg.TopicPartition.Offset.String()),
		zap.ByteString("key", msg.Key),
	)

	handler, exists := c.handlers[topic]
	if !exists {
		logger.Warn("No handler registered for topic",
			zap.String("topic", topic),
		)
		return nil
	}

	// Process message with timeout
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := handler(processCtx, msg); err != nil {
		return fmt.Errorf("handler error: %w", err)
	}

	return nil
}

// Close closes the consumer
func (c *Consumer) Close() error {
	logger.Info("Closing Kafka consumer...")
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("error closing consumer: %w", err)
	}
	logger.Info("Kafka consumer closed successfully")
	return nil
}
