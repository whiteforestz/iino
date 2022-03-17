package tglistener

import (
	"context"
	"fmt"
	"time"
)

const (
	apiMethodGetUpdates = "getUpdates"
	getUpdatesTimeout   = 10 * time.Second
	getUpdatesLimit     = 100
)

var (
	longPollAllowedUpdates = []string{getUpdatesUpdateTypeMessage}
)

func (d *Domain) getUpdates(ctx context.Context, lastUpdateID int64) ([]dtoUpdate, error) {
	ctx, cancel := context.WithTimeout(ctx, getUpdatesTimeout)
	defer cancel()

	var (
		host = fmt.Sprintf(apiHostTmpl, d.cfg.APIToken, apiMethodGetUpdates)

		in = getUpdatesIn{
			Limit:          getUpdatesLimit,
			Timeout:        int64(getUpdatesTimeout.Seconds()),
			AllowedUpdates: longPollAllowedUpdates,
		}
		out getUpdatesOut
	)

	if lastUpdateID != 0 {
		in.Offset = lastUpdateID + 1
	}

	if err := d.performRequest(ctx, host, &in, &out); err != nil {
		return nil, fmt.Errorf("can't perform request: %w", err)
	}

	if out.Result == nil {
		return nil, fmt.Errorf("empty result")
	}

	return *out.Result, nil
}
