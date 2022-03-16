package sysmon

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/whiteforestz/iino/internal/pkg/logger"
)

const (
	tickPeriod = 1 * time.Second
)

type Domain struct {
	started  chan struct{}
	finished chan struct{}

	cpuMux   *sync.RWMutex
	cpuUsage []CPUCoreUsage
}

func New() *Domain {
	return &Domain{
		started:  make(chan struct{}),
		finished: make(chan struct{}),
		cpuMux:   &sync.RWMutex{},
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

	d.cpuMux.RLock()
	defer d.cpuMux.RUnlock()

	if len(d.cpuUsage) == 0 {
		return nil, ErrEmptyUsage
	}

	usage.CPU = append(usage.CPU, d.cpuUsage...)

	return &usage, nil
}

func (d *Domain) loop(ctx context.Context) {
	ticker := time.NewTicker(tickPeriod)
	defer ticker.Stop()

	close(d.started)

	var lastCPULoad []cpuCoreLoad
	for {
		select {
		case <-ctx.Done():
			close(d.finished)
			return
		case <-ticker.C:
			cpuLoad, err := d.updateCPUUsage(lastCPULoad)
			if err != nil {
				logger.Instance().Error("can't update cpu usage", zap.Error(err))
				continue
			}

			lastCPULoad = cpuLoad
		}
	}
}

func loadMagicFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("can't open file %q: %w", path, err)
	}

	raw, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("can't read file %q: %w", path, err)
	}

	return string(raw), nil
}
