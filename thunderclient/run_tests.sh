#!/bin/bash

# Parsec Thunder Client Test Runner
# This script helps you run tests using cURL as an alternative to Thunder Client

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
API_KEY="${API_KEY:-test-api-key-123}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Parsec Log Ingest API Test Suite${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Counter for passed/failed tests
PASSED=0
FAILED=0

# Helper function to run a test
run_test() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local payload="$4"
    local expected_code="$5"
    
    echo -ne "${YELLOW}Running:${NC} $test_name ... "
    
    if [ -z "$payload" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "$BASE_URL$endpoint" \
            -H "X-API-Key: $API_KEY" 2>&1)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "X-API-Key: $API_KEY" \
            -d "$payload" 2>&1)
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" == "$expected_code" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC} (Expected: $expected_code, Got: $http_code)"
        echo -e "  Response: $body"
        ((FAILED++))
        return 1
    fi
}

echo -e "${BLUE}[1/4] Valid Requests${NC}\n"

# Test 1: Single Log - INFO
run_test "Ingest Single Log (INFO)" "POST" "/ingest" \
'{
  "id": "log-001",
  "tenant_id": "tenant-acme",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INFO",
  "source": "api-gateway",
  "message": "User authentication successful",
  "metadata": {
    "user_id": "user-12345"
  }
}' "200"

# Test 2: Single Log - ERROR
run_test "Ingest Single Log (ERROR)" "POST" "/ingest" \
'{
  "id": "log-002",
  "tenant_id": "tenant-acme",
  "timestamp": "2024-12-07T10:35:00Z",
  "severity": "ERROR",
  "source": "payment-service",
  "message": "Payment processing failed"
}' "200"

# Test 3: Batch Logs
run_test "Ingest Batch Logs (5 logs)" "POST" "/ingest" \
'[
  {"id":"log-101","tenant_id":"tenant-test","timestamp":"2024-12-07T11:00:00Z","severity":"DEBUG","source":"test","message":"Log 1"},
  {"id":"log-102","tenant_id":"tenant-test","timestamp":"2024-12-07T11:01:00Z","severity":"INFO","source":"test","message":"Log 2"},
  {"id":"log-103","tenant_id":"tenant-test","timestamp":"2024-12-07T11:02:00Z","severity":"WARN","source":"test","message":"Log 3"},
  {"id":"log-104","tenant_id":"tenant-test","timestamp":"2024-12-07T11:03:00Z","severity":"ERROR","source":"test","message":"Log 4"},
  {"id":"log-105","tenant_id":"tenant-test","timestamp":"2024-12-07T11:04:00Z","severity":"CRITICAL","source":"test","message":"Log 5"}
]' "200"

echo ""
echo -e "${BLUE}[2/4] Invalid Payloads${NC}\n"

# Test 4: Missing tenant_id
run_test "Missing Required Field (tenant_id)" "POST" "/ingest" \
'{
  "id": "log-invalid-001",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INFO",
  "source": "test",
  "message": "Missing tenant_id"
}' "400"

# Test 5: Invalid severity
run_test "Invalid Severity Level" "POST" "/ingest" \
'{
  "id": "log-invalid-002",
  "tenant_id": "tenant-test",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INVALID_LEVEL",
  "source": "test",
  "message": "Invalid severity"
}' "400"

# Test 6: Empty message
run_test "Empty Message" "POST" "/ingest" \
'{
  "id": "log-invalid-003",
  "tenant_id": "tenant-test",
  "timestamp": "2024-12-07T10:30:00Z",
  "severity": "INFO",
  "source": "test",
  "message": ""
}' "400"

# Test 7: Malformed JSON
run_test "Malformed JSON" "POST" "/ingest" \
'{invalid json here' "400"

# Test 8: Partial success
run_test "Mixed Valid/Invalid (Partial Success)" "POST" "/ingest" \
'[
  {"id":"log-301","tenant_id":"tenant-test","timestamp":"2024-12-07T13:00:00Z","severity":"INFO","source":"test","message":"Valid 1"},
  {"id":"log-302","timestamp":"2024-12-07T13:01:00Z","severity":"INFO","source":"test","message":"Missing tenant_id"},
  {"id":"log-303","tenant_id":"tenant-test","timestamp":"2024-12-07T13:02:00Z","severity":"INFO","source":"test","message":"Valid 2"}
]' "207"

echo ""
echo -e "${BLUE}[3/4] Authentication Tests${NC}\n"

# Test 9: Missing API Key
echo -ne "${YELLOW}Running:${NC} Missing API Key ... "
response=$(curl -s -w "\n%{http_code}" -X POST \
    "$BASE_URL/ingest" \
    -H "Content-Type: application/json" \
    -d '{"id":"log-auth-001","tenant_id":"tenant-test","timestamp":"2024-12-07T14:00:00Z","severity":"INFO","source":"test","message":"No API key"}' 2>&1)
http_code=$(echo "$response" | tail -n1)
if [ "$http_code" == "401" ]; then
    echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
    ((PASSED++))
else
    echo -e "${RED}✗ FAILED${NC} (Expected: 401, Got: $http_code)"
    ((FAILED++))
fi

# Test 10: Invalid API Key
echo -ne "${YELLOW}Running:${NC} Invalid API Key ... "
response=$(curl -s -w "\n%{http_code}" -X POST \
    "$BASE_URL/ingest" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: invalid-key-xyz" \
    -d '{"id":"log-auth-002","tenant_id":"tenant-test","timestamp":"2024-12-07T14:01:00Z","severity":"INFO","source":"test","message":"Invalid API key"}' 2>&1)
http_code=$(echo "$response" | tail -n1)
if [ "$http_code" == "401" ]; then
    echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
    ((PASSED++))
else
    echo -e "${RED}✗ FAILED${NC} (Expected: 401, Got: $http_code)"
    ((FAILED++))
fi

echo ""
echo -e "${BLUE}[4/4] Health & Monitoring${NC}\n"

# Test 11: Health Check
run_test "Health Check" "GET" "/health" "" "200"

# Test 12: Runtime Stats
run_test "Runtime Stats" "GET" "/stats" "" "200"

# Test 13: Prometheus Metrics
run_test "Prometheus Metrics" "GET" "/metrics" "" "200"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Test Results${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo -e "${BLUE}Total:${NC}  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC} 🎉"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
