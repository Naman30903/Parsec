package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"parsec/internal/config"
	"parsec/internal/processor"
)

func main() {
	// Load configuration from environment
	cfg := config.FromEnv()

	log.Printf("Starting Parsec Log Processor")
	log.Printf("Kafka Brokers: %v", cfg.Kafka.Brokers)
	log.Printf("Kafka Topic: %s", cfg.Kafka.Topic)
	log.Printf("Worker Pool Size: %d", cfg.Kafka.Producer.PoolSize)

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
		log.Printf("received signal: %v", sig)
		cancel()
	case err := <-errChan:
		log.Printf("processor error: %v", err)
		cancel()
	case <-ctx.Done():
		log.Println("context cancelled")
	}

	log.Println("shutdown complete")
}
