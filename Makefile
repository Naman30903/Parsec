REPO_ROOT := $(shell pwd)
BINARY := $(REPO_ROOT)/bin/processor
DOCKER_IMAGE := parsec-processor:latest

.PHONY: up down down-clean build build-docker rebuild test test-integration test-verbose test-cover fmt clean deps lint logs health help

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## up: Start all services with Docker Compose
up:
	docker-compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@docker-compose ps

## down: Stop all services
down:
	docker-compose down

## down-clean: Stop all services and remove volumes
down-clean:
	docker-compose down -v
	@echo "All volumes removed"

## build: Build local binary
build:
	@echo "Building processor binary..."
	go build -o $(BINARY) ./cmd/processor
	@echo "Binary built: $(BINARY)"

## build-docker: Build Docker image
build-docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

## rebuild: Rebuild and restart processor service
rebuild:
	@echo "Rebuilding processor..."
	docker-compose up -d --build processor
	@echo "Processor rebuilt and restarted"

## restart-processor: Restart processor service only
restart-processor:
	docker-compose restart processor
	@sleep 2
	@docker-compose logs -f processor

## test: Run all tests
test:
	@echo "Running tests..."
	go test ./...

## test-integration: Run integration tests (requires Kafka)
test-integration:
	@echo "Running integration tests..."
	KAFKA_TEST=1 go test -v ./test/kafka_test/...

## test-verbose: Run tests with verbose output
test-verbose:
	go test -v ./...

## test-cover: Run tests with coverage report
test-cover:
	go test -cover ./...

## test-thunder: Run Thunder Client test suite
test-thunder:
	@echo "Running Thunder Client tests..."
	./thunderclient/run_tests.sh

## fmt: Format code
fmt:
	go fmt ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/
	@echo "Build artifacts removed"

## deps: Download and tidy dependencies
deps:
	go mod download
	go mod tidy
	@echo "Dependencies updated"

## lint: Run linter
lint:
	go vet ./...

## logs: View all service logs
logs:
	docker-compose logs -f

## logs-processor: View processor logs only
logs-processor:
	docker-compose logs -f processor

## logs-kafka: View Kafka logs
logs-kafka:
	docker-compose logs -f kafka

## logs-clickhouse: View ClickHouse logs
logs-clickhouse:
	docker-compose logs -f clickhouse

## health: Check service health
health:
	@echo "Checking service health..."
	@echo "\n=== Processor ==="
	@curl -s http://localhost:8080/health | jq . || echo "❌ Processor not ready"
	@echo "\n=== ClickHouse ==="
	@curl -s http://localhost:8123/ping || echo "❌ ClickHouse not ready"
	@echo "\n=== Redis ==="
	@redis-cli -h localhost ping || echo "❌ Redis not ready"
	@echo "\n=== Docker Services ==="
	@docker-compose ps

## stats: Check processor statistics
stats:
	@curl -s http://localhost:8080/stats | jq .

## metrics: Check Prometheus metrics
metrics:
	@curl -s http://localhost:8080/metrics | grep parsec

## kafka-topics: List Kafka topics
kafka-topics:
	docker exec parsec-kafka kafka-topics --bootstrap-server localhost:9092 --list

## kafka-consume: Consume messages from log-events topic
kafka-consume:
	docker exec -it parsec-kafka kafka-console-consumer \
		--bootstrap-server localhost:9092 \
		--topic log-events \
		--from-beginning

## clickhouse-client: Connect to ClickHouse CLI
clickhouse-client:
	docker exec -it parsec-clickhouse clickhouse-client \
		--user parsec \
		--password parsec123 \
		--database logs

## redis-cli: Connect to Redis CLI
redis-cli:
	docker exec -it parsec-redis redis-cli

## test-ingest: Send test event to ingest endpoint
test-ingest:
	@curl -X POST http://localhost:8080/ingest \
		-H "Content-Type: application/json" \
		-H "X-API-Key: test-api-key-123" \
		-d '{"id":"test-$(shell date +%s)","tenant_id":"demo","timestamp":"$(shell date -u +%Y-%m-%dT%H:%M:%SZ)","severity":"INFO","source":"makefile","message":"Test message from Makefile"}' | jq .

## open-kafdrop: Open Kafdrop UI in browser
open-kafdrop:
	@echo "Opening Kafdrop UI..."
	@open http://localhost:9000 || xdg-open http://localhost:9000 || echo "Open http://localhost:9000 in your browser"

.DEFAULT_GOAL := help
