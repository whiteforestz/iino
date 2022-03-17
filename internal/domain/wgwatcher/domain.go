package wgwatcher

import (
	"context"
	"sync"
	"time"

	"github.com/whiteforestz/iino/internal/pkg/infloop"
	"github.com/whiteforestz/iino/internal/pkg/logger"
	"go.uber.org/zap"
)

const (
	tickerPeriod = 1 * time.Second
)

type Domain struct {
	started   chan struct{}
	finished  chan struct{}
	cfg       Config
	persistor PersistorDomain

	mux   *sync.RWMutex
	usage Usage
}

func New(
	cfg Config,
	persistorDomain PersistorDomain,
) *Domain {
	return &Domain{
		started:   make(chan struct{}),
		finished:  make(chan struct{}),
		cfg:       cfg,
		persistor: persistorDomain,

		mux: &sync.RWMutex{},
	}
}

func (d *Domain) Listen(ctx context.Context) {
	go d.loop(ctx)
	<-d.started
}

func (d *Domain) Wait() {
	<-d.finished
}

func (d *Domain) GetUsage() (*Usage, error) {
	var usage Usage

	d.mux.RLock()
	defer d.mux.RUnlock()

	if len(d.usage.Peer) == 0 {
		return nil, ErrEmptyUsage
	}

	usage.Peer = append(usage.Peer, d.usage.Peer...)

	return &usage, nil
}

func (d *Domain) loop(ctx context.Context) {
	ticker := time.NewTicker(tickerPeriod)
	defer ticker.Stop()

	infloop.InfLoop(ctx, ticker, infloop.Caller{
		OnStart: func(_ context.Context) {
			close(d.started)
		},
		OnFinish: func(_ context.Context) {
			close(d.finished)
		},
		OnTick: func(ctx context.Context) {
			if err := d.updateUsage(ctx); err != nil {
				logger.Instance().Error("can't update usage", zap.Error(err))
			}
		},
	})
}
