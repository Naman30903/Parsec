package models_test

import (
	"parsec/internal/models"
	"strings"
	"testing"
	"time"
)

func TestLogEventValidate(t *testing.T) {
	validEvent := func() *models.LogEvent {
		return &models.LogEvent{
			ID:        "evt-123",
			TenantID:  "tenant-1",
			Timestamp: time.Now(),
			Severity:  models.SeverityInfo,
			Source:    "api-gateway",
			Message:   "Request processed successfully",
		}
	}

	tests := []struct {
		name    string
		modify  func(*models.LogEvent)
		wantErr error
	}{
		{"valid event", func(e *models.LogEvent) {}, nil},
		{"empty ID", func(e *models.LogEvent) { e.ID = "" }, models.ErrEmptyID},
		{"empty tenant ID", func(e *models.LogEvent) { e.TenantID = "" }, models.ErrEmptyTenantID},
		{"zero timestamp", func(e *models.LogEvent) { e.Timestamp = time.Time{} }, models.ErrZeroTimestamp},
		{"invalid severity", func(e *models.LogEvent) { e.Severity = "UNKNOWN" }, models.ErrInvalidSeverity},
		{"empty source", func(e *models.LogEvent) { e.Source = "" }, models.ErrEmptySource},
		{"empty message", func(e *models.LogEvent) { e.Message = "" }, models.ErrEmptyMessage},
		{"message too long", func(e *models.LogEvent) { e.Message = strings.Repeat("a", models.MaxMessageLength+1) }, models.ErrMessageTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := validEvent()
			tt.modify(e)
			err := e.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSeverityIsValid(t *testing.T) {
	validSeverities := []models.Severity{
		models.SeverityDebug,
		models.SeverityInfo,
		models.SeverityWarning,
		models.SeverityError,
		models.SeverityCritical,
	}
	for _, s := range validSeverities {
		if !s.IsValid() {
			t.Errorf("Severity %s should be valid", s)
		}
	}

	if models.Severity("INVALID").IsValid() {
		t.Error("Invalid severity should return false")
	}
}

func TestLogEventFutureTimestamp(t *testing.T) {
	e := &models.LogEvent{
		ID:        "evt-123",
		TenantID:  "tenant-1",
		Timestamp: time.Now().Add(time.Hour), // 1 hour in future
		Severity:  models.SeverityInfo,
		Source:    "api-gateway",
		Message:   "Test",
	}

	if err := e.Validate(); err != models.ErrFutureTimestamp {
		t.Errorf("expected ErrFutureTimestamp, got %v", err)
	}
}
