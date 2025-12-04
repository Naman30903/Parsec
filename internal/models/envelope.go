package models

import (
	"time"
)

// Envelope wraps a LogEvent with internal metadata for processing
type Envelope struct {
	// Original event
	Event *LogEvent `json:"event"`

	// Internal processing metadata
	ReceivedAt   time.Time `json:"received_at"`
	IngestNode   string    `json:"ingest_node"`
	BatchID      string    `json:"batch_id,omitempty"`
	BatchIndex   int       `json:"batch_index,omitempty"`
	RetryCount   int       `json:"retry_count"`
	PartitionKey string    `json:"partition_key"`
}

// NewEnvelope creates a new envelope wrapping a log event
func NewEnvelope(event *LogEvent, ingestNode string) *Envelope {
	return &Envelope{
		Event:        event,
		ReceivedAt:   time.Now().UTC(),
		IngestNode:   ingestNode,
		RetryCount:   0,
		PartitionKey: event.TenantID, // partition by tenant for ordering
	}
}

// WithBatch sets batch metadata on the envelope
func (e *Envelope) WithBatch(batchID string, index int) *Envelope {
	e.BatchID = batchID
	e.BatchIndex = index
	return e
}
