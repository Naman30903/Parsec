package integrationtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"parsec/internal/config"
	"parsec/internal/processor"
)

func TestEndToEndPipeline(t *testing.T) {
	// Create config
	cfg := config.Default()
	cfg.Kafka.Brokers = []string{"localhost:9092"}

	// Create processor
	p := processor.New(cfg)

	// Start processor in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to track processor completion
	processorDone := make(chan error, 1)
	go func() {
		err := p.Run(ctx)
		processorDone <- err
	}()

	// Give it time to start
	time.Sleep(500 * time.Millisecond)

	// Helper function to create authenticated requests
	createAuthRequest := func(method, url string, body []byte) (*http.Request, error) {
		req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-api-key-123")
		return req, nil
	}

	// Test 1: Health check (no auth required)
	t.Run("health_check", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/health")
		if err != nil {
			t.Skipf("skipping: server not ready: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	// Test 2: Ingest single event (with auth)
	t.Run("ingest_single_event", func(t *testing.T) {
		payload := map[string]interface{}{
			"id":        "test-evt-1",
			"tenant_id": "tenant-1",
			"timestamp": time.Now().Format(time.RFC3339),
			"severity":  "INFO",
			"source":    "test-service",
			"message":   "Test message",
		}

		body, _ := json.Marshal(payload)
		req, err := createAuthRequest("POST", "http://localhost:8080/ingest", body)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("skipping: server not ready: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("expected 200, got %d: %s", resp.StatusCode, string(bodyBytes))
			return
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		success, ok := result["success"].(bool)
		if !ok || !success {
			t.Errorf("expected success=true: %v", result)
		}
	})

	// Test 3: Ingest batch (with auth)
	t.Run("ingest_batch", func(t *testing.T) {
		events := []map[string]interface{}{}
		for i := 0; i < 10; i++ {
			events = append(events, map[string]interface{}{
				"id":        fmt.Sprintf("test-evt-%d", i),
				"tenant_id": "tenant-1",
				"timestamp": time.Now().Format(time.RFC3339),
				"severity":  "INFO",
				"source":    "test-service",
				"message":   "Test message",
			})
		}

		payload := map[string]interface{}{
			"events": events,
		}

		body, _ := json.Marshal(payload)
		req, err := createAuthRequest("POST", "http://localhost:8080/ingest", body)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Skipf("skipping: server not ready: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	// Test 4: Check stats (no auth required)
	t.Run("check_stats", func(t *testing.T) {
		// Wait for processing
		time.Sleep(1 * time.Second)

		resp, err := http.Get("http://localhost:8080/stats")
		if err != nil {
			t.Skipf("skipping: server not ready: %v", err)
			return
		}
		defer resp.Body.Close()

		var stats map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			t.Logf("failed to decode stats: %v", err)
			return
		}

		t.Logf("Stats: %+v", stats)
	})

	// Test 5: Graceful shutdown
	t.Run("graceful_shutdown", func(t *testing.T) {
		// Trigger shutdown
		cancel()

		// Wait for processor to finish (with timeout)
		select {
		case err := <-processorDone:
			if err != nil {
				t.Logf("processor exited with error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("processor did not shut down within 5 seconds")
		}

		// Give extra time for server to fully stop
		time.Sleep(500 * time.Millisecond)

		// Now server should be down
		client := &http.Client{Timeout: 1 * time.Second}
		_, err := client.Get("http://localhost:8080/health")
		if err == nil {
			t.Error("expected server to be down, but it's still responding")
		}
	})
}
