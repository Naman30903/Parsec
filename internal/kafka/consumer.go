package kafka

import "context"

// Consumer is a lightweight interface representing a Kafka consumer.
type Consumer interface {
    Start(ctx context.Context) error
    Stop() error
}

// NewStubConsumer returns a stub consumer for the scaffold.
func NewStubConsumer(brokers string) Consumer {
    return &stubConsumer{brokers: brokers}
}

type stubConsumer struct{
    brokers string
}

func (s *stubConsumer) Start(ctx context.Context) error {
    // TODO: wire real kafka client
    <-ctx.Done()
    return nil
}

func (s *stubConsumer) Stop() error { return nil }
