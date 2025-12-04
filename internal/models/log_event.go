package models

import (
	"errors"
	"time"
)

// Severity represents log severity levels
type Severity string

const (
	SeverityDebug    Severity = "DEBUG"
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityError    Severity = "ERROR"
	SeverityCritical Severity = "CRITICAL"
)

// LogEvent represents a single log entry from a tenant's system
type LogEvent struct {
	// Unique identifier for the log event
	ID string `json:"id"`

	// Tenant identifier for multi-tenant support
	TenantID string `json:"tenant_id"`

	// Timestamp when the log was generated
	Timestamp time.Time `json:"timestamp"`

	// Severity level of the log
	Severity Severity `json:"severity"`

	// Source service or application that generated the log
	Source string `json:"source"`

	// Log message content
	Message string `json:"message"`

	// Optional structured metadata
	Metadata map[string]string `json:"metadata,omitempty"`

	// Optional trace ID for distributed tracing
	TraceID string `json:"trace_id,omitempty"`

	// Optional span ID for distributed tracing
	SpanID string `json:"span_id,omitempty"`
}

// Validation errors
var (
	ErrEmptyID          = errors.New("log event ID cannot be empty")
	ErrEmptyTenantID    = errors.New("tenant ID cannot be empty")
	ErrZeroTimestamp    = errors.New("timestamp cannot be zero")
	ErrFutureTimestamp  = errors.New("timestamp cannot be in the future")
	ErrInvalidTimestamp = errors.New("invalid timestamp format")
	ErrInvalidSeverity  = errors.New("invalid severity level")
	ErrEmptySource      = errors.New("source cannot be empty")
	ErrEmptyMessage     = errors.New("message cannot be empty")
	ErrMessageTooLong   = errors.New("message exceeds maximum length")
	ErrTooManyMetadata  = errors.New("too many metadata keys")
)

const (
	MaxMessageLength = 65536 // 64KB max message size
	MaxMetadataKeys  = 50
)

// Validate checks if the LogEvent has all required fields and valid values
func (e *LogEvent) Validate() error {
	if e.ID == "" {
		return ErrEmptyID
	}

	if e.TenantID == "" {
		return ErrEmptyTenantID
	}

	if e.Timestamp.IsZero() {
		return ErrZeroTimestamp
	}

	if e.Timestamp.After(time.Now().Add(time.Minute)) {
		return ErrFutureTimestamp
	}

	if !e.Severity.IsValid() {
		return ErrInvalidSeverity
	}

	if e.Source == "" {
		return ErrEmptySource
	}

	if e.Message == "" {
		return ErrEmptyMessage
	}

	if len(e.Message) > MaxMessageLength {
		return ErrMessageTooLong
	}

	if len(e.Metadata) > MaxMetadataKeys {
		return ErrTooManyMetadata
	}

	return nil
}

// IsValid checks if the severity level is valid
func (s Severity) IsValid() bool {
	switch s {
	case SeverityDebug, SeverityInfo, SeverityWarning, SeverityError, SeverityCritical:
		return true
	default:
		return false
	}
}
