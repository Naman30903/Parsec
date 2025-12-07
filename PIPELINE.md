# End-to-End Pipeline Setup

## Quick Start

### 1. Start Infrastructure

```bash
make up
```

This starts:
- Kafka (localhost:9092)
- Zookeeper
- Redis
- ClickHouse/PostgreSQL

### 2. Build and Run Processor

```bash
make build
./bin/processor
```

Or with custom config:

```bash
KAFKA_BROKERS=localhost:9092 \
KAFKA_TOPIC=log-events \
KAFKA_POOL_SIZE=4 \
./bin/processor
```

### 3. Test Ingestion

**Single Event:**

```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt-123",
    "tenant_id": "tenant-1",
    "timestamp": "2024-01-15T10:30:00Z",
    "severity": "INFO",
    "source": "api-gateway",
    "message": "Request processed successfully"
  }'
```

**Batch Events:**

```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "events": [
      {
        "id": "evt-1",
        "tenant_id": "tenant-1",
        "timestamp": "2024-01-15T10:30:00Z",
        "severity": "INFO",
        "source": "service-a",
        "message": "Event 1"
      },
      {
        "id": "evt-2",
        "tenant_id": "tenant-1",
        "timestamp": "2024-01-15T10:30:01Z",
        "severity": "ERROR",
        "source": "service-b",
        "message": "Event 2"
      }
    ]
  }'
```

### 4. Check Health & Stats

**Health Check:**

```bash
curl http://localhost:8080/health
```

**Statistics:**

```bash
curl http://localhost:8080/stats
```

## Pipeline Architecture

```
HTTP Request → Ingest Handler → Envelope Channel → Worker Pool → Kafka Producer → Kafka
                    ↓                                    ↓
                Validation                          Batching + Retry
                Normalization                       
```

## Components

### 1. Ingest Handler (`/internal/handlers/ingest.go`)
- Accepts single events or batches
- Validates and normalizes log events
- Pushes envelopes to channel (non-blocking)
- Returns partial success for batch errors

### 2. Worker Pool (`/internal/worker/worker.go`)
- N concurrent workers (configurable)
- Batches events for efficient Kafka publishing
- Timeout-based flushing
- Graceful shutdown with flush

### 3. Kafka Producer (`/internal/kafka/producer.go`)
- Connection pooling
- Exponential backoff retry
- Compression (snappy/gzip/lz4/zstd)
- Partitioning by tenant ID

### 4. Processor (`/internal/processor/processor.go`)
- Orchestrates all components
- HTTP server lifecycle
- Graceful shutdown sequence
- Metrics reporting

## Configuration

Environment variables:

```bash
# Kafka
KAFKA_BROKERS=localhost:9092,broker2:9092
KAFKA_TOPIC=log-events
KAFKA_BATCH_SIZE=100
KAFKA_BATCH_TIMEOUT_MS=100
KAFKA_MAX_RETRIES=3
KAFKA_POOL_SIZE=4
KAFKA_COMPRESSION=snappy
KAFKA_CONSUMER_GROUP=parsec-processor

# Storage
STORAGE_BACKEND=clickhouse

# Redis
REDIS_ADDR=localhost:6379
```

## Graceful Shutdown

The processor handles shutdown in this order:

1. **Stop HTTP Server** - No new requests accepted
2. **Close Envelope Channel** - No new envelopes
3. **Flush Worker Pool** - Process remaining envelopes
4. **Close Kafka Producer** - Flush remaining messages
5. **Wait for Goroutines** - Clean exit

Press `Ctrl+C` or send `SIGTERM` for graceful shutdown.

## Testing

```bash
# Unit tests
make test

# Integration tests (requires Kafka)
make up
make test-integration

# Run with verbose output
make test-verbose
```

## Monitoring

The processor logs:
- Startup/shutdown events
- HTTP request handling
- Worker pool activity
- Kafka publish success/failures
- Periodic statistics (every 30s)

Stats are available at `/stats` endpoint:

```json
{
  "worker": {
    "processed": 1000,
    "failed": 5
  },
  "producer": {
    "messages_sent": 995,
    "messages_failed": 5,
    "bytes_written": 524288
  },
  "channel": {
    "buffered": 10,
    "capacity": 1000
  }
}
```

## Error Handling

### Validation Errors
- Returned immediately to client
- Event not enqueued
- Client gets detailed error info

### Publishing Errors
- Exponential backoff retry (3 attempts)
- Fallback to individual publish
- Logged for monitoring
- Counted in failure metrics

### Channel Full
- Non-blocking send with immediate rejection
- Client gets "queue full" error
- Backpressure mechanism

## Performance Tuning

**High Throughput:**
- Increase `KAFKA_POOL_SIZE` (more workers)
- Increase `KAFKA_BATCH_SIZE` (larger batches)
- Increase envelope channel buffer size

**Low Latency:**
- Decrease `KAFKA_BATCH_TIMEOUT_MS`
- Decrease `KAFKA_BATCH_SIZE`
- Use `compression=none`

**Reliability:**
- Set `KAFKA_MAX_RETRIES=5+`
- Increase `WriteTimeout`
- Use `RequiredAcks=-1` (all replicas)
