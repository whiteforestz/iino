package infloop

import (
	"context"
	"time"
)

type Callback func(ctx context.Context)

type Caller struct {
	OnStart  Callback
	OnFinish Callback
	OnTick   Callback
}

func InfLoop(ctx context.Context, ticker *time.Ticker, caller Caller) {
	caller.OnStart(ctx)

	for {
		select {
		case <-ctx.Done():
			caller.OnFinish(ctx)
			return
		case <-ticker.C:
			caller.OnTick(ctx)
		}
	}
}
