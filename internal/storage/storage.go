package storage

import "context"

// Aggregator persists aggregated metrics and supports checkpointing.
type Aggregator interface {
	Persist(ctx context.Context, key string, payload []byte) error
	Close() error
}
