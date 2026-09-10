package worker

import (
	"context"
	"log/slog"
	"time"
)

type Processor interface {
	ProcessNext(context.Context) (bool, error)
}

// Run drains durable pending jobs and backs off on idle/database errors.
func Run(ctx context.Context, p Processor, logger *slog.Logger) {
	for ctx.Err() == nil {
		jobCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		worked, err := p.ProcessNext(jobCtx)
		cancel()
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			logger.Warn("worker storage operation failed")
		}
		if worked && err == nil {
			continue
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
