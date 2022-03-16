package tg

import (
	"context"
	"fmt"
)

const (
	cmdUsageStats = "/usage_stats"
)

func (d *Domain) handleUpdate(ctx context.Context, update dtoUpdate) error {
	if d.shouldSkipUpdate(update) {
		return nil
	}

	var err error
	switch update.Message.Text {
	case cmdUsageStats:
		_, err = d.sendUsageStatsMessage(ctx)
	default:
		_, err = d.sendHelpMessage(ctx)
	}
	if err != nil {
		return fmt.Errorf("can't send help message: %w", err)
	}

	return nil
}

func (d *Domain) shouldSkipUpdate(update dtoUpdate) bool {
	if update.Message == nil || update.Message.From == nil {
		return true
	}

	if update.Message.From.ID != d.cfg.AdminID {
		return true
	}

	if update.Message.Chat.Type != chatTypePrivate || update.Message.Chat.ID != d.cfg.AdminID {
		return true
	}

	return false
}
