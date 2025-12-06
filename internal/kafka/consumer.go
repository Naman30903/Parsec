package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/segmentio/kafka-go"

	"parsec/internal/config"
	"parsec/internal/models"
)

// MessageHandler processes consumed messages
type MessageHandler func(ctx context.Context, envelope *models.Envelope) error

// Consumer is a Kafka consumer for processing log events
type Consumer struct {
	reader  *kafka.Reader
	handler MessageHandler
	cfg     config.ConsumerConfig
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, topic string, cfg config.ConsumerConfig, handler MessageHandler) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, errors.New("at least one broker is required")
	}

	if topic == "" {
		return nil, errors.New("topic is required")
	}

	if handler == nil {
		return nil, errors.New("handler is required")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  cfg.GroupID,
		MinBytes: cfg.MinBytes,
		MaxBytes: cfg.MaxBytes,
		MaxWait:  cfg.MaxWait,
	})

	return &Consumer{
		reader:  reader,
		handler: handler,
		cfg:     cfg,
	}, nil
}

// Start begins consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	ctx, c.cancel = context.WithCancel(ctx)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consumeLoop(ctx)
	}()

	return nil
}

// consumeLoop reads and processes messages
func (c *Consumer) consumeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("error reading message: %v", err)
			continue
		}

		// Deserialize envelope
		var envelope models.Envelope
		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			log.Printf("error deserializing message: %v", err)
			continue
		}

		// Process message
		if err := c.handler(ctx, &envelope); err != nil {
			log.Printf("error handling message %s: %v", envelope.Event.ID, err)
			// Could implement dead-letter queue here
		}
	}
}

// Stop stops the consumer
func (c *Consumer) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	return c.reader.Close()
}

// NewStubConsumer returns a stub consumer for testing
func NewStubConsumer(brokers string) *stubConsumer {
	return &stubConsumer{brokers: brokers}
}

type stubConsumer struct {
	brokers string
}

func (s *stubConsumer) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (s *stubConsumer) Stop() error { return nil }
