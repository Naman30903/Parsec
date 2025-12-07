#!/bin/bash
# Test script for the end-to-end ingest pipeline

set -e

BASE_URL="http://localhost:8080"

echo "Testing Parsec Log Ingestion Pipeline"
echo "======================================"
echo ""

# Test 1: Health Check
echo "1. Testing health endpoint..."
curl -s "$BASE_URL/health" | jq '.'
echo ""

# Test 2: Single Event
echo "2. Ingesting single event..."
curl -s -X POST "$BASE_URL/ingest" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt-001",
    "tenant_id": "demo-tenant",
    "timestamp": "2024-12-07T10:30:00Z",
    "severity": "INFO",
    "source": "api-gateway",
    "message": "User login successful",
    "metadata": {
      "user_id": "user-123",
      "ip": "192.168.1.1"
    }
  }' | jq '.'
echo ""

# Test 3: Batch Events
echo "3. Ingesting batch of events..."
curl -s -X POST "$BASE_URL/ingest" \
  -H "Content-Type: application/json" \
  -d '{
    "events": [
      {
        "id": "evt-002",
        "tenant_id": "demo-tenant",
        "timestamp": "2024-12-07T10:31:00Z",
        "severity": "WARNING",
        "source": "payment-service",
        "message": "Payment processing delayed",
        "metadata": {
          "transaction_id": "txn-456"
        }
      },
      {
        "id": "evt-003",
        "tenant_id": "demo-tenant",
        "timestamp": "2024-12-07T10:32:00Z",
        "severity": "ERROR",
        "source": "email-service",
        "message": "Failed to send notification email",
        "metadata": {
          "recipient": "user@example.com"
        }
      },
      {
        "id": "evt-004",
        "tenant_id": "demo-tenant",
        "timestamp": "2024-12-07T10:33:00Z",
        "severity": "INFO",
        "source": "cache-service",
        "message": "Cache cleared successfully"
      }
    ]
  }' | jq '.'
echo ""

# Test 4: Invalid Event (should fail validation)
echo "4. Testing validation (missing required field)..."
curl -s -X POST "$BASE_URL/ingest" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt-005",
    "tenant_id": "",
    "timestamp": "2024-12-07T10:34:00Z",
    "severity": "INFO",
    "source": "test",
    "message": "This should fail"
  }' | jq '.'
echo ""

# Test 5: Check Stats
echo "5. Checking pipeline statistics..."
sleep 1
curl -s "$BASE_URL/stats" | jq '.'
echo ""

echo "======================================"
echo "Pipeline test complete!"
