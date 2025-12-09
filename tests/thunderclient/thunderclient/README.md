# Thunderclient Test Suite for Parsec Log Ingest API

This directory contains comprehensive test cases for the Parsec log ingestion API using Thunder Client (VS Code extension).

## 📁 Files

- **`thunder-collection_parsec-log-ingest.json`** - Thunder Client collection with 14 test requests
- **`sample_payloads.md`** - Comprehensive documentation of all sample payloads
- **`README.md`** - This file

## 🚀 Quick Start

### Prerequisites
1. Install [Thunder Client](https://marketplace.visualstudio.com/items?itemName=rangav.vscode-thunder-client) VS Code extension
2. Start the Parsec infrastructure: `make up`
3. Build and run the processor: `make build && ./bin/processor`

### Import Collection
1. Open VS Code
2. Click on Thunder Client icon in the sidebar
3. Click on "Collections" tab
4. Click on menu (⋮) → "Import"
5. Select `thunder-collection_parsec-log-ingest.json`

## 📋 Test Collection Overview

### Folder Structure
The collection is organized into 4 folders:

#### 1️⃣ Valid Requests (5 tests)
- ✅ Ingest Single Log - INFO
- ✅ Ingest Single Log - ERROR
- ✅ Ingest Batch Logs - Mixed Severity (5 logs)
- ✅ Ingest Batch - Large Payload (10 logs)
- ✅ Ingest Batch - E-commerce Workflow

#### 2️⃣ Invalid Payloads (5 tests)
- ❌ Missing Required Field (tenant_id)
- ❌ Invalid Severity Level
- ❌ Empty Message
- ❌ Malformed JSON
- ❌ Batch with Mixed Valid/Invalid (Partial Success)

#### 3️⃣ Auth Tests (2 tests)
- 🔐 Missing API Key → 401
- 🔐 Invalid API Key → 401

#### 4️⃣ Health & Stats (3 tests)
- 💚 Health Check
- 📊 Runtime Stats
- 📈 Prometheus Metrics

## 🧪 Running Tests

### Run All Tests
1. Open Thunder Client
2. Go to Collections tab
3. Click on "Parsec Log Ingest API" collection
4. Click "Run All" button

### Run Individual Tests
1. Navigate to the specific request
2. Click "Send" button
3. View response and test results

### Run by Folder
1. Click on folder name (e.g., "Valid Requests")
2. Click "Run All" for that folder

## 📊 Expected Results

### Success Criteria
| Test Category | Expected Results |
|---------------|------------------|
| Valid Single Logs | ✅ 200 OK, `accepted: 1` |
| Valid Batch Logs | ✅ 200 OK, `accepted: N` (where N = batch size) |
| Invalid Payloads | ❌ 400 Bad Request with error details |
| Partial Success | ⚠️ 207 Multi-Status, `accepted: X, rejected: Y` |
| Missing API Key | 🔒 401 Unauthorized |
| Invalid API Key | 🔒 401 Unauthorized |
| Health Check | 💚 200 OK, `status: healthy` |
| Stats Endpoint | 📊 200 OK, JSON with runtime stats |
| Metrics Endpoint | 📈 200 OK, Prometheus metrics text |

## 🎯 Test Coverage

### Functional Coverage
- ✅ Single log ingestion
- ✅ Batch log ingestion (5, 10 logs)
- ✅ All severity levels (DEBUG, INFO, WARN, ERROR, CRITICAL)
- ✅ Validation rules (required fields, format checks)
- ✅ Partial success handling
- ✅ Error responses

### Security Coverage
- ✅ API key authentication
- ✅ Missing credentials handling
- ✅ Invalid credentials handling

### Observability Coverage
- ✅ Health check endpoint
- ✅ Runtime statistics
- ✅ Prometheus metrics

## 🔧 Customization

### Modify API Key
Default API key: `test-api-key-123`

To change:
1. Update the `X-API-Key` header in each request
2. Or set up Thunder Client environment variable:
   - Create environment: "Parsec Dev"
   - Add variable: `api_key` = `your-key`
   - Use in headers: `{{api_key}}`

### Modify Base URL
Default: `http://localhost:8080`

To change:
1. Use environment variable `{{base_url}}`
2. Or find/replace in the JSON file

### Add Custom Tests
1. Create new request in Thunder Client
2. Set URL, method, headers, body
3. Add tests in "Tests" tab:
   ```javascript
   // Example tests
   json.accepted > 0
   res.status == 200
   res.body contains "success"
   ```

## 📝 Sample Payloads

For detailed payload examples, see [`sample_payloads.md`](./sample_payloads.md)

Quick examples:

### Single Log
```json
{
  "id": "log-001",
  "tenant_id": "tenant-acme",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INFO",
  "source": "api-gateway",
  "message": "User authentication successful",
  "metadata": {
    "user_id": "user-12345"
  }
}
```

### Batch Logs
```json
[
  { "id": "log-001", ... },
  { "id": "log-002", ... }
]
```

## 🐛 Troubleshooting

### Connection Refused
**Problem:** `ECONNREFUSED localhost:8080`
**Solution:** 
- Start the processor: `./bin/processor`
- Check if port 8080 is available: `lsof -i :8080`

### 401 Unauthorized
**Problem:** All requests return 401
**Solution:**
- Check if API key header is included: `X-API-Key: test-api-key-123`
- Verify auth middleware is configured correctly

### Kafka Errors in Logs
**Problem:** "Failed to publish to Kafka"
**Solution:**
- Start infrastructure: `make up`
- Check Kafka is running: `docker ps | grep kafka`
- Verify KAFKA_BROKERS env var: `echo $KAFKA_BROKERS`

### Empty Response
**Problem:** Request hangs or returns empty
**Solution:**
- Check processor logs for panics
- Verify queue isn't full (check `/stats` endpoint)
- Increase worker count: `export WORKER_COUNT=10`

## 📈 Performance Testing

### Using Thunder Client
For basic performance testing:
1. Use "Repeat Request" feature (N times)
2. Monitor response times in Thunder Client UI

### Using External Tools
For serious load testing, use:

```bash
# Apache Bench
ab -n 1000 -c 10 \
   -H "Content-Type: application/json" \
   -H "X-API-Key: test-api-key-123" \
   -p payload.json \
   http://localhost:8080/ingest

# Or use the provided test script
./test_pipeline.sh
```

## 🔍 Monitoring During Tests

### Watch Logs
```bash
# Terminal 1: Run processor with debug logging
LOG_LEVEL=debug ./bin/processor

# Terminal 2: Run tests in Thunder Client
```

### Check Metrics
```bash
# Watch metrics change in real-time
watch -n 1 'curl -s http://localhost:8080/metrics | grep -E "ingest_events_total|worker_processed_total"'
```

### Check Stats
```bash
# Monitor queue depth and goroutines
watch -n 1 'curl -s http://localhost:8080/stats | jq'
```

## 📚 Additional Resources

- **Full API Documentation:** See `../README.md`
- **Integration Tests:** See `../test/integration_test/`
- **Pipeline Test Script:** See `../test_pipeline.sh`
- **Architecture:** See `../PIPELINE.md`

## ✅ Test Checklist

Before deployment, verify:
- [ ] All 14 tests pass
- [ ] Response times < 100ms for single log
- [ ] Batch processing handles 1000+ logs
- [ ] Validation errors are descriptive
- [ ] Auth properly blocks unauthorized requests
- [ ] Health check returns 200
- [ ] Metrics endpoint is accessible
- [ ] Structured logs appear in console
- [ ] No panic recovery logged (unless testing panic scenarios)
- [ ] Queue depth stays below 80% (check `/stats`)

## 🎉 Success Criteria

Your test suite is ready when:
- ✅ All valid requests return 200/207
- ✅ All invalid requests return 400
- ✅ All auth tests return 401
- ✅ Health/Stats/Metrics all return 200
- ✅ Response times are acceptable
- ✅ No errors in processor logs
- ✅ Kafka receives all accepted messages

---

**Happy Testing! 🚀**
