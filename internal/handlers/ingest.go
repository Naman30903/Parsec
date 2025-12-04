package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"parsec/internal/models"
)

// IngestHandler handles log event ingestion via HTTP
type IngestHandler struct {
	// Channel to push envelopes to Kafka producer
	envelopeChan chan<- *models.Envelope

	// Node identifier for tracking
	nodeID string

	// Batch counter for generating batch IDs
	batchCounter uint64

	// Max body size (default 10MB)
	maxBodySize int64
}

// IngestConfig holds configuration for the ingest handler
type IngestConfig struct {
	EnvelopeChan chan<- *models.Envelope
	NodeID       string
	MaxBodySize  int64
}

// NewIngestHandler creates a new ingest handler
func NewIngestHandler(cfg IngestConfig) *IngestHandler {
	nodeID := cfg.NodeID
	if nodeID == "" {
		nodeID, _ = os.Hostname()
		if nodeID == "" {
			nodeID = "unknown"
		}
	}

	maxBodySize := cfg.MaxBodySize
	if maxBodySize == 0 {
		maxBodySize = 10 * 1024 * 1024 // 10MB default
	}

	return &IngestHandler{
		envelopeChan: cfg.EnvelopeChan,
		nodeID:       nodeID,
		maxBodySize:  maxBodySize,
	}
}

// IngestRequest represents the incoming JSON payload (single or batch)
type IngestRequest struct {
	// Single event (if Events is empty)
	Event *LogEventInput `json:"event,omitempty"`

	// Batch of events
	Events []LogEventInput `json:"events,omitempty"`
}

// LogEventInput is the input format for log events (with string timestamp)
type LogEventInput struct {
	ID        string            `json:"id"`
	TenantID  string            `json:"tenant_id"`
	Timestamp string            `json:"timestamp"` // String for flexible parsing
	Severity  string            `json:"severity"`
	Source    string            `json:"source"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	TraceID   string            `json:"trace_id,omitempty"`
	SpanID    string            `json:"span_id,omitempty"`
}

// IngestResponse is the response returned to clients
type IngestResponse struct {
	Success  bool          `json:"success"`
	Accepted int           `json:"accepted"`
	Rejected int           `json:"rejected"`
	Errors   []IngestError `json:"errors,omitempty"`
}

// IngestError describes a validation error for a specific event
type IngestError struct {
	Index   int    `json:"index"`
	EventID string `json:"event_id,omitempty"`
	Error   string `json:"error"`
}

// ServeHTTP handles the ingest HTTP request
func (h *IngestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Check content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" && contentType != "" {
		h.writeError(w, http.StatusUnsupportedMediaType, "content-type must be application/json")
		return
	}

	// Limit body size
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodySize)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}

	// Parse JSON
	events, err := h.parseBody(body)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(events) == 0 {
		h.writeError(w, http.StatusBadRequest, "no events provided")
		return
	}

	// Generate batch ID
	batchID := h.generateBatchID()

	// Process events
	response := h.processEvents(events, batchID)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if response.Rejected > 0 && response.Accepted == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(response)
}

// parseBody parses the JSON body into a slice of LogEventInput
func (h *IngestHandler) parseBody(body []byte) ([]LogEventInput, error) {
	// Try parsing as IngestRequest first
	var req IngestRequest
	if err := json.Unmarshal(body, &req); err == nil {
		if len(req.Events) > 0 {
			return req.Events, nil
		}
		if req.Event != nil {
			return []LogEventInput{*req.Event}, nil
		}
	}

	// Try parsing as array of events
	var events []LogEventInput
	if err := json.Unmarshal(body, &events); err == nil && len(events) > 0 {
		return events, nil
	}

	// Try parsing as single event
	var single LogEventInput
	if err := json.Unmarshal(body, &single); err == nil && single.ID != "" {
		return []LogEventInput{single}, nil
	}

	return nil, fmt.Errorf("invalid JSON format: expected event object or array of events")
}

// processEvents validates, normalizes, and pushes events to the channel
func (h *IngestHandler) processEvents(inputs []LogEventInput, batchID string) IngestResponse {
	response := IngestResponse{
		Success: true,
		Errors:  make([]IngestError, 0),
	}

	for i, input := range inputs {
		// Convert input to LogEvent
		event, err := h.convertInput(input)
		if err != nil {
			response.Errors = append(response.Errors, IngestError{
				Index:   i,
				EventID: input.ID,
				Error:   err.Error(),
			})
			response.Rejected++
			continue
		}

		// Normalize the event
		event.Normalize()

		// Validate the event
		if err := event.Validate(); err != nil {
			response.Errors = append(response.Errors, IngestError{
				Index:   i,
				EventID: event.ID,
				Error:   err.Error(),
			})
			response.Rejected++
			continue
		}

		// Create envelope and push to channel
		envelope := models.NewEnvelope(event, h.nodeID).WithBatch(batchID, i)

		// Non-blocking send with timeout
		select {
		case h.envelopeChan <- envelope:
			response.Accepted++
		default:
			// Channel full - reject event
			response.Errors = append(response.Errors, IngestError{
				Index:   i,
				EventID: event.ID,
				Error:   "internal queue full, try again later",
			})
			response.Rejected++
		}
	}

	response.Success = response.Rejected == 0
	return response
}

// convertInput converts LogEventInput to LogEvent
func (h *IngestHandler) convertInput(input LogEventInput) (*models.LogEvent, error) {
	// Parse timestamp
	ts, err := models.ParseTimestamp(input.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("timestamp: %w", err)
	}

	return &models.LogEvent{
		ID:        input.ID,
		TenantID:  input.TenantID,
		Timestamp: ts,
		Severity:  models.Severity(input.Severity),
		Source:    input.Source,
		Message:   input.Message,
		Metadata:  input.Metadata,
		TraceID:   input.TraceID,
		SpanID:    input.SpanID,
	}, nil
}

// generateBatchID generates a unique batch ID
func (h *IngestHandler) generateBatchID() string {
	counter := atomic.AddUint64(&h.batchCounter, 1)
	return fmt.Sprintf("%s-%d-%d", h.nodeID, time.Now().UnixNano(), counter)
}

// writeError writes an error response
func (h *IngestHandler) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
