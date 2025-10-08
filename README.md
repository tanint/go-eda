# Go Event-Driven Architecture with Kafka

A production-grade POC project for Go and Kafka Event-Driven Architecture

## 🏗️ Architecture

Event-Driven Order Management System consisting of 3 microservices:

```text
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│  Order Service  │─────▶│  Kafka Cluster  │─────▶│Inventory Service│
│   (HTTP API)    │      │                 │      │   (Consumer)    │
└─────────────────┘      └─────────────────┘      └─────────────────┘
                                  │
                                  ▼
                         ┌─────────────────┐
                         │Notification Svc │
                         │   (Consumer)    │
                         └─────────────────┘
```

### Event Flow

1. **Order Service**: Receives HTTP requests to create orders → publishes `order.created` event
2. **Inventory Service**: Consumes `order.created` → reserves inventory → publishes `inventory.reserved` event
3. **Notification Service**: Consumes `inventory.reserved` → sends notifications

## 🚀 Tech Stack

- **Language**: Go 1.25+
- **Kafka Client**: [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go) - Official Kafka client
- **HTTP Framework**: [Gin](https://github.com/gin-gonic/gin) - High-performance web framework
- **Configuration**: [Viper](https://github.com/spf13/viper) - Environment & file configuration
- **Logging**: [Zap](https://github.com/uber-go/zap) - Structured, high-performance logging
- **Container**: Docker & Docker Compose

## 📁 Project Structure

```text
.
├── cmd/                          # Application entry points
│   ├── order-service/           # Order HTTP API service
│   ├── inventory-service/       # Inventory consumer service
│   └── notification-service/    # Notification consumer service
├── internal/                     # Private application code
│   ├── config/                  # Configuration management
│   ├── kafka/                   # Kafka producer/consumer wrappers
│   ├── logger/                  # Logging utilities
│   ├── models/                  # Domain models & errors
│   └── handlers/                # HTTP & event handlers
├── pkg/                         # Public libraries
│   └── events/                  # Event definitions
├── configs/                     # Configuration files
│   ├── config.local.yaml       # Local development config
│   └── config.confluent.yaml   # Confluent Cloud config
├── docker-compose.yml           # Local Kafka setup
├── Makefile                     # Development commands
└── README.md
```

## 🛠️ Prerequisites

- **Go**: 1.25 or higher
- **Docker**: For running Kafka locally
- **Make**: For using Makefile commands (optional)

## 📦 Installation

### 1. Clone the repository

```bash
git clone <repository-url>
cd go-eda
```

### 2. Install dependencies

```bash
make install
# or
go mod download
```

### 3. Setup local environment

```bash
make dev-setup
```

This command will:

- Start Kafka and Zookeeper with Docker Compose
- Create necessary Kafka topics
- Download Go dependencies

## 🏃 Running the Application

### Local Development

1. **Start Kafka and dependencies**

   ```bash
   make docker-up
   ```

   Kafka UI will be available at: <http://localhost:8090>

2. **Run services** (each service in a separate terminal)

   Terminal 1 - Order Service:

   ```bash
   make run-order
   ```

   Terminal 2 - Inventory Service:

   ```bash
   make run-inventory
   ```

   Terminal 3 - Notification Service:

   ```bash
   make run-notification
   ```

### Build and Run

```bash
# Build all services
make build

# Run individual services
./bin/order-service
./bin/inventory-service
./bin/notification-service
```

## 🧪 Testing the Application

### 1. Create an Order

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-123",
    "items": [
      {
        "product_id": "product-001",
        "quantity": 2,
        "price": 99.99
      },
      {
        "product_id": "product-002",
        "quantity": 1,
        "price": 149.99
      }
    ]
  }'
```

### 2. Check Order Status

```bash
curl http://localhost:8080/api/v1/orders/{order_id}
```

### 3. Health Check

```bash
curl http://localhost:8080/health
```

### 4. Monitor Events in Kafka UI

Open <http://localhost:8090> and view topics:

- `order.created`
- `inventory.reserved`

## ⚙️ Configuration

### Local Development Configuration

Use `configs/config.local.yaml` or environment variables:

```bash
export APP_SERVER_PORT=8080
export APP_KAFKA_BROKERS=localhost:9092
export APP_LOGGER_LEVEL=info
```

### Confluent Cloud

1. **Copy and edit config**

   ```bash
   cp configs/config.confluent.yaml configs/config.yaml
   ```

2. **Set credentials via environment variables**

   ```bash
   export APP_KAFKA_BROKERS=pkc-xxxxx.us-east-1.aws.confluent.cloud:9092
   export APP_KAFKA_SECURITY_PROTOCOL=SASL_SSL
   export APP_KAFKA_SASL_MECHANISM=PLAIN
   export APP_KAFKA_SASL_USERNAME=your-api-key
   export APP_KAFKA_SASL_PASSWORD=your-api-secret
   ```

3. **Run services**

   ```bash
   make run-order
   ```

### Configuration Priority

1. Environment variables (highest priority)
2. Config file specified via command line
3. `./config.yaml`
4. `./configs/config.yaml`
5. Default values (lowest priority)

## 🔧 Available Make Commands

```bash
make help              # Show all available commands
make install           # Install Go dependencies
make build             # Build all services
make run-order         # Run order service
make run-inventory     # Run inventory service
make run-notification  # Run notification service
make docker-up         # Start Kafka with Docker Compose
make docker-down       # Stop Docker Compose services
make docker-logs       # Show Docker logs
make test              # Run tests
make test-coverage     # Run tests with coverage report
make clean             # Clean build artifacts
make fmt               # Format Go code
make dev-setup         # Setup local development environment
make dev-clean         # Clean up development environment
```

## 🔍 Production Best Practices

### 1. Configuration Management

- Viper supports multiple configuration sources
- Environment variables override file config
- Sensitive data (credentials) are not committed to code

### 2. Logging

- Structured logging with Zap
- Configurable log levels
- JSON encoding for production, console for development

### 3. Kafka Producer

- Idempotent producer (acks=all)
- Retry mechanism
- Delivery confirmation
- Graceful shutdown with flush

### 4. Kafka Consumer

- Manual offset commit (at-least-once delivery)
- Consumer groups for load balancing
- Graceful shutdown
- Error handling and logging

### 5. HTTP Server

- Timeouts configuration
- Graceful shutdown
- Structured logging middleware
- Error handling

### 6. Error Handling

- Custom error types
- Proper error wrapping
- Comprehensive logging

### 7. Code Organization

- Clean architecture
- Separation of concerns
- Reusable components

## 📝 Environment Variables Reference

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `APP_SERVER_PORT` | HTTP server port | `8080` | `8080` |
| `APP_SERVER_HOST` | HTTP server host | `0.0.0.0` | `0.0.0.0` |
| `APP_KAFKA_BROKERS` | Kafka broker addresses | `localhost:9092` | `localhost:9092` |
| `APP_KAFKA_SECURITY_PROTOCOL` | Security protocol | `PLAINTEXT` | `SASL_SSL` |
| `APP_KAFKA_SASL_MECHANISM` | SASL mechanism | - | `PLAIN` |
| `APP_KAFKA_SASL_USERNAME` | Kafka username/API key | - | `your-api-key` |
| `APP_KAFKA_SASL_PASSWORD` | Kafka password/secret | - | `your-api-secret` |
| `APP_KAFKA_GROUP_ID` | Consumer group ID | `default-group` | `inventory-group` |
| `APP_LOGGER_LEVEL` | Log level | `info` | `debug`, `info`, `warn`, `error` |
| `APP_LOGGER_ENCODING` | Log encoding | `json` | `json`, `console` |

## 🐛 Troubleshooting

### Kafka not connecting

```bash
# Check if Kafka is running
docker ps

# Check Kafka logs
make docker-logs

# Restart Kafka
make docker-down && make docker-up
```

### Dependencies issues

```bash
# Clean and reinstall
go clean -modcache
make install
```

### Build issues

```bash
# Clean and rebuild
make clean
make build
```

## 📚 Additional Resources

- [Confluent Kafka Go Client Docs](https://docs.confluent.io/kafka-clients/go/current/overview.html)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [Uber Zap Logger](https://github.com/uber-go/zap)
- [Viper Configuration](https://github.com/spf13/viper)

## 📄 License

This is a POC project for learning purposes.

---

**Happy Coding!** 🚀
