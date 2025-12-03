package storage

import (
	"context"
	"errors"
)

// NewClickHouse returns a stub ClickHouse implementation (placeholder).
func NewClickHouse(dsn string) (Aggregator, error) {
	// TODO: implement real ClickHouse client
	return &stubStorage{}, errors.New("clickhouse client not implemented in scaffold")
}

type stubStorage struct{}

func (s *stubStorage) Persist(ctx context.Context, key string, payload []byte) error { return nil }
func (s *stubStorage) Close() error                                                  { return nil }
