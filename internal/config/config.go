package config

// Config holds runtime configuration for the processor.
type Config struct {
	// Kafka brokers (CSV or list in future)
	KafkaBrokers string
	// Storage backend: clickhouse or postgres
	StorageBackend string
	// Redis address
	RedisAddr string
}

// Default returns a sensible default config for local dev.
func Default() *Config {
	return &Config{
		KafkaBrokers:   "localhost:9092",
		StorageBackend: "clickhouse",
		RedisAddr:      "localhost:6379",
	}
}
