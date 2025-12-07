package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "parsec_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "parsec_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "parsec_http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "parsec_http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	// Ingest metrics
	IngestEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "parsec_ingest_events_total",
			Help: "Total number of events received",
		},
		[]string{"tenant_id", "status"}, // status: accepted, rejected
	)

	IngestBatchSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "parsec_ingest_batch_size",
			Help:    "Size of event batches received",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		},
	)

	IngestValidationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "parsec_ingest_validation_errors_total",
			Help: "Total number of validation errors",
		},
		[]string{"error_type"},
	)

	// Worker metrics
	WorkerQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "parsec_worker_queue_size",
			Help: "Current size of the worker queue",
		},
	)

	WorkerQueueCapacity = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "parsec_worker_queue_capacity",
			Help: "Capacity of the worker queue",
		},
	)

	WorkerProcessedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "parsec_worker_processed_total",
			Help: "Total number of events processed by workers",
		},
	)

	WorkerFailedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "parsec_worker_failed_total",
			Help: "Total number of events failed in workers",
		},
	)

	WorkerBatchPublishDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "parsec_worker_batch_publish_duration_seconds",
			Help:    "Time taken to publish a batch to Kafka",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
	)

	// Kafka producer metrics
	KafkaPublishTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "parsec_kafka_publish_total",
			Help: "Total number of messages published to Kafka",
		},
		[]string{"status"}, // status: success, failed
	)

	KafkaPublishDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "parsec_kafka_publish_duration_seconds",
			Help:    "Time taken to publish to Kafka",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
	)

	KafkaPublishRetries = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "parsec_kafka_publish_retries_total",
			Help: "Total number of Kafka publish retries",
		},
	)

	KafkaBytesWritten = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "parsec_kafka_bytes_written_total",
			Help: "Total bytes written to Kafka",
		},
	)

	// Panic recovery
	PanicsRecovered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "parsec_panics_recovered_total",
			Help: "Total number of panics recovered",
		},
		[]string{"component"},
	)
)
