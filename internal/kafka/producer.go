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

// Producer wraps Kafka producer with additional functionality
type Producer struct {
	producer *kafka.Producer
	config   config.KafkaConfig
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg config.KafkaConfig) (*Producer, error) {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers": cfg.Brokers,
		"client.id":         "go-eda-producer",
		"acks":              "all",
		"retries":           3,
		"max.in.flight.requests.per.connection": 5,
		"compression.type":                      "snappy",
		"linger.ms":                             5,
		"batch.size":                            16384,
	}

	// Add security configuration if needed
	if cfg.SecurityProtocol != "PLAINTEXT" {
		configMap.SetKey("security.protocol", cfg.SecurityProtocol)
		configMap.SetKey("sasl.mechanism", cfg.SASLMechanism)
		configMap.SetKey("sasl.username", cfg.SASLUsername)
		configMap.SetKey("sasl.password", cfg.SASLPassword)
	}

	producer, err := kafka.NewProducer(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	p := &Producer{
		producer: producer,
		config:   cfg,
	}

	// Start delivery report handler
	go p.handleDeliveryReports()

	logger.Info("Kafka producer initialized successfully",
		zap.Strings("brokers", cfg.Brokers),
	)

	return p, nil
}

// Publish publishes a message to the specified topic
func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)

	err := p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   key,
		Value: value,
		Headers: []kafka.Header{
			{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}, deliveryChan)

	if err != nil {
		logger.Error("Failed to produce message",
			zap.Error(err),
			zap.String("topic", topic),
		)
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for delivery report or context cancellation
	select {
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			logger.Error("Message delivery failed",
				zap.Error(m.TopicPartition.Error),
				zap.String("topic", topic),
			)
			return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
		}
		logger.Debug("Message delivered successfully",
			zap.String("topic", *m.TopicPartition.Topic),
			zap.Int32("partition", m.TopicPartition.Partition),
			zap.String("offset", m.TopicPartition.Offset.String()),
		)
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// handleDeliveryReports handles delivery reports from Kafka
func (p *Producer) handleDeliveryReports() {
	for e := range p.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				logger.Error("Delivery failed",
					zap.Error(ev.TopicPartition.Error),
					zap.String("topic", *ev.TopicPartition.Topic),
				)
			}
		case kafka.Error:
			logger.Error("Kafka error",
				zap.Error(ev),
				zap.String("code", ev.Code().String()),
			)
		}
	}
}

// Close closes the producer and flushes any pending messages
func (p *Producer) Close() error {
	logger.Info("Closing Kafka producer...")

	// Wait for all messages to be delivered (with timeout)
	outstanding := p.producer.Flush(15 * 1000) // 15 seconds
	if outstanding > 0 {
		logger.Warn("Some messages were not delivered before close",
			zap.Int("outstanding", outstanding),
		)
	}

	p.producer.Close()
	logger.Info("Kafka producer closed successfully")
	return nil
}
