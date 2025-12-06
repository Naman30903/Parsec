package kafka_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"parsec/internal/config"
	"parsec/internal/kafka"
	"parsec/internal/models"
)

// skipIfNoKafka skips the test if Kafka is not available
func skipIfNoKafka(t *testing.T) {
	if os.Getenv("KAFKA_TEST") != "1" {
		t.Skip("Skipping Kafka integration test. Set KAFKA_TEST=1 to run.")
	}
}

func TestProducerPublish(t *testing.T) {
	skipIfNoKafka(t)

	cfg := config.Default()
	producer, err := kafka.NewProducer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.Producer,
	)
	if err != nil {
		t.Fatalf("failed to create producer: %v", err)
	}
	defer producer.Close()

	// Create test envelope
	event := &models.LogEvent{
		ID:        "test-evt-1",
		TenantID:  "tenant-1",
		Timestamp: time.Now(),
		Severity:  models.SeverityInfo,
		Source:    "test-service",
		Message:   "Test message",
	}
	envelope := models.NewEnvelope(event, "test-node")

	// Publish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = producer.Publish(ctx, envelope)
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	// Check stats
	stats := producer.Stats()
	if stats.MessagesSent != 1 {
		t.Errorf("expected 1 message sent, got %d", stats.MessagesSent)
	}
}

func TestProducerPublishBatch(t *testing.T) {
	skipIfNoKafka(t)

	cfg := config.Default()
	producer, err := kafka.NewProducer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.Producer,
	)
	if err != nil {
		t.Fatalf("failed to create producer: %v", err)
	}
	defer producer.Close()

	// Create test envelopes
	envelopes := make([]*models.Envelope, 10)
	for i := 0; i < 10; i++ {
		event := &models.LogEvent{
			ID:        fmt.Sprintf("test-evt-%d", i),
			TenantID:  "tenant-1",
			Timestamp: time.Now(),
			Severity:  models.SeverityInfo,
			Source:    "test-service",
			Message:   fmt.Sprintf("Test message %d", i),
		}
		envelopes[i] = models.NewEnvelope(event, "test-node")
	}

	// Publish batch
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = producer.PublishBatch(ctx, envelopes)
	if err != nil {
		t.Fatalf("failed to publish batch: %v", err)
	}

	// Check stats
	stats := producer.Stats()
	if stats.MessagesSent != 10 {
		t.Errorf("expected 10 messages sent, got %d", stats.MessagesSent)
	}
}

func TestProducerClose(t *testing.T) {
	skipIfNoKafka(t)

	cfg := config.Default()
	producer, err := kafka.NewProducer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.Producer,
	)
	if err != nil {
		t.Fatalf("failed to create producer: %v", err)
	}

	// Close producer
	err = producer.Close()
	if err != nil {
		t.Errorf("failed to close producer: %v", err)
	}

	// Try to publish after close
	event := &models.LogEvent{
		ID:        "test-evt",
		TenantID:  "tenant-1",
		Timestamp: time.Now(),
		Severity:  models.SeverityInfo,
		Source:    "test-service",
		Message:   "Test",
	}
	envelope := models.NewEnvelope(event, "test-node")

	ctx := context.Background()
	err = producer.Publish(ctx, envelope)
	if err != kafka.ErrProducerClosed {
		t.Errorf("expected ErrProducerClosed, got %v", err)
	}
}
