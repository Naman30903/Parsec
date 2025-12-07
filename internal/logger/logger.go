package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	// Logger is the global logger instance
	Logger zerolog.Logger
)

// Init initializes the global logger
func Init(level string) {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Configure output
	var output io.Writer = os.Stdout

	// Pretty console logging in development
	if os.Getenv("ENV") == "development" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Create logger with context
	Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	Logger.Info().
		Str("level", logLevel.String()).
		Msg("logger initialized")
}

// WithComponent returns a logger with a component field
func WithComponent(component string) zerolog.Logger {
	return Logger.With().Str("component", component).Logger()
}

// WithRequestID returns a logger with a request ID field
func WithRequestID(requestID string) zerolog.Logger {
	return Logger.With().Str("request_id", requestID).Logger()
}

// WithError returns a logger with an error field
func WithError(err error) zerolog.Logger {
	return Logger.With().Err(err).Logger()
}
