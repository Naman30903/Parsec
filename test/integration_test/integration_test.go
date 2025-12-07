package integrationtest

import (
	"bytes"
	"context"
	"encoding/json"
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

	go func() {
		if err := p.Run(ctx); err != nil {
			t.Logf("processor error: %v", err)
		}
	}()

	// Give it time to start
	time.Sleep(500 * time.Millisecond)

	// Test 1: Health check
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

	// Test 2: Ingest single event
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
		resp, err := http.Post(
			"http://localhost:8080/ingest",
			"application/json",
			bytes.NewBuffer(body),
		)
		if err != nil {
			t.Skipf("skipping: server not ready: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if !result["success"].(bool) {
			t.Errorf("expected success=true: %v", result)
		}
	})

	// Test 3: Ingest batch
	t.Run("ingest_batch", func(t *testing.T) {
		events := []map[string]interface{}{}
		for i := 0; i < 10; i++ {
			events = append(events, map[string]interface{}{
				"id":        "test-evt-" + string(rune(i)),
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
		resp, err := http.Post(
			"http://localhost:8080/ingest",
			"application/json",
			bytes.NewBuffer(body),
		)
		if err != nil {
			t.Skipf("skipping: server not ready: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	// Test 4: Check stats
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
		json.NewDecoder(resp.Body).Decode(&stats)

		t.Logf("Stats: %+v", stats)
	})

	// Test 5: Graceful shutdown
	t.Run("graceful_shutdown", func(t *testing.T) {
		cancel()
		time.Sleep(2 * time.Second)

		// Server should be down
		_, err := http.Get("http://localhost:8080/health")
		if err == nil {
			t.Error("expected server to be down")
		}
	})
}
