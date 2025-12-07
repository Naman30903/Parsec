package worker

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"parsec/internal/models"
)

// Publisher defines the interface for publishing envelopes
type Publisher interface {
	Publish(ctx context.Context, envelope *models.Envelope) error
	PublishBatch(ctx context.Context, envelopes []*models.Envelope) error
}

// Pool manages a pool of workers that consume envelopes and publish to Kafka
type Pool struct {
	publisher    Publisher
	envelopeChan chan *models.Envelope
	workers      int
	batchSize    int
	batchTimeout time.Duration

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	// Metrics
	processed atomic.Uint64
	failed    atomic.Uint64
}

// Config holds worker pool configuration
type Config struct {
	Publisher    Publisher
	EnvelopeChan chan *models.Envelope
	Workers      int
	BatchSize    int
	BatchTimeout time.Duration
}

// NewPool creates a new worker pool
func NewPool(cfg Config) *Pool {
	if cfg.Workers <= 0 {
		cfg.Workers = 4
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	if cfg.BatchTimeout <= 0 {
		cfg.BatchTimeout = 100 * time.Millisecond
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		publisher:    cfg.Publisher,
		envelopeChan: cfg.EnvelopeChan,
		workers:      cfg.Workers,
		batchSize:    cfg.BatchSize,
		batchTimeout: cfg.BatchTimeout,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start begins processing envelopes
func (p *Pool) Start() {
	log.Printf("starting %d workers with batch size %d", p.workers, p.batchSize)

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Stop gracefully stops all workers
func (p *Pool) Stop() {
	log.Println("stopping worker pool...")
	p.cancel()
	p.wg.Wait()
	log.Println("worker pool stopped")
}

// worker processes envelopes from the channel
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	log.Printf("worker %d started", id)
	defer log.Printf("worker %d stopped", id)

	batch := make([]*models.Envelope, 0, p.batchSize)
	timer := time.NewTimer(p.batchTimeout)
	defer timer.Stop()

	for {
		select {
		case <-p.ctx.Done():
			// Flush remaining batch before exiting
			if len(batch) > 0 {
				p.publishBatch(batch)
			}
			return

		case envelope, ok := <-p.envelopeChan:
			if !ok {
				// Channel closed, flush and exit
				if len(batch) > 0 {
					p.publishBatch(batch)
				}
				return
			}

			batch = append(batch, envelope)

			// Publish when batch is full
			if len(batch) >= p.batchSize {
				p.publishBatch(batch)
				batch = batch[:0] // Reset batch
				timer.Reset(p.batchTimeout)
			}

		case <-timer.C:
			// Publish on timeout if we have any messages
			if len(batch) > 0 {
				p.publishBatch(batch)
				batch = batch[:0]
			}
			timer.Reset(p.batchTimeout)
		}
	}
}

// publishBatch publishes a batch of envelopes
func (p *Pool) publishBatch(batch []*models.Envelope) {
	if len(batch) == 0 {
		return
	}

	// Create a timeout context for the publish operation
	ctx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
	defer cancel()

	err := p.publisher.PublishBatch(ctx, batch)
	if err != nil {
		log.Printf("failed to publish batch of %d: %v", len(batch), err)
		p.failed.Add(uint64(len(batch)))

		// Fallback: try publishing individually
		p.publishIndividually(batch)
	} else {
		p.processed.Add(uint64(len(batch)))
	}
}

// publishIndividually tries to publish each envelope separately (fallback)
func (p *Pool) publishIndividually(batch []*models.Envelope) {
	log.Printf("attempting individual publish for %d envelopes", len(batch))

	for _, envelope := range batch {
		ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
		err := p.publisher.Publish(ctx, envelope)
		cancel()

		if err != nil {
			log.Printf("failed to publish envelope %s: %v", envelope.Event.ID, err)
		} else {
			// Don't count twice - subtract from failed, add to processed
			p.failed.Add(^uint64(0)) // Subtract 1
			p.processed.Add(1)
		}
	}
}

// Stats returns worker pool statistics
func (p *Pool) Stats() Stats {
	return Stats{
		Processed: p.processed.Load(),
		Failed:    p.failed.Load(),
	}
}

// Stats holds worker pool metrics
type Stats struct {
	Processed uint64
	Failed    uint64
}
