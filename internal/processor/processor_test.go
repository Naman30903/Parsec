package processor

import (
	"context"
	"testing"
	"time"

	"parsec/internal/config"
)

func TestProcessorRun(t *testing.T) {
	cfg := config.Default()
	p := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := p.Run(ctx); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

}
