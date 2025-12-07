package worker

import (
	"context"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"parsec/internal/logger"
	"parsec/internal/metrics"
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
	log := logger.WithComponent("worker_pool")
	log.Info().
		Int("workers", p.workers).
		Int("batch_size", p.batchSize).
		Dur("batch_timeout", p.batchTimeout).
		Msg("starting worker pool")

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// Stop gracefully stops all workers
func (p *Pool) Stop() {
	log := logger.WithComponent("worker_pool")
	log.Info().Msg("stopping worker pool")
	p.cancel()
	p.wg.Wait()
	log.Info().Msg("worker pool stopped")
}

// worker processes envelopes from the channel
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	log := logger.WithComponent("worker").With().Int("worker_id", id).Logger()

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			log.Error().
				Interface("panic", r).
				Bytes("stack", stack).
				Msg("worker panic recovered")
			metrics.PanicsRecovered.WithLabelValues("worker").Inc()
		}
	}()

	log.Info().Msg("worker started")
	defer log.Info().Msg("worker stopped")

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

	log := logger.WithComponent("worker")
	start := time.Now()

	// Create a timeout context for the publish operation
	ctx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
	defer cancel()

	log.Debug().Int("batch_size", len(batch)).Msg("publishing batch to kafka")

	err := p.publisher.PublishBatch(ctx, batch)
	duration := time.Since(start)

	metrics.WorkerBatchPublishDuration.Observe(duration.Seconds())

	if err != nil {
		log.Error().
			Err(err).
			Int("batch_size", len(batch)).
			Dur("duration", duration).
			Msg("failed to publish batch")

		p.failed.Add(uint64(len(batch)))
		metrics.WorkerFailedTotal.Add(float64(len(batch)))

		// Fallback: try publishing individually
		p.publishIndividually(batch)
	} else {
		log.Info().
			Int("batch_size", len(batch)).
			Dur("duration", duration).
			Msg("batch published successfully")

		p.processed.Add(uint64(len(batch)))
		metrics.WorkerProcessedTotal.Add(float64(len(batch)))
	}
}

// publishIndividually tries to publish each envelope separately (fallback)
func (p *Pool) publishIndividually(batch []*models.Envelope) {
	log := logger.WithComponent("worker")
	log.Warn().Int("count", len(batch)).Msg("attempting individual publish for failed batch")

	for _, envelope := range batch {
		ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
		err := p.publisher.Publish(ctx, envelope)
		cancel()

		if err != nil {
			log.Error().
				Err(err).
				Str("event_id", envelope.Event.ID).
				Str("tenant_id", envelope.Event.TenantID).
				Msg("failed to publish envelope individually")
		} else {
			log.Debug().
				Str("event_id", envelope.Event.ID).
				Msg("envelope published individually")

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
