# Parsec Log Ingest - Sample Payloads

This document contains sample payloads for testing the Parsec log ingestion API.

## Table of Contents
- [Valid Payloads](#valid-payloads)
- [Invalid Payloads](#invalid-payloads)
- [Edge Cases](#edge-cases)
- [Performance Testing](#performance-testing)

---

## Valid Payloads

### 1. Single Log - INFO Level
```json
{
  "id": "log-001",
  "tenant_id": "tenant-acme",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INFO",
  "source": "api-gateway",
  "message": "User authentication successful",
  "metadata": {
    "user_id": "user-12345",
    "ip_address": "192.168.1.100",
    "endpoint": "/api/v1/auth",
    "response_time_ms": 45
  }
}
```

### 2. Single Log - ERROR Level
```json
{
  "id": "log-002",
  "tenant_id": "tenant-acme",
  "timestamp": "2024-12-07T10:35:00Z",
  "severity": "ERROR",
  "source": "payment-service",
  "message": "Payment processing failed: Insufficient funds",
  "metadata": {
    "transaction_id": "txn-98765",
    "user_id": "user-54321",
    "amount": 150.00,
    "currency": "USD",
    "error_code": "INSUFFICIENT_FUNDS"
  }
}
```

### 3. Single Log - CRITICAL Level
```json
{
  "id": "log-003",
  "tenant_id": "tenant-techcorp",
  "timestamp": "2024-12-07T10:40:00Z",
  "severity": "CRITICAL",
  "source": "database-primary",
  "message": "Database connection pool exhausted",
  "metadata": {
    "pool_size": 100,
    "active_connections": 100,
    "waiting_connections": 250,
    "alert_id": "alert-db-001"
  }
}
```

### 4. Batch Logs - Mixed Severity
```json
[
  {
    "id": "log-101",
    "tenant_id": "tenant-techcorp",
    "timestamp": "2024-12-07T11:00:00Z",
    "severity": "DEBUG",
    "source": "data-pipeline",
    "message": "Processing batch 1234",
    "metadata": {
      "batch_id": "batch-1234",
      "record_count": 1000
    }
  },
  {
    "id": "log-102",
    "tenant_id": "tenant-techcorp",
    "timestamp": "2024-12-07T11:01:00Z",
    "severity": "INFO",
    "source": "data-pipeline",
    "message": "Batch processing completed successfully",
    "metadata": {
      "batch_id": "batch-1234",
      "duration_ms": 5420
    }
  },
  {
    "id": "log-103",
    "tenant_id": "tenant-techcorp",
    "timestamp": "2024-12-07T11:05:00Z",
    "severity": "WARN",
    "source": "cache-service",
    "message": "Cache miss rate exceeding threshold",
    "metadata": {
      "miss_rate": 0.35,
      "threshold": 0.30,
      "cache_size_mb": 512
    }
  },
  {
    "id": "log-104",
    "tenant_id": "tenant-techcorp",
    "timestamp": "2024-12-07T11:10:00Z",
    "severity": "ERROR",
    "source": "database-replica",
    "message": "Replication lag detected",
    "metadata": {
      "lag_seconds": 45,
      "replica_id": "replica-03",
      "master_id": "master-01"
    }
  },
  {
    "id": "log-105",
    "tenant_id": "tenant-techcorp",
    "timestamp": "2024-12-07T11:15:00Z",
    "severity": "CRITICAL",
    "source": "database-replica",
    "message": "Replication stopped - manual intervention required",
    "metadata": {
      "replica_id": "replica-03",
      "error": "Connection lost to master"
    }
  }
]
```

### 5. E-commerce Workflow Batch
```json
[
  {
    "id": "log-201",
    "tenant_id": "tenant-retail",
    "timestamp": "2024-12-07T12:00:00Z",
    "severity": "INFO",
    "source": "checkout-service",
    "message": "Order placed successfully",
    "metadata": {
      "order_id": "ord-1001",
      "amount": 99.99,
      "items": 3
    }
  },
  {
    "id": "log-202",
    "tenant_id": "tenant-retail",
    "timestamp": "2024-12-07T12:01:00Z",
    "severity": "INFO",
    "source": "inventory-service",
    "message": "Stock updated",
    "metadata": {
      "sku": "PROD-5678",
      "quantity": -1,
      "remaining": 245
    }
  },
  {
    "id": "log-203",
    "tenant_id": "tenant-retail",
    "timestamp": "2024-12-07T12:02:00Z",
    "severity": "INFO",
    "source": "notification-service",
    "message": "Order confirmation sent",
    "metadata": {
      "order_id": "ord-1001",
      "email": "customer@example.com",
      "channel": "email"
    }
  }
]
```

### 6. Microservices Distributed Trace
```json
[
  {
    "id": "log-trace-001",
    "tenant_id": "tenant-fintech",
    "timestamp": "2024-12-07T13:00:00.000Z",
    "severity": "INFO",
    "source": "api-gateway",
    "message": "Request received",
    "metadata": {
      "trace_id": "trace-abc-123",
      "span_id": "span-001",
      "method": "POST",
      "path": "/api/v1/transfer"
    }
  },
  {
    "id": "log-trace-002",
    "tenant_id": "tenant-fintech",
    "timestamp": "2024-12-07T13:00:00.050Z",
    "severity": "INFO",
    "source": "auth-service",
    "message": "Token validated",
    "metadata": {
      "trace_id": "trace-abc-123",
      "span_id": "span-002",
      "parent_span": "span-001",
      "user_id": "user-999"
    }
  },
  {
    "id": "log-trace-003",
    "tenant_id": "tenant-fintech",
    "timestamp": "2024-12-07T13:00:00.150Z",
    "severity": "INFO",
    "source": "account-service",
    "message": "Account balance checked",
    "metadata": {
      "trace_id": "trace-abc-123",
      "span_id": "span-003",
      "parent_span": "span-002",
      "account_id": "acc-5555",
      "balance": 1500.00
    }
  },
  {
    "id": "log-trace-004",
    "tenant_id": "tenant-fintech",
    "timestamp": "2024-12-07T13:00:00.300Z",
    "severity": "INFO",
    "source": "transaction-service",
    "message": "Transfer completed",
    "metadata": {
      "trace_id": "trace-abc-123",
      "span_id": "span-004",
      "parent_span": "span-003",
      "transaction_id": "txn-777",
      "amount": 250.00
    }
  }
]
```

---

## Invalid Payloads

### 1. Missing Required Field - tenant_id
```json
{
  "id": "log-invalid-001",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INFO",
  "source": "test-service",
  "message": "This log is missing tenant_id"
}
```
**Expected Response:** `400 Bad Request` with validation error mentioning `tenant_id`

### 2. Missing Required Field - message
```json
{
  "id": "log-invalid-002",
  "tenant_id": "tenant-test",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INFO",
  "source": "test-service"
}
```
**Expected Response:** `400 Bad Request` with validation error mentioning `message`

### 3. Invalid Severity Level
```json
{
  "id": "log-invalid-003",
  "tenant_id": "tenant-test",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INVALID_LEVEL",
  "source": "test-service",
  "message": "This has an invalid severity"
}
```
**Expected Response:** `400 Bad Request` with validation error about severity

### 4. Empty Message
```json
{
  "id": "log-invalid-004",
  "tenant_id": "tenant-test",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INFO",
  "source": "test-service",
  "message": ""
}
```
**Expected Response:** `400 Bad Request` with validation error about empty message

### 5. Invalid Timestamp Format
```json
{
  "id": "log-invalid-005",
  "tenant_id": "tenant-test",
  "timestamp": "not-a-valid-timestamp",
  "severity": "INFO",
  "source": "test-service",
  "message": "Invalid timestamp format"
}
```
**Expected Response:** `400 Bad Request` with timestamp parsing error

### 6. Malformed JSON
```
{id": "log-001", "tenant_id": "tenant-test", invalid json here
```
**Expected Response:** `400 Bad Request` with JSON parsing error

### 7. Empty Array
```json
[]
```
**Expected Response:** `400 Bad Request` with "empty batch" error

### 8. Batch with All Invalid Logs
```json
[
  {
    "id": "log-invalid-101",
    "timestamp": "2024-12-07T10:30:00Z",
    "severity": "INFO",
    "source": "test-service",
    "message": "Missing tenant_id"
  },
  {
    "id": "log-invalid-102",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T10:31:00Z",
    "severity": "INVALID",
    "source": "test-service",
    "message": "Invalid severity"
  }
]
```
**Expected Response:** `400 Bad Request` with multiple validation errors

---

## Edge Cases

### 1. Partial Success - Mixed Valid/Invalid in Batch
```json
[
  {
    "id": "log-301",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T13:00:00Z",
    "severity": "INFO",
    "source": "test-service",
    "message": "Valid log entry"
  },
  {
    "id": "log-302",
    "timestamp": "2024-12-07T13:01:00Z",
    "severity": "INFO",
    "source": "test-service",
    "message": "Missing tenant_id - should fail"
  },
  {
    "id": "log-303",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T13:02:00Z",
    "severity": "INFO",
    "source": "test-service",
    "message": "Another valid log entry"
  }
]
```
**Expected Response:** `207 Multi-Status` with `accepted: 2, rejected: 1`

### 2. Maximum Metadata Size
```json
{
  "id": "log-large-001",
  "tenant_id": "tenant-test",
  "timestamp": "2024-12-07T14:00:00Z",
  "severity": "INFO",
  "source": "test-service",
  "message": "Log with large metadata",
  "metadata": {
    "key1": "value1",
    "key2": "value2",
    "large_data": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
    "nested_object": {
      "level1": {
        "level2": {
          "level3": {
            "data": "deeply nested"
          }
        }
      }
    }
  }
}
```

### 3. Unicode and Special Characters
```json
{
  "id": "log-unicode-001",
  "tenant_id": "tenant-test",
  "timestamp": "2024-12-07T15:00:00Z",
  "severity": "INFO",
  "source": "internationalization-service",
  "message": "用户登录成功 🎉 Привет مرحبا",
  "metadata": {
    "user_name": "José María",
    "location": "São Paulo",
    "emoji": "✅ ❌ ⚠️"
  }
}
```

### 4. Flexible Timestamp Formats
```json
[
  {
    "id": "log-ts-001",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T15:30:00Z",
    "severity": "INFO",
    "source": "test",
    "message": "RFC3339 format"
  },
  {
    "id": "log-ts-002",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07 15:30:00",
    "severity": "INFO",
    "source": "test",
    "message": "DateTime format"
  },
  {
    "id": "log-ts-003",
    "tenant_id": "tenant-test",
    "timestamp": "1701961800",
    "severity": "INFO",
    "source": "test",
    "message": "Unix timestamp"
  }
]
```

### 5. Null Metadata
```json
{
  "id": "log-null-001",
  "tenant_id": "tenant-test",
  "timestamp": "2024-12-07T16:00:00Z",
  "severity": "INFO",
  "source": "test-service",
  "message": "Log without metadata",
  "metadata": null
}
```

---

## Performance Testing

### Load Test Batch (100 logs)
Generate 100 logs programmatically:
```bash
# Use jq or similar tool to generate
for i in {1..100}; do
  echo "{
    \"id\": \"load-test-$i\",
    \"tenant_id\": \"tenant-loadtest\",
    \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
    \"severity\": \"INFO\",
    \"source\": \"load-generator\",
    \"message\": \"Load test message $i\",
    \"metadata\": {
      \"iteration\": $i,
      \"batch_id\": \"batch-001\"
    }
  }"
done | jq -s '.'
```

### Stress Test - Concurrent Requests
Use the provided `test_pipeline.sh` script or tools like:
- Apache Bench: `ab -n 1000 -c 10 -T application/json -p payload.json http://localhost:8080/ingest`
- k6: For sophisticated load testing scenarios
- wrk: For HTTP benchmarking

---

## Testing Checklist

### Functional Tests
- [ ] Single log ingestion (all severity levels)
- [ ] Batch log ingestion (5-10 logs)
- [ ] Large batch ingestion (100+ logs)
- [ ] All validation rules (missing fields, invalid values)
- [ ] Partial success scenarios
- [ ] Malformed JSON handling

### Auth Tests
- [ ] Missing API key (401)
- [ ] Invalid API key (401)
- [ ] Valid API key (200)

### Edge Cases
- [ ] Empty message
- [ ] Empty batch array
- [ ] Unicode characters
- [ ] Large metadata objects
- [ ] Null/missing metadata
- [ ] Various timestamp formats

### Performance Tests
- [ ] Single log latency (< 50ms)
- [ ] Batch throughput (1000+ logs/sec)
- [ ] Concurrent requests (10+ simultaneous)
- [ ] Queue handling under load

### Monitoring
- [ ] Check `/health` endpoint
- [ ] Check `/stats` endpoint
- [ ] Check `/metrics` endpoint (Prometheus)
- [ ] Verify structured logs in console
- [ ] Verify metrics increments

---

## cURL Examples

### Valid Single Log
```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{
    "id": "log-001",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T10:30:00Z",
    "severity": "INFO",
    "source": "curl-test",
    "message": "Test log from cURL"
  }'
```

### Batch Logs
```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '[
    {"id": "log-001", "tenant_id": "tenant-test", "timestamp": "2024-12-07T10:30:00Z", "severity": "INFO", "source": "curl-test", "message": "First log"},
    {"id": "log-002", "tenant_id": "tenant-test", "timestamp": "2024-12-07T10:31:00Z", "severity": "WARN", "source": "curl-test", "message": "Second log"}
  ]'
```

### Invalid Payload (Missing tenant_id)
```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{
    "id": "log-001",
    "timestamp": "2024-12-07T10:30:00Z",
    "severity": "INFO",
    "source": "curl-test",
    "message": "Missing tenant_id"
  }'
```

### Missing API Key
```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "id": "log-001",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T10:30:00Z",
    "severity": "INFO",
    "source": "curl-test",
    "message": "No API key"
  }'
```

### Health Check
```bash
curl http://localhost:8080/health
```

### Prometheus Metrics
```bash
curl http://localhost:8080/metrics
```

---

## Environment Setup

Ensure these environment variables are set:
```bash
# Application
export PORT=8080
export LOG_LEVEL=info

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

## Notes
- All timestamps should be in RFC3339 format or parseable alternatives
- Valid severity levels: DEBUG, INFO, WARN, ERROR, CRITICAL
- Metadata is optional and can contain arbitrary JSON
- Batch size is recommended to be between 1-1000 logs per request
- API keys are validated via X-API-Key header
