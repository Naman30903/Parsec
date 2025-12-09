# 🎯 Thunder Client Test Suite - Quick Reference

## 📦 What's Included

```
thunderclient/
├── thunder-collection_parsec-log-ingest.json  # Thunder Client collection
├── sample_payloads.md                          # Comprehensive payload examples
├── run_tests.sh                                # Automated cURL test runner
└── README.md                                   # Full documentation
```

## 🚀 Quick Start

### Option 1: Thunder Client (VS Code)
1. Install Thunder Client extension
2. Import `thunder-collection_parsec-log-ingest.json`
3. Click "Run All" to execute 14 tests

### Option 2: Automated Script
```bash
# Start infrastructure and processor first
make up && make build && ./bin/processor

# In another terminal, run tests
./thunderclient/run_tests.sh
```

### Option 3: Individual cURL Commands
```bash
# Single log
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{"id":"log-001","tenant_id":"tenant-test","timestamp":"2024-12-07T10:30:00Z","severity":"INFO","source":"test","message":"Hello"}'

# Batch logs
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '[
    {"id":"log-001","tenant_id":"tenant-test","timestamp":"2024-12-07T10:30:00Z","severity":"INFO","source":"test","message":"Log 1"},
    {"id":"log-002","tenant_id":"tenant-test","timestamp":"2024-12-07T10:31:00Z","severity":"WARN","source":"test","message":"Log 2"}
  ]'
```

## 📋 Test Cases Summary

| # | Test Name | Method | Expected | What It Tests |
|---|-----------|--------|----------|---------------|
| 1 | Single Log (INFO) | POST /ingest | 200 | Basic ingestion |
| 2 | Single Log (ERROR) | POST /ingest | 200 | Error severity |
| 3 | Batch (5 logs) | POST /ingest | 200 | Batch processing |
| 4 | Batch (10 logs) | POST /ingest | 200 | Larger batch |
| 5 | Missing tenant_id | POST /ingest | 400 | Required field validation |
| 6 | Invalid severity | POST /ingest | 400 | Enum validation |
| 7 | Empty message | POST /ingest | 400 | Non-empty validation |
| 8 | Malformed JSON | POST /ingest | 400 | JSON parsing |
| 9 | Mixed valid/invalid | POST /ingest | 207 | Partial success |
| 10 | Missing API key | POST /ingest | 401 | Auth required |
| 11 | Invalid API key | POST /ingest | 401 | Auth validation |
| 12 | Health check | GET /health | 200 | Health endpoint |
| 13 | Runtime stats | GET /stats | 200 | Stats endpoint |
| 14 | Prometheus metrics | GET /metrics | 200 | Metrics endpoint |

## 🎨 Test Output Example

```
========================================
  Parsec Log Ingest API Test Suite
========================================

[1/4] Valid Requests

Running: Ingest Single Log (INFO) ... ✓ PASSED (HTTP 200)
Running: Ingest Single Log (ERROR) ... ✓ PASSED (HTTP 200)
Running: Ingest Batch Logs (5 logs) ... ✓ PASSED (HTTP 200)

[2/4] Invalid Payloads

Running: Missing Required Field (tenant_id) ... ✓ PASSED (HTTP 400)
Running: Invalid Severity Level ... ✓ PASSED (HTTP 400)
Running: Empty Message ... ✓ PASSED (HTTP 400)

[3/4] Authentication Tests

Running: Missing API Key ... ✓ PASSED (HTTP 401)
Running: Invalid API Key ... ✓ PASSED (HTTP 401)

[4/4] Health & Monitoring

Running: Health Check ... ✓ PASSED (HTTP 200)
Running: Runtime Stats ... ✓ PASSED (HTTP 200)
Running: Prometheus Metrics ... ✓ PASSED (HTTP 200)

========================================
  Test Results
========================================
Passed: 14
Failed: 0
Total:  14

✓ All tests passed! 🎉
```

## 📊 Coverage Matrix

| Category | Tests | Coverage |
|----------|-------|----------|
| **Functional** | 9 | Single log, batch, validation, partial success |
| **Security** | 2 | Missing/invalid API keys |
| **Observability** | 3 | Health, stats, metrics endpoints |
| **Total** | **14** | **Complete API coverage** |

## 🎯 Sample Payloads

