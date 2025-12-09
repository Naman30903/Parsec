package workertest

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"parsec/internal/models"
	"parsec/internal/worker"
)

// MockPublisher is a mock implementation of Publisher for testing
type MockPublisher struct {
	published  atomic.Uint64
	failed     atomic.Uint64
	shouldFail bool
}

func (m *MockPublisher) Publish(ctx context.Context, envelope *models.Envelope) error {
	if m.shouldFail {
		m.failed.Add(1)
		return context.DeadlineExceeded
	}
	m.published.Add(1)
	return nil
}

func (m *MockPublisher) PublishBatch(ctx context.Context, envelopes []*models.Envelope) error {
	if m.shouldFail {
		m.failed.Add(uint64(len(envelopes)))
		return context.DeadlineExceeded
	}
	m.published.Add(uint64(len(envelopes)))
	return nil
}

func TestWorkerPool_ProcessEnvelopes(t *testing.T) {
	ch := make(chan *models.Envelope, 100)
	mock := &MockPublisher{}

	pool := worker.NewPool(worker.Config{
		Publisher:    mock,
		EnvelopeChan: ch,
		Workers:      2,
		BatchSize:    10,
		BatchTimeout: 100 * time.Millisecond,
	})

	pool.Start()
	defer pool.Stop()

	// Send test envelopes
	numEvents := 25
	for i := 0; i < numEvents; i++ {
		event := &models.LogEvent{
			ID:        "test-evt",
			TenantID:  "tenant-1",
			Timestamp: time.Now(),
			Severity:  models.SeverityInfo,
			Source:    "test",
			Message:   "test message",
		}
		ch <- models.NewEnvelope(event, "test-node")
	}

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	stats := pool.Stats()
	if stats.Processed != uint64(numEvents) {
		t.Errorf("expected %d processed, got %d", numEvents, stats.Processed)
	}

	if mock.published.Load() != uint64(numEvents) {
		t.Errorf("expected %d published, got %d", numEvents, mock.published.Load())
	}
}

func TestWorkerPool_Batching(t *testing.T) {
	ch := make(chan *models.Envelope, 100)
	mock := &MockPublisher{}

	pool := worker.NewPool(worker.Config{
		Publisher:    mock,
		EnvelopeChan: ch,
		Workers:      1,
		BatchSize:    5,
		BatchTimeout: 1 * time.Second, // Long timeout to force batching
	})

	pool.Start()
	defer pool.Stop()

	// Send exactly one batch worth of events
	for i := 0; i < 5; i++ {
		event := &models.LogEvent{
			ID:        "test-evt",
			TenantID:  "tenant-1",
			Timestamp: time.Now(),
			Severity:  models.SeverityInfo,
			Source:    "test",
			Message:   "test message",
		}
		ch <- models.NewEnvelope(event, "test-node")
	}

	// Wait for batch processing
	time.Sleep(200 * time.Millisecond)

	if mock.published.Load() != 5 {
		t.Errorf("expected 5 published in batch, got %d", mock.published.Load())
	}
}

func TestWorkerPool_TimeoutBatch(t *testing.T) {
	ch := make(chan *models.Envelope, 100)
	mock := &MockPublisher{}

	pool := worker.NewPool(worker.Config{
		Publisher:    mock,
		EnvelopeChan: ch,
		Workers:      1,
		BatchSize:    100,                    // Large batch size
		BatchTimeout: 100 * time.Millisecond, // Short timeout
	})

	pool.Start()
	defer pool.Stop()

	// Send only 3 events (less than batch size)
	for i := 0; i < 3; i++ {
		event := &models.LogEvent{
			ID:        "test-evt",
			TenantID:  "tenant-1",
			Timestamp: time.Now(),
			Severity:  models.SeverityInfo,
			Source:    "test",
			Message:   "test message",
		}
		ch <- models.NewEnvelope(event, "test-node")
	}

	// Wait for timeout to trigger
	time.Sleep(300 * time.Millisecond)

	if mock.published.Load() != 3 {
		t.Errorf("expected 3 published via timeout, got %d", mock.published.Load())
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	ch := make(chan *models.Envelope, 100)
	mock := &MockPublisher{}

	pool := worker.NewPool(worker.Config{
		Publisher:    mock,
		EnvelopeChan: ch,
		Workers:      2,
		BatchSize:    10,
		BatchTimeout: 100 * time.Millisecond,
	})

	pool.Start()

	// Send events
	for i := 0; i < 7; i++ {
		event := &models.LogEvent{
			ID:        "test-evt",
			TenantID:  "tenant-1",
			Timestamp: time.Now(),
			Severity:  models.SeverityInfo,
			Source:    "test",
			Message:   "test message",
		}
		ch <- models.NewEnvelope(event, "test-node")
	}

	// Give workers a moment to pick up events
	time.Sleep(50 * time.Millisecond)

	// Stop (should flush remaining events)
	pool.Stop()

	// All events should be processed
	if mock.published.Load() != 7 {
		t.Errorf("expected 7 published after shutdown, got %d", mock.published.Load())
	}
}

func TestWorkerPool_ErrorHandling(t *testing.T) {
	ch := make(chan *models.Envelope, 100)
	mock := &MockPublisher{shouldFail: true}

	pool := worker.NewPool(worker.Config{
		Publisher:    mock,
		EnvelopeChan: ch,
		Workers:      1,
		BatchSize:    5,
		BatchTimeout: 100 * time.Millisecond,
	})

	pool.Start()
	defer pool.Stop()

	// Send events
	for i := 0; i < 5; i++ {
		event := &models.LogEvent{
			ID:        "test-evt",
			TenantID:  "tenant-1",
			Timestamp: time.Now(),
			Severity:  models.SeverityInfo,
			Source:    "test",
			Message:   "test message",
		}
		ch <- models.NewEnvelope(event, "test-node")
	}

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	stats := pool.Stats()
	// With our fallback logic, failed batch attempts individual retries
	// So we expect all to be marked as failed
	if stats.Failed == 0 {
		t.Error("expected some failures")
	}
}
