package tg

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

	"github.com/whiteforestz/iino/internal/pkg/logger"
)

const (
	tickPeriod = 1 * time.Second

	apiHostTmpl = "https://api.telegram.org/bot%s/%s"
)

type Domain struct {
	started  chan struct{}
	finished chan struct{}

	httpClient HTTPClient
	sysMon     SysMonDomain
	cfg        Config
}

func New(
	httpClient HTTPClient,
	sysMonDomain SysMonDomain,
	cfg Config,
) *Domain {
	return &Domain{
		started:    make(chan struct{}),
		finished:   make(chan struct{}),
		httpClient: httpClient,
		sysMon:     sysMonDomain,
		cfg:        cfg,
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

	close(d.started)

	var lastOffset int64
	for {
		select {
		case <-ctx.Done():
			close(d.finished)
			return
		case <-ticker.C:
			updates, err := d.getUpdates(ctx, lastOffset)
			if err != nil {
				if !isTimeout(err) {
					logger.Instance().Error("can't get updates", zap.Error(err))
				}
				continue
			}

			for _, update := range updates {
				if err = d.handleUpdate(ctx, update); err != nil {
					if !isTimeout(err) {
						logger.Instance().Error("can't handle update", zap.Error(err))
					}
					break
				}

				lastOffset = update.UpdateID
			}
		}
	}
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	rawOut, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("can't read raw out: %w", err)
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
