package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"

	"parsec/internal/logger"
	"parsec/internal/metrics"
)

// responseWriter wraps http.ResponseWriter to capture status and size
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Logging middleware logs all HTTP requests with structured logging
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate request ID
		requestID := uuid.New().String()
		r.Header.Set("X-Request-ID", requestID)

		// Wrap response writer
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Create logger with request context
		log := logger.Logger.With().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Logger()

		log.Info().
			Int("content_length", int(r.ContentLength)).
			Msg("request received")

		// Call next handler
		next.ServeHTTP(rw, r)

		// Log response
		duration := time.Since(start)
		log.Info().
			Int("status", rw.status).
			Int("response_size", rw.size).
			Dur("duration_ms", duration).
			Msg("request completed")

		// Record metrics
		metrics.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			fmt.Sprintf("%d", rw.status),
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			fmt.Sprintf("%d", rw.status),
		).Observe(duration.Seconds())

		if r.ContentLength > 0 {
			metrics.HTTPRequestSize.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(float64(r.ContentLength))
		}

		if rw.size > 0 {
			metrics.HTTPResponseSize.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(float64(rw.size))
		}
	})
}

// Recovery middleware recovers from panics and logs them
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := debug.Stack()

				// Log panic
				requestID := r.Header.Get("X-Request-ID")
				logger.Logger.Error().
					Str("request_id", requestID).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Interface("panic", err).
					Bytes("stack", stack).
					Msg("panic recovered")

				// Record metric
				metrics.PanicsRecovered.WithLabelValues("http_handler").Inc()

				// Return error response
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Chain applies middlewares in order
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
