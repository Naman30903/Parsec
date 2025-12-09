package processor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"parsec/internal/config"
	"parsec/internal/api"
	"parsec/internal/kafka"
	"parsec/internal/logger"
	"parsec/internal/metrics"
	"parsec/internal/middleware"
	"parsec/internal/models"
	"parsec/internal/worker"
)

// Processor is the high-level coordinator for consuming, processing, and alerting.
type Processor struct {
	cfg          *config.Config
	producer     *kafka.Producer
	workerPool   *worker.Pool
	httpServer   *http.Server
	envelopeChan chan *models.Envelope
	wg           sync.WaitGroup
}

// New constructs a Processor with given config.
func New(cfg *config.Config) *Processor {
	return &Processor{
		cfg:          cfg,
		envelopeChan: make(chan *models.Envelope, 1000), // Buffer for 1000 envelopes
	}
}

// Run starts background goroutines and blocks until context cancelled.
func (p *Processor) Run(ctx context.Context) error {
	log := logger.WithComponent("processor")
	log.Info().Msg("processor starting")

	// Initialize Kafka producer
	if err := p.initProducer(); err != nil {
		log.Error().Err(err).Msg("failed to initialize producer")
		return fmt.Errorf("failed to initialize producer: %w", err)
	}
	defer p.producer.Close()

	// Initialize worker pool
	p.initWorkerPool()
	p.workerPool.Start()
	defer p.workerPool.Stop()

	// Initialize HTTP server
	if err := p.initHTTPServer(); err != nil {
		log.Error().Err(err).Msg("failed to initialize HTTP server")
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	// Start HTTP server in background
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		log.Info().Str("addr", ":8080").Msg("starting HTTP server")
		if err := p.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("HTTP server error")
		}
	}()

	// Stats reporting goroutine
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.reportStats(ctx)
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Info().Msg("shutdown signal received")

	// Graceful shutdown
	return p.shutdown()
}

// initProducer initializes the Kafka producer
func (p *Processor) initProducer() error {
	log := logger.WithComponent("processor")
	producer, err := kafka.NewProducer(
		p.cfg.Kafka.Brokers,
		p.cfg.Kafka.Topic,
		p.cfg.Kafka.Producer,
	)
	if err != nil {
		return err
	}

	p.producer = producer
	log.Info().
		Strs("brokers", p.cfg.Kafka.Brokers).
		Str("topic", p.cfg.Kafka.Topic).
		Msg("kafka producer initialized")
	return nil
}

// initWorkerPool initializes the worker pool
func (p *Processor) initWorkerPool() {
	log := logger.WithComponent("processor")
	p.workerPool = worker.NewPool(worker.Config{
		Publisher:    p.producer,
		EnvelopeChan: p.envelopeChan,
		Workers:      p.cfg.Kafka.Producer.PoolSize,
		BatchSize:    p.cfg.Kafka.Producer.BatchSize,
		BatchTimeout: p.cfg.Kafka.Producer.BatchTimeout,
	})
	log.Info().Int("workers", p.cfg.Kafka.Producer.PoolSize).Msg("worker pool initialized")
}

// initHTTPServer initializes the HTTP server with handlers
func (p *Processor) initHTTPServer() error {
	mux := http.NewServeMux()

	// Ingest handler (with middleware)
	ingestHandler := handlers.NewIngestHandler(handlers.IngestConfig{
		EnvelopeChan: p.envelopeChan,
		NodeID:       "",               // Will use hostname
		MaxBodySize:  10 * 1024 * 1024, // 10MB
	})
	mux.Handle("/ingest", middleware.Chain(
		ingestHandler,
		middleware.Recovery,
		middleware.Logging,
		middleware.Auth,
	))

	// Health check
	mux.HandleFunc("/health", p.healthHandler)

	// Stats endpoint
	mux.HandleFunc("/stats", p.statsHandler)

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Initialize queue capacity metric
	metrics.WorkerQueueCapacity.Set(float64(cap(p.envelopeChan)))

	p.httpServer = &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

// shutdown performs graceful shutdown
func (p *Processor) shutdown() error {
	log := logger.WithComponent("processor")
	log.Info().Msg("initiating graceful shutdown")

	// 1. Stop accepting new HTTP requests
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info().Msg("stopping HTTP server")
	if err := p.httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown error")
	}

	// 2. Close envelope channel to signal no more incoming envelopes
	log.Info().Msg("closing envelope channel")
	close(p.envelopeChan)

	// 3. Wait for workers to finish processing (with timeout)
	done := make(chan struct{})
	go func() {
		p.workerPool.Stop()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("workers stopped gracefully")
	case <-time.After(15 * time.Second):
		log.Warn().Msg("worker shutdown timeout - forcing exit")
	}

	// 4. Close producer
	log.Info().Msg("closing kafka producer")
	if err := p.producer.Close(); err != nil {
		log.Error().Err(err).Msg("producer close error")
	}

	// 5. Wait for all goroutines
	p.wg.Wait()

	log.Info().Msg("processor stopped gracefully")
	return nil
}

// reportStats periodically logs statistics
func (p *Processor) reportStats(ctx context.Context) {
	log := logger.WithComponent("processor")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			workerStats := p.workerPool.Stats()
			producerStats := p.producer.Stats()

			// Update metrics
			metrics.WorkerQueueSize.Set(float64(len(p.envelopeChan)))

			log.Info().
				Uint64("worker_processed", workerStats.Processed).
				Uint64("worker_failed", workerStats.Failed).
				Uint64("producer_sent", producerStats.MessagesSent).
				Uint64("producer_failed", producerStats.MessagesFailed).
				Uint64("producer_bytes", producerStats.BytesWritten).
				Int("queue_size", len(p.envelopeChan)).
				Msg("stats")
		}
	}
}

// healthHandler handles health check requests
func (p *Processor) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check Kafka connectivity
	if err := p.producer.HealthCheck(ctx); err != nil {
		http.Error(w, fmt.Sprintf("unhealthy: %v", err), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// statsHandler returns current statistics
func (p *Processor) statsHandler(w http.ResponseWriter, r *http.Request) {
	workerStats := p.workerPool.Stats()
	producerStats := p.producer.Stats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{
		"worker": {
			"processed": %d,
			"failed": %d
		},
		"producer": {
			"messages_sent": %d,
			"messages_failed": %d,
			"bytes_written": %d
		},
		"channel": {
			"buffered": %d,
			"capacity": %d
		}
	}`,
		workerStats.Processed,
		workerStats.Failed,
		producerStats.MessagesSent,
		producerStats.MessagesFailed,
		producerStats.BytesWritten,
		len(p.envelopeChan),
		cap(p.envelopeChan),
	)
}
