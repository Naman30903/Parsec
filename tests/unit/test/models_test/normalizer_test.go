package models_test

import (
	"testing"
	"time"

	"parsec/internal/models"
)

func TestLogEventNormalize(t *testing.T) {
	e := &models.LogEvent{
		ID:        "  evt-123  ",
		TenantID:  "  tenant-1  ",
		Timestamp: time.Now(),
		Severity:  "info",
		Source:    "  API-Gateway  ",
		Message:   "  Request processed  ",
		Metadata: map[string]string{
			"  KEY  ": "  value  ",
		},
		TraceID: "  trace-123  ",
		SpanID:  "  span-456  ",
	}

	e.Normalize()

	if e.ID != "evt-123" {
		t.Errorf("ID not trimmed: got %q", e.ID)
	}
	if e.TenantID != "tenant-1" {
		t.Errorf("TenantID not trimmed: got %q", e.TenantID)
	}
	if e.Source != "api-gateway" {
		t.Errorf("Source not normalized: got %q", e.Source)
	}
	if e.Message != "Request processed" {
		t.Errorf("Message not trimmed: got %q", e.Message)
	}
	if e.Severity != models.SeverityInfo {
		t.Errorf("Severity not normalized: got %q", e.Severity)
	}
	if e.TraceID != "trace-123" {
		t.Errorf("TraceID not trimmed: got %q", e.TraceID)
	}
	if e.SpanID != "span-456" {
		t.Errorf("SpanID not trimmed: got %q", e.SpanID)
	}
	if val, ok := e.Metadata["key"]; !ok || val != "value" {
		t.Errorf("Metadata not normalized: got %v", e.Metadata)
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"RFC3339", "2024-01-15T10:30:00Z", false},
		{"RFC3339Nano", "2024-01-15T10:30:00.123456789Z", false},
		{"datetime with T", "2024-01-15T10:30:00", false},
		{"datetime with space", "2024-01-15 10:30:00", false},
		{"with whitespace", "  2024-01-15T10:30:00Z  ", false},
		{"invalid", "not-a-timestamp", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := models.ParseTimestamp(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimestamp(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParseTimestampReturnsUTC(t *testing.T) {
	ts, err := models.ParseTimestamp("2024-01-15T10:30:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.Location() != time.UTC {
		t.Errorf("expected UTC timezone, got %v", ts.Location())
	}
}
