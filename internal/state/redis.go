package state

import "context"

// StateStore is a minimal interface for ephemeral state (e.g., windows, counters)
type StateStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Close() error
}

type noopStore struct{}

func NewNoopStore(addr string) StateStore { return &noopStore{} }

func (n *noopStore) Get(ctx context.Context, key string) ([]byte, error)     { return nil, nil }
func (n *noopStore) Set(ctx context.Context, key string, value []byte) error { return nil }
func (n *noopStore) Close() error                                            { return nil }