### Valid Single Log
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
    "ip_address": "192.168.1.100"
  }
}
```

### Valid Batch
```json
[
  {
    "id": "log-101",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T11:00:00Z",
    "severity": "DEBUG",
    "source": "service-a",
    "message": "Processing started"
  },
  {
    "id": "log-102",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T11:01:00Z",
    "severity": "INFO",
    "source": "service-a",
    "message": "Processing completed"
  }
]
```

### Invalid - Missing tenant_id
```json
{
  "id": "log-001",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INFO",
  "source": "test",
  "message": "Missing tenant_id field"
}
```
**Expected:** `400 Bad Request` with validation error

### Invalid - Bad Severity
```json
{
  "id": "log-001",
  "tenant_id": "tenant-test",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INVALID_LEVEL",
  "source": "test",
  "message": "Invalid severity"
}
```
**Expected:** `400 Bad Request` with severity validation error

### Partial Success
```json
[
  {
    "id": "log-301",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T13:00:00Z",
    "severity": "INFO",
    "source": "test",
    "message": "Valid log 1"
  },
  {
    "id": "log-302",
    "timestamp": "2024-12-07T13:01:00Z",
    "severity": "INFO",
    "source": "test",
    "message": "Missing tenant_id - INVALID"
  },
  {
    "id": "log-303",
    "tenant_id": "tenant-test",
    "timestamp": "2024-12-07T13:02:00Z",
    "severity": "INFO",
    "source": "test",
    "message": "Valid log 2"
  }
]
```
**Expected:** `207 Multi-Status` with `accepted: 2, rejected: 1`

## 🔍 Response Examples

### Success (200 OK)
```json
{
  "status": "success",
  "accepted": 5,
  "rejected": 0,
  "batch_id": "batch-uuid-here"
}
```

### Partial Success (207 Multi-Status)
```json
{
  "status": "partial_success",
  "accepted": 2,
  "rejected": 1,
  "errors": [
    {
      "index": 1,
      "error": "validation failed: tenant_id is required"
    }
  ],
  "batch_id": "batch-uuid-here"
}
```

### Validation Error (400 Bad Request)
```json
{
  "error": "validation failed: tenant_id is required"
}
```

### Auth Error (401 Unauthorized)
```json
{
  "error": "unauthorized: missing or invalid API key"
}
```

## 🛠️ Testing Workflow

### 1. Pre-Test Setup
```bash
# Terminal 1: Start infrastructure
make up

# Terminal 2: Build and run processor
make build
LOG_LEVEL=debug ./bin/processor
```

### 2. Run Tests
```bash
# Terminal 3: Run test suite
./thunderclient/run_tests.sh

# Or run individual test with cURL
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d @thunderclient/samples/single-log.json
```

### 3. Monitor Results
```bash
# Watch metrics
curl -s http://localhost:8080/metrics | grep ingest_events_total

# Check stats
curl -s http://localhost:8080/stats | jq

# View logs (in processor terminal)
# Look for JSON structured logs with request_id tracking
```

## 📈 Performance Benchmarks

Expected performance (on local machine):
- **Single log latency:** < 50ms
- **Batch throughput:** 1000+ logs/sec
- **Concurrent requests:** 100+ simultaneous
- **Queue capacity:** 10,000 events

Test with:
```bash
# Apache Bench
ab -n 1000 -c 10 -T application/json -H "X-API-Key: test-api-key-123" \
   -p payload.json http://localhost:8080/ingest

# Or use existing test script
./test_pipeline.sh
```

## ✅ Validation Checklist

Before considering tests complete:
- [ ] All 14 Thunder Client tests pass
- [ ] `run_tests.sh` script passes (14/14)
- [ ] Response times acceptable (< 100ms)
- [ ] No errors in processor logs
- [ ] Kafka receives messages (check with Kafka consumer)
- [ ] Metrics increment properly
- [ ] Health check returns healthy status
- [ ] Stats show reasonable queue depth
- [ ] Panic recovery works (no crashes)
- [ ] Structured logs appear correctly

## 🐛 Common Issues

### Connection Refused
```
Error: Failed to connect to localhost port 8080
```
**Fix:** Start processor: `./bin/processor`

### 401 Unauthorized
```
{"error": "unauthorized: missing or invalid API key"}
```
**Fix:** Add header: `-H "X-API-Key: test-api-key-123"`

### Kafka Error in Logs
```
ERROR Failed to publish to Kafka error="connection refused"
```
**Fix:** Start infrastructure: `make up`

### Queue Full
```
{"error": "queue full, try again later"}
```
**Fix:** Increase workers or queue size in config

## 📚 Additional Resources

- **Full Documentation:** `thunderclient/README.md`
- **All Sample Payloads:** `thunderclient/sample_payloads.md`
- **Integration Tests:** `test/integration_test/`
- **Pipeline Overview:** `PIPELINE.md`

---

**Ready to test? Run:** `./thunderclient/run_tests.sh` 🚀
