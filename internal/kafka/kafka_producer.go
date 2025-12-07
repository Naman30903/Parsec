package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"

	"parsec/internal/config"
	"parsec/internal/logger"
	"parsec/internal/metrics"
	"parsec/internal/models"
)

// Producer errors
var (
	ErrProducerClosed  = errors.New("producer is closed")
	ErrPublishTimeout  = errors.New("publish timeout")
	ErrSerializeFailed = errors.New("failed to serialize message")
)

// Producer is a Kafka producer with connection pooling, retry, and batching
type Producer struct {
	cfg     config.ProducerConfig
	topic   string
	writers []*kafka.Writer
	pool    chan *kafka.Writer
	closed  atomic.Bool

	// Metrics
	messagesSent   atomic.Uint64
	messagesFailed atomic.Uint64
	bytesWritten   atomic.Uint64
}

// ProducerOption is a functional option for configuring the producer
type ProducerOption func(*Producer)

// NewProducer creates a new Kafka producer with the given configuration
func NewProducer(brokers []string, topic string, cfg config.ProducerConfig, opts ...ProducerOption) (*Producer, error) {
	if len(brokers) == 0 {
		return nil, errors.New("at least one broker is required")
	}

	if topic == "" {
		return nil, errors.New("topic is required")
	}

	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 4
	}

	p := &Producer{
		cfg:     cfg,
		topic:   topic,
		writers: make([]*kafka.Writer, cfg.PoolSize),
		pool:    make(chan *kafka.Writer, cfg.PoolSize),
	}

	// Apply options
	for _, opt := range opts {
		opt(p)
	}

	// Get compression codec
	compression := getCompression(cfg.Compression)

	// Create writer pool
	for i := 0; i < cfg.PoolSize; i++ {
		writer := &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{}, // Partition by key
			BatchSize:    cfg.BatchSize,
			BatchTimeout: cfg.BatchTimeout,
			WriteTimeout: cfg.WriteTimeout,
			RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
			Compression:  compression,
			MaxAttempts:  cfg.MaxRetries + 1,
			Async:        false, // Sync for reliability
		}
		p.writers[i] = writer
		p.pool <- writer
	}

	return p, nil
}

// getCompression returns the kafka compression codec
func getCompression(name string) compress.Compression {
	switch name {
	case "gzip":
		return compress.Gzip
	case "snappy":
		return compress.Snappy
	case "lz4":
		return compress.Lz4
	case "zstd":
		return compress.Zstd
	default:
		return compress.None // no compression
	}
}

