package tg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/whiteforestz/iino/internal/domain/hwwatcher"
	"github.com/whiteforestz/iino/internal/domain/wgwatcher"
)

const (
	apiMethodSendMessage = "sendMessage"
	sendMessageTimeout   = 2 * time.Second

	timeFormat = "2006-01-02 15:04"

	messageHelp = `
🏷 *Commands*
🔧 /hw\_usage \- returns hardware's usage
🥷🏻 /wg\_usage \- returns WireGuard's usage`
)

func (d *Domain) sendHWUsageMessage(ctx context.Context) (*dtoMessage, error) {
	var b strings.Builder

	usage, err := d.hwWatcher.GetUsage()
	if err != nil && !errors.Is(err, hwwatcher.ErrEmptyUsage) {
		return nil, fmt.Errorf("can't get usage: %w", err)
	}

	if errors.Is(err, hwwatcher.ErrEmptyUsage) {
		b.WriteString("🔧 Hardware usage is not found 🗿\n")
	} else {
		b.WriteString("🔧 *Hardware usage*\n")
		for _, core := range usage.CPU {
			b.WriteString(fmt.Sprintf("`%s` \\- %d%%\n", core.Slug, core.Percentage))
		}
	}

	b.WriteString(messageHelp)

	return d.sendMessage(ctx, sendMessageIn{
		ChatID:              d.cfg.AdminID,
		Text:                b.String(),
		ParseMode:           sendMessageParseModeMarkdownV2,
		DisableNotification: true,
	})
}

func (d *Domain) sendWGUsageMessage(ctx context.Context) (*dtoMessage, error) {
	var b strings.Builder

	usage, err := d.wgWatcher.GetUsage()
	if err != nil && !errors.Is(err, wgwatcher.ErrEmptyUsage) {
		return nil, fmt.Errorf("can't get usage: %w", err)
	}

	if errors.Is(err, wgwatcher.ErrEmptyUsage) {
		b.WriteString("🥷🏻 WireGuard usage is not found or empty 🗿\n")
	} else {
		b.WriteString("🥷🏻 *WireGuard usage*\n")
		for _, peer := range usage.Peer {
			if peer.LatestHandshakeUnix != 0 {
				handshakedAt := time.Unix(peer.LatestHandshakeUnix, 0).Format(timeFormat)
				b.WriteString(fmt.Sprintf("`%s` \\- handshaked at `%s`\n", peer.Name, handshakedAt))
			} else {
				b.WriteString(fmt.Sprintf("`%s` \\- offline\n", peer.Name))
			}
		}
	}

	b.WriteString(messageHelp)

	return d.sendMessage(ctx, sendMessageIn{
		ChatID:              d.cfg.AdminID,
		Text:                b.String(),
		ParseMode:           sendMessageParseModeMarkdownV2,
		DisableNotification: true,
	})
}

func (d *Domain) sendHelpMessage(ctx context.Context) (*dtoMessage, error) {
	var b strings.Builder

	b.WriteString("🤖 Hello!\n")
	b.WriteString(messageHelp)

	return d.sendMessage(ctx, sendMessageIn{
		ChatID:              d.cfg.AdminID,
		Text:                b.String(),
		ParseMode:           sendMessageParseModeMarkdownV2,
		DisableNotification: true,
	})
}

func (d *Domain) sendMessage(ctx context.Context, in sendMessageIn) (*dtoMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, sendMessageTimeout)
	defer cancel()

	var (
		host = fmt.Sprintf(apiHostTmpl, d.cfg.APIToken, apiMethodSendMessage)

		out sendMessageOut
	)

	if err := d.performRequest(ctx, host, &in, &out); err != nil {
		return nil, fmt.Errorf("can't perform request: %w", err)
	}

	return (*dtoMessage)(&out), nil
}
