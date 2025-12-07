Parsec — Distributed Log Processing & Alerting (skeleton)

This repository is a skeleton project for a distributed log processing and alerting system with real-time analytics.

Core design choices (configurable):
- Broker: Kafka (topic-per-tenant or keyed partitions)
- Processor: Go (low-latency, efficient concurrency)
- Storage: ClickHouse for large-scale analytics; Postgres (partitioned) for smaller scale
- Stateful processing: Kafka consumer offsets + Redis for ephemeral state; persist aggregated checkpoints to DB
- Alerting: Start with rule-based alerts; add ML anomaly detectors as a next step

This scaffold provides:
- Docker Compose for local dev (Kafka, Zookeeper, Redis, ClickHouse, Postgres)
- Minimal Go service skeleton under `cmd/processor` and `internal/*` packages
- Simple unit test and Makefile targets

Quick start (local dev):

1. Start infrastructure:

```bash
cd /Users/namanjain.3009/Documents/Parsec
make up
```

2. Build and run processor locally:

```bash
make build
./bin/processor
```

3. Run tests:

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Or use the Thunder Client test suite
./thunderclient/run_tests.sh
```

## 🎯 What's Implemented

### ✅ Log Ingestion Pipeline
- **HTTP Ingest Handler** (`/ingest`)
  - Single & batch log ingestion
  - JSON validation with detailed error messages
  - Field normalization (timestamps, severity, source)
  - Partial success handling (207 Multi-Status)
  - API key authentication

- **Async Worker Pool**
  - Configurable worker count
  - Channel-based message queue
  - Batch processing with timeout
  - Graceful shutdown with proper cleanup

- **Kafka Producer**
  - Connection pooling
  - Exponential backoff retry
  - Snappy compression
  - Per-partition publishing

### 🔍 Observability
- **Structured Logging** (Zerolog)
  - JSON-formatted logs with timestamps
  - Request ID tracking across pipeline
  - Component-based log hierarchy
  - Configurable log levels (DEBUG, INFO, WARN, ERROR)

- **Prometheus Metrics** (`/metrics`)
  - HTTP request duration & status codes
  - Ingest events (by tenant, severity, status)
  - Worker processing metrics
  - Kafka publish success/failure rates
  - Queue depth & active requests

- **Panic Recovery**
  - HTTP middleware with stack traces
  - Worker goroutine recovery
  - Metrics for panic events

### 📊 Monitoring Endpoints
- **`/health`** - Health check
- **`/stats`** - Runtime statistics (goroutines, memory, queue depth)
- **`/metrics`** - Prometheus metrics

### 🧪 Test Suite
- **Unit Tests** - 6 test packages covering all components
- **Integration Tests** - End-to-end pipeline testing
- **Thunder Client Collection** - 14 API test cases
  - Valid requests (single & batch)
  - Invalid payloads & validation
  - Authentication tests
  - Health & monitoring endpoints
- See `thunderclient/` directory for complete test documentation

## 🔧 Configuration

Environment variables:

```bash
# Application
export PORT=8080
export LOG_LEVEL=info  # debug, info, warn, error

# Kafka
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC=logs

# Worker Pool
export WORKER_COUNT=5
export BATCH_SIZE=100
export BATCH_TIMEOUT=1s

# Queue
export QUEUE_SIZE=10000
```

## What to add next:
- Implement Kafka consumers and commit/seek semantics
- Add ClickHouse/Postgres connectors and schema migration
- Add Redis stateful logic and checkpoint persistence
- Implement alerting rules and an alerts delivery subsystem

See `Makefile` and `docker-compose.yml` for local bootstrapping.
