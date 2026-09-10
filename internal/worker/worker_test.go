package worker

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

type blockedProcessor struct{ entered chan struct{} }

func (p blockedProcessor) ProcessNext(ctx context.Context) (bool, error) {
	close(p.entered)
	<-ctx.Done()
	return false, ctx.Err()
}

func TestShutdownCancelsInFlightWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := blockedProcessor{entered: make(chan struct{})}
	done := make(chan struct{})
	go func() { Run(ctx, p, slog.New(slog.NewTextHandler(io.Discard, nil))); close(done) }()
	select {
	case <-p.entered:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
}
