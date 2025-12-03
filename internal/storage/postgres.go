package storage

import "errors"

// NewPostgres returns a stub Postgres implementation (placeholder).
func NewPostgres(dsn string) (Aggregator, error) {
	// TODO: implement real Postgres client with partitioned tables
	return nil, errors.New("postgres client not implemented in scaffold")
}
