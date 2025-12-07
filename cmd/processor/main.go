package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"parsec/internal/config"
	"parsec/internal/logger"
	"parsec/internal/processor"
)

func main() {
	// Initialize logger
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logger.Init(logLevel)

	log := logger.Logger.With().Str("component", "main").Logger()
	log.Info().Msg("Starting Parsec Log Processor")

	// Load configuration from environment
	cfg := config.FromEnv()

	log.Info().
		Strs("kafka_brokers", cfg.Kafka.Brokers).
		Str("kafka_topic", cfg.Kafka.Topic).
		Int("worker_pool_size", cfg.Kafka.Producer.PoolSize).
		Msg("configuration loaded")

	// Create processor
	p := processor.New(cfg)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Run processor in background
	errChan := make(chan error, 1)
	go func() {
		if err := p.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for termination signal or error
	select {
	case sig := <-sigs:
		log.Info().Str("signal", sig.String()).Msg("received signal")
		cancel()
	case err := <-errChan:
		log.Error().Err(err).Msg("processor error")
		cancel()
	case <-ctx.Done():
		log.Info().Msg("context cancelled")
	}

	log.Info().Msg("shutdown complete")
}
