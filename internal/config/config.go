package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration for the processor.
type Config struct {
	// Kafka configuration
	Kafka KafkaConfig

	// Storage backend: clickhouse or postgres
	StorageBackend string

	// Redis address
	RedisAddr string
}

// KafkaConfig holds Kafka-specific configuration
type KafkaConfig struct {
	// Brokers is a comma-separated list of Kafka broker addresses
	Brokers []string

	// Topic for log events
	Topic string

	// Producer settings
	Producer ProducerConfig

	// Consumer settings
	Consumer ConsumerConfig
}

// ProducerConfig holds Kafka producer settings
type ProducerConfig struct {
	// BatchSize is the number of messages to batch before sending
	BatchSize int

	// BatchTimeout is the max time to wait before sending a batch
	BatchTimeout time.Duration

	// MaxRetries is the number of retries for failed sends
	MaxRetries int

	// RetryBackoff is the initial backoff between retries
	RetryBackoff time.Duration

	// RequiredAcks: 0=none, 1=leader, -1=all
	RequiredAcks int

	// Compression: none, gzip, snappy, lz4, zstd
	Compression string

	// MaxMessageBytes is the max size of a single message
	MaxMessageBytes int

	// WriteTimeout is the timeout for write operations
	WriteTimeout time.Duration

	// PoolSize is the number of concurrent writers
	PoolSize int
}

// ConsumerConfig holds Kafka consumer settings
type ConsumerConfig struct {
	// GroupID is the consumer group ID
	GroupID string

	// MinBytes is the minimum batch size
	MinBytes int

	// MaxBytes is the maximum batch size
	MaxBytes int

	// MaxWait is the max time to wait for new data
	MaxWait time.Duration
}

// Default returns a sensible default config for local dev.
func Default() *Config {
	return &Config{
		Kafka: KafkaConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "log-events",
			Producer: ProducerConfig{
				BatchSize:       100,
				BatchTimeout:    100 * time.Millisecond,
				MaxRetries:      3,
				RetryBackoff:    100 * time.Millisecond,
				RequiredAcks:    -1, // wait for all replicas
				Compression:     "snappy",
				MaxMessageBytes: 1024 * 1024, // 1MB
				WriteTimeout:    10 * time.Second,
				PoolSize:        4,
			},
			Consumer: ConsumerConfig{
				GroupID:  "parsec-processor",
				MinBytes: 10e3, // 10KB
				MaxBytes: 10e6, // 10MB
				MaxWait:  time.Second,
			},
		},
		StorageBackend: "clickhouse",
		RedisAddr:      "localhost:6379",
	}
}

// FromEnv loads configuration from environment variables
func FromEnv() *Config {
	cfg := Default()

	// Kafka brokers
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Kafka.Brokers = strings.Split(brokers, ",")
	}

	// Kafka topic
	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		cfg.Kafka.Topic = topic
	}

	// Producer settings
	if batchSize := os.Getenv("KAFKA_BATCH_SIZE"); batchSize != "" {
		if v, err := strconv.Atoi(batchSize); err == nil {
			cfg.Kafka.Producer.BatchSize = v
		}
	}

	if batchTimeout := os.Getenv("KAFKA_BATCH_TIMEOUT_MS"); batchTimeout != "" {
		if v, err := strconv.Atoi(batchTimeout); err == nil {
			cfg.Kafka.Producer.BatchTimeout = time.Duration(v) * time.Millisecond
		}
	}

	if maxRetries := os.Getenv("KAFKA_MAX_RETRIES"); maxRetries != "" {
		if v, err := strconv.Atoi(maxRetries); err == nil {
			cfg.Kafka.Producer.MaxRetries = v
		}
	}

	if poolSize := os.Getenv("KAFKA_POOL_SIZE"); poolSize != "" {
		if v, err := strconv.Atoi(poolSize); err == nil {
			cfg.Kafka.Producer.PoolSize = v
		}
	}

	if compression := os.Getenv("KAFKA_COMPRESSION"); compression != "" {
		cfg.Kafka.Producer.Compression = compression
	}

	// Consumer settings
	if groupID := os.Getenv("KAFKA_CONSUMER_GROUP"); groupID != "" {
		cfg.Kafka.Consumer.GroupID = groupID
	}

	// Storage backend
	if backend := os.Getenv("STORAGE_BACKEND"); backend != "" {
		cfg.StorageBackend = backend
	}

	// Redis
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		cfg.RedisAddr = redisAddr
	}

	return cfg
}
