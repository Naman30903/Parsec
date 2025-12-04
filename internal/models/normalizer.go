package models

import (
	"strings"
	"time"
)

// SupportedTimestampFormats lists formats we attempt to parse
var SupportedTimestampFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	time.RFC1123,
	time.UnixDate,
}

// Normalize applies field normalization to a LogEvent
// - lower-cases Source
// - trims Message
// - ensures Timestamp is valid
func (e *LogEvent) Normalize() {
	// Lower-case the source/service name
	e.Source = strings.ToLower(strings.TrimSpace(e.Source))

	// Trim whitespace from message
	e.Message = strings.TrimSpace(e.Message)

	// Trim ID and TenantID
	e.ID = strings.TrimSpace(e.ID)
	e.TenantID = strings.TrimSpace(e.TenantID)

	// Normalize severity to uppercase
	e.Severity = Severity(strings.ToUpper(strings.TrimSpace(string(e.Severity))))

	// Trim trace/span IDs
	e.TraceID = strings.TrimSpace(e.TraceID)
	e.SpanID = strings.TrimSpace(e.SpanID)

	// Normalize metadata keys to lowercase
	if e.Metadata != nil {
		normalized := make(map[string]string, len(e.Metadata))
		for k, v := range e.Metadata {
			normalized[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
		}
		e.Metadata = normalized
	}
}

// ParseTimestamp attempts to parse a timestamp string into time.Time
func ParseTimestamp(ts string) (time.Time, error) {
	ts = strings.TrimSpace(ts)

	for _, format := range SupportedTimestampFormats {
		if t, err := time.Parse(format, ts); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, ErrInvalidTimestamp
}
