package processor

import (
	"context"
	"log"
	"sync"

	"parsec/internal/config"
)

// Processor is the high-level coordinator for consuming, processing, and alerting.
type Processor struct {
	cfg *config.Config
	// add clients here (kafka, redis, storage)
	wg sync.WaitGroup
}

// New constructs a Processor with given config.
func New(cfg *config.Config) *Processor {
	return &Processor{cfg: cfg}
}

// Run starts background goroutines and blocks until context cancelled.
func (p *Processor) Run(ctx context.Context) error {
	log.Println("processor starting")

	// Example: start a worker that does nothing useful yet
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		<-ctx.Done()
		log.Println("worker: received shutdown")
	}()

	// block until cancelled
	<-ctx.Done()

	// wait for goroutines
	p.wg.Wait()
	log.Println("processor stopped")
	return nil
}
