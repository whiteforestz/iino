package sig

import (
	"context"
	"os"
	"os/signal"
)

func WithCancel(ctx context.Context, sigs ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)

	go func() {
		select {
		case <-ctx.Done():
			break
		case <-ch:
			cancel()
		}

		signal.Stop(ch)
	}()

	return ctx, func() {
		cancel()
		signal.Stop(ch)
	}
}
