package middleware

import (
	"fmt"
	"net/http"
	"os"
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
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Auth middleware validates the X-API-Key header against an env var
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get expected API key from env (default for testing)
		expectedKey := os.Getenv("API_KEY")
		if expectedKey == "" {
			expectedKey = "test-api-key-123" // Default for testing
		}

		// Check X-API-Key header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			log := logger.Logger.With().
				Str("remote_addr", r.RemoteAddr).
				Str("path", r.URL.Path).
				Logger()
			log.Warn().Msg("missing API key")

			http.Error(w, `{"error":"missing X-API-Key header"}`, http.StatusUnauthorized)
			return
		}

		if apiKey != expectedKey {
			log := logger.Logger.With().
				Str("remote_addr", r.RemoteAddr).
				Str("path", r.URL.Path).
				Logger()
			log.Warn().Msg("invalid API key")

			http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
			return
		}

		// Valid API key, continue
		next.ServeHTTP(w, r)
	})
}

// Logging middleware logs all HTTP requests with structured logging
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Add request ID if not present
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-ID", requestID)
		}

		// Wrap response writer
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Process request
		next.ServeHTTP(rw, r)

		// Log request
		duration := time.Since(start)
		log := logger.Logger.With().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.status).
			Dur("duration_ms", duration).
			Int("size", rw.size).
			Str("remote_addr", r.RemoteAddr).
			Logger()

		if rw.status >= 400 {
			log.Warn().Msg("request completed with error")
		} else {
			log.Info().Msg("request completed")
		}

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
	})
}

// Recovery middleware recovers from panics and logs them
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log := logger.Logger.With().
					Str("request_id", r.Header.Get("X-Request-ID")).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Logger()

				log.Error().
					Interface("panic", err).
					Str("stack", string(debug.Stack())).
					Msg("panic recovered")

				metrics.PanicsRecovered.WithLabelValues("http_handler").Inc()

				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
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
