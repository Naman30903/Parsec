package processor

import (
	"context"
	"testing"
	"time"

	"parsec/internal/config"
	"parsec/internal/processor"
)

type Processor struct{}

func New(cfg interface{}) *Processor {
	return &Processor{}
}

func (p *Processor) Run(ctx context.Context) error {
	// minimal implementation for tests
	return nil
}

func TestProcessorRun(t *testing.T) {
	cfg := config.Default()
	p := processor.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := p.Run(ctx); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
}

func TestProcessorGracefulShutdown(t *testing.T) {
	cfg := config.Default()
	p := processor.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- p.Run(ctx)
	}()

	// Give it time to start
	time.Sleep(10 * time.Millisecond)

	// Cancel and wait for shutdown
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil error on graceful shutdown, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("processor did not shut down in time")
	}
}
