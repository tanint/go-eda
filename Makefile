.PHONY: help install build run-order run-inventory run-notification docker-up docker-down test clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install Go dependencies
	go mod download
	go mod tidy

build: ## Build all services
	@echo "Building services..."
	@mkdir -p bin
	go build -o bin/order-service ./cmd/order-service
	go build -o bin/inventory-service ./cmd/inventory-service
	go build -o bin/notification-service ./cmd/notification-service
	@echo "Build completed!"

run-order: ## Run order service
	go run ./cmd/order-service/main.go

run-inventory: ## Run inventory service
	go run ./cmd/inventory-service/main.go

run-notification: ## Run notification service
	go run ./cmd/notification-service/main.go

docker-up: ## Start Kafka and dependencies with Docker Compose
	docker-compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 10
	@echo "Services are ready!"
	@echo "Kafka UI: http://localhost:8090"

docker-down: ## Stop Docker Compose services
	docker-compose down -v

docker-logs: ## Show Docker Compose logs
	docker-compose logs -f

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	go tool cover -html=coverage.out

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

fmt: ## Format Go code
	go fmt ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run

# Local development workflow
dev-setup: docker-up install ## Setup local development environment
	@echo "Development environment is ready!"
	@echo "Run services in separate terminals:"
	@echo "  - make run-order"
	@echo "  - make run-inventory"
	@echo "  - make run-notification"

dev-clean: docker-down clean ## Clean up development environment
