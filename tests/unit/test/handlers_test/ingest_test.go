package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"parsec/internal/api"
	"parsec/internal/models"
)

func TestIngestHandler_SingleEvent(t *testing.T) {
	ch := make(chan *models.Envelope, 10)
	handler := handlers.NewIngestHandler(handlers.IngestConfig{
		EnvelopeChan: ch,
		NodeID:       "test-node",
	})

	body := `{
        "id": "evt-1",
        "tenant_id": "tenant-1",
        "timestamp": "2024-01-15T10:30:00Z",
        "severity": "info",
        "source": "API-Gateway",
        "message": "  Request processed  "
    }`

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp handlers.IngestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.Success || resp.Accepted != 1 || resp.Rejected != 0 {
		t.Errorf("unexpected response: %+v", resp)
	}

	// Check envelope was pushed
	select {
	case envelope := <-ch:
		if envelope.Event.Source != "api-gateway" {
			t.Errorf("source not normalized: got %s", envelope.Event.Source)
		}
		if envelope.Event.Message != "Request processed" {
			t.Errorf("message not trimmed: got %q", envelope.Event.Message)
		}
		if envelope.Event.Severity != models.SeverityInfo {
			t.Errorf("severity not normalized: got %s", envelope.Event.Severity)
		}
	case <-time.After(time.Second):
		t.Fatal("no envelope received")
	}
}

func TestIngestHandler_BatchEvents(t *testing.T) {
	ch := make(chan *models.Envelope, 10)
	handler := handlers.NewIngestHandler(handlers.IngestConfig{
		EnvelopeChan: ch,
		NodeID:       "test-node",
	})

	body := `{
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
    }`

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp handlers.IngestResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Accepted != 2 {
		t.Errorf("expected 2 accepted, got %d", resp.Accepted)
	}

	// Drain channel
	for i := 0; i < 2; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatal("missing envelope")
		}
	}
}

func TestIngestHandler_ValidationErrors(t *testing.T) {
	ch := make(chan *models.Envelope, 10)
	handler := handlers.NewIngestHandler(handlers.IngestConfig{
		EnvelopeChan: ch,
		NodeID:       "test-node",
	})

	body := `{
        "events": [
            {
                "id": "evt-1",
                "tenant_id": "tenant-1",
                "timestamp": "2024-01-15T10:30:00Z",
                "severity": "INFO",
                "source": "service-a",
                "message": "Valid event"
            },
            {
                "id": "",
                "tenant_id": "tenant-1",
                "timestamp": "2024-01-15T10:30:00Z",
                "severity": "INFO",
                "source": "service-a",
                "message": "Invalid - no ID"
            }
        ]
    }`

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var resp handlers.IngestResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Accepted != 1 || resp.Rejected != 1 {
		t.Errorf("expected 1 accepted, 1 rejected, got %d/%d", resp.Accepted, resp.Rejected)
	}

	if len(resp.Errors) != 1 || resp.Errors[0].Index != 1 {
		t.Errorf("expected error at index 1: %+v", resp.Errors)
	}
}

func TestIngestHandler_MethodNotAllowed(t *testing.T) {
	ch := make(chan *models.Envelope, 10)
	handler := handlers.NewIngestHandler(handlers.IngestConfig{
		EnvelopeChan: ch,
	})

	req := httptest.NewRequest(http.MethodGet, "/ingest", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
