package tglistener

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/whiteforestz/iino/internal/pkg/infloop"
	"github.com/whiteforestz/iino/internal/pkg/logger"
)

const (
	tickPeriod = 1 * time.Second

	apiHostTmpl = "https://api.telegram.org/bot%s/%s"
)

type Domain struct {
	started    chan struct{}
	finished   chan struct{}
	cfg        Config
	httpClient HTTPClient
	hwWatcher  HWWatcherDomain
	wgWatcher  WGWatcherDomain
}

func New(
	cfg Config,
	httpClient HTTPClient,
	hwWatcherDomain HWWatcherDomain,
	wgWatcherDomain WGWatcherDomain,
) *Domain {
	return &Domain{
		started:    make(chan struct{}),
		finished:   make(chan struct{}),
		cfg:        cfg,
		httpClient: httpClient,
		hwWatcher:  hwWatcherDomain,
		wgWatcher:  wgWatcherDomain,
	}
}

func (d *Domain) Listen(ctx context.Context) {
	go d.loop(ctx)
	<-d.started
}

func (d *Domain) Wait() {
	<-d.finished
}

func (d *Domain) loop(ctx context.Context) {
	ticker := time.NewTicker(tickPeriod)
	defer ticker.Stop()

	var (
		lastUpdateID int64
	)

	infloop.InfLoop(ctx, ticker, infloop.Caller{
		OnStart: func(_ context.Context) {
			close(d.started)
		},
		OnFinish: func(_ context.Context) {
			close(d.finished)
		},
		OnTick: func(ctx context.Context) {
			updates, err := d.getUpdates(ctx, lastUpdateID)
			if err != nil {
				if !isTimeout(err) {
					logger.Instance().Error("can't get updates", zap.Error(err))
				}
				return
			}

			for _, update := range updates {
				if err = d.handleUpdate(ctx, update); err != nil {
					if !isTimeout(err) {
						logger.Instance().Error("can't handle update", zap.Error(err))
					}
					break
				}

				lastUpdateID = update.UpdateID
			}
		},
	})
}

func (d *Domain) performRequest(ctx context.Context, host string, in, out interface{}) error {
	rawIn, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("can't marshal in: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host, bytes.NewBuffer(rawIn))
	if err != nil {
		return fmt.Errorf("can't create req: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("can't do req: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	rawOut, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("can't read raw out: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Instance().Warn("http error", zap.String("response", string(rawOut)))
		return fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	if err = json.Unmarshal(rawOut, out); err != nil {
		return fmt.Errorf("can't unmarshal out: %w", err)
	}

	return nil
}

func isTimeout(err error) bool {
	return os.IsTimeout(err) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled)
}
