package processor

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"parsec/internal/config"
	"parsec/internal/processor"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Default()

	// create processor
	p := processor.New(cfg)

	// run processor in background
	go func() {
		if err := p.Run(ctx); err != nil {
			log.Printf("processor exited: %v", err)
			cancel()
		}
	}()

	// wait for termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		log.Println("shutting down")
		cancel()
	case <-ctx.Done():
	}

	// give graceful shutdown some time
	time.Sleep(500 * time.Millisecond)
	log.Println("exited")
}
