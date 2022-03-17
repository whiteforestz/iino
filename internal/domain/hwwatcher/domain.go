package hwwatcher

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/whiteforestz/iino/internal/pkg/infloop"
	"github.com/whiteforestz/iino/internal/pkg/logger"
)

const (
	tickerPeriod = 1 * time.Second
)

type Domain struct {
	started  chan struct{}
	finished chan struct{}
	cfg      Config

	mux   *sync.RWMutex
	usage Usage
}

func New(
	cfg Config,
) *Domain {
	return &Domain{
		started:  make(chan struct{}),
		finished: make(chan struct{}),
		cfg:      cfg,

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

	if len(d.usage.CPU) == 0 {
		return nil, ErrEmptyUsage
	}

	usage.CPU = append(usage.CPU, d.usage.CPU...)

	return &usage, nil
}

func (d *Domain) loop(ctx context.Context) {
	ticker := time.NewTicker(tickerPeriod)
	defer ticker.Stop()

	var (
		lastCPULoad []cpuCoreLoad
	)

	infloop.InfLoop(ctx, ticker, infloop.Caller{
		OnStart: func(_ context.Context) {
			close(d.started)
		},
		OnFinish: func(_ context.Context) {
			close(d.finished)
		},
		OnTick: func(_ context.Context) {
			cpuLoad, err := d.updateCPUUsage(lastCPULoad)
			if err != nil {
				logger.Instance().Error("can't update cpu usage", zap.Error(err))
				return
			}

			lastCPULoad = cpuLoad
		},
	})
}

func loadMagicFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("can't open file %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	raw, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("can't read file %q: %w", path, err)
	}

	return string(raw), nil
}