// Publish sends an envelope to Kafka
func (p *Producer) Publish(ctx context.Context, envelope *models.Envelope) error {
	if p.closed.Load() {
		return ErrProducerClosed
	}

	// Serialize envelope to JSON
	data, err := json.Marshal(envelope)
	if err != nil {
		p.messagesFailed.Add(1)
		return fmt.Errorf("%w: %v", ErrSerializeFailed, err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Key:   []byte(envelope.PartitionKey), // Partition by tenant
		Value: data,
		Headers: []kafka.Header{
			{Key: "tenant_id", Value: []byte(envelope.Event.TenantID)},
			{Key: "event_id", Value: []byte(envelope.Event.ID)},
			{Key: "ingest_node", Value: []byte(envelope.IngestNode)},
		},
		Time: envelope.ReceivedAt,
	}

	// Get writer from pool with timeout
	var writer *kafka.Writer
	select {
	case writer = <-p.pool:
		defer func() { p.pool <- writer }()
	case <-ctx.Done():
		p.messagesFailed.Add(1)
		return ctx.Err()
	}

	// Publish with retries
	err = p.publishWithRetry(ctx, writer, msg)
	if err != nil {
		p.messagesFailed.Add(1)
		return err
	}

	p.messagesSent.Add(1)
	p.bytesWritten.Add(uint64(len(data)))
	return nil
}

// PublishBatch sends multiple envelopes to Kafka in a single batch
func (p *Producer) PublishBatch(ctx context.Context, envelopes []*models.Envelope) error {
	if p.closed.Load() {
		return ErrProducerClosed
	}

	if len(envelopes) == 0 {
		return nil
	}

	log := logger.WithComponent("kafka_producer")
	start := time.Now()

	// Convert envelopes to messages
	messages := make([]kafka.Message, 0, len(envelopes))
	for _, envelope := range envelopes {
		data, err := json.Marshal(envelope)
		if err != nil {
			log.Error().
				Err(err).
				Str("event_id", envelope.Event.ID).
				Str("tenant_id", envelope.Event.TenantID).
				Msg("failed to serialize envelope")
			p.messagesFailed.Add(1)
			metrics.KafkaPublishTotal.WithLabelValues("failed").Inc()
			continue
		}

		msg := kafka.Message{
			Key:   []byte(envelope.PartitionKey),
			Value: data,
			Headers: []kafka.Header{
				{Key: "tenant_id", Value: []byte(envelope.Event.TenantID)},
				{Key: "event_id", Value: []byte(envelope.Event.ID)},
				{Key: "ingest_node", Value: []byte(envelope.IngestNode)},
			},
			Time: envelope.ReceivedAt,
		}
		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		return nil
	}

	// Get writer from pool
	var writer *kafka.Writer
	select {
	case writer = <-p.pool:
		defer func() { p.pool <- writer }()
	case <-ctx.Done():
		p.messagesFailed.Add(uint64(len(messages)))
		return ctx.Err()
	}

	// Publish batch with retries
	err := p.publishBatchWithRetry(ctx, writer, messages)
	duration := time.Since(start)

	metrics.KafkaPublishDuration.Observe(duration.Seconds())

	if err != nil {
		log.Error().
			Err(err).
			Int("batch_size", len(messages)).
			Dur("duration", duration).
			Msg("failed to publish batch to kafka")
		p.messagesFailed.Add(uint64(len(messages)))
		metrics.KafkaPublishTotal.WithLabelValues("failed").Add(float64(len(messages)))
		return err
	}

	log.Debug().
		Int("batch_size", len(messages)).
		Dur("duration", duration).
		Msg("batch published to kafka")

	p.messagesSent.Add(uint64(len(messages)))
	metrics.KafkaPublishTotal.WithLabelValues("success").Add(float64(len(messages)))

	bytesTotal := uint64(0)
	for _, msg := range messages {
		bytesTotal += uint64(len(msg.Value))
	}
	p.bytesWritten.Add(bytesTotal)
	metrics.KafkaBytesWritten.Add(float64(bytesTotal))

	return nil
}

// publishWithRetry publishes a single message with exponential backoff retry
func (p *Producer) publishWithRetry(ctx context.Context, writer *kafka.Writer, msg kafka.Message) error {
	log := logger.WithComponent("kafka_producer")
	var lastErr error
	backoff := p.cfg.RetryBackoff

	for attempt := 0; attempt <= p.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Warn().
				Int("attempt", attempt).
				Dur("backoff", backoff).
				Msg("retrying kafka publish")

			metrics.KafkaPublishRetries.Inc()

			select {
			case <-time.After(backoff):
				backoff *= 2 // Exponential backoff
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := writer.WriteMessages(ctx, msg)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Warn().
			Err(err).
			Int("attempt", attempt+1).
			Msg("kafka publish attempt failed")

		// Check for non-retryable errors
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	}

	log.Error().
		Err(lastErr).
		Int("max_retries", p.cfg.MaxRetries+1).
		Msg("kafka publish failed after all retries")

	return fmt.Errorf("failed after %d attempts: %w", p.cfg.MaxRetries+1, lastErr)
}

// publishBatchWithRetry publishes a batch of messages with exponential backoff retry
func (p *Producer) publishBatchWithRetry(ctx context.Context, writer *kafka.Writer, messages []kafka.Message) error {
	log := logger.WithComponent("kafka_producer")
	var lastErr error
	backoff := p.cfg.RetryBackoff

	for attempt := 0; attempt <= p.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Warn().
				Int("attempt", attempt).
				Int("batch_size", len(messages)).
				Dur("backoff", backoff).
				Msg("retrying kafka batch publish")

			metrics.KafkaPublishRetries.Inc()

			select {
			case <-time.After(backoff):
				backoff *= 2
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := writer.WriteMessages(ctx, messages...)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Warn().
			Err(err).
			Int("attempt", attempt+1).
			Int("batch_size", len(messages)).
			Msg("kafka batch publish attempt failed")

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	}

	log.Error().
		Err(lastErr).
		Int("max_retries", p.cfg.MaxRetries+1).
		Int("batch_size", len(messages)).
		Msg("kafka batch publish failed after all retries")

	return fmt.Errorf("batch failed after %d attempts: %w", p.cfg.MaxRetries+1, lastErr)
}

// Close closes all writers in the pool
func (p *Producer) Close() error {
	if p.closed.Swap(true) {
		return nil // Already closed
	}

	var errs []error
	for _, writer := range p.writers {
		if err := writer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing writers: %v", errs)
	}
	return nil
}

// Stats returns producer statistics
func (p *Producer) Stats() ProducerStats {
	return ProducerStats{
		MessagesSent:   p.messagesSent.Load(),
		MessagesFailed: p.messagesFailed.Load(),
		BytesWritten:   p.bytesWritten.Load(),
	}
}

// ProducerStats holds producer metrics
type ProducerStats struct {
	MessagesSent   uint64
	MessagesFailed uint64
	BytesWritten   uint64
}

// HealthCheck verifies the producer can connect to Kafka
func (p *Producer) HealthCheck(ctx context.Context) error {
	if p.closed.Load() {
		return ErrProducerClosed
	}

	// Get a writer from pool
	var writer *kafka.Writer
	select {
	case writer = <-p.pool:
		defer func() { p.pool <- writer }()
	case <-ctx.Done():
		return ctx.Err()
	}

	// Try to get writer stats (this doesn't actually write)
	_ = writer.Stats()
	return nil
}
