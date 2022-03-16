package tg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/whiteforestz/iino/internal/domain/sysmon"
)

const (
	apiMethodSendMessage = "sendMessage"
	sendMessageTimeout   = 2 * time.Second
)

func (d *Domain) sendUsageStatsMessage(ctx context.Context) (*dtoMessage, error) {
	var b strings.Builder

	usage, err := d.sysMon.GetUsage()
	if err != nil && !errors.Is(err, sysmon.ErrEmptyUsage) {
		return nil, fmt.Errorf("can't get usage: %w", err)
	}

	if errors.Is(err, sysmon.ErrEmptyUsage) {
		b.WriteString("📈 Usage stats not found 🗿\n")
	} else {
		b.WriteString("📈 Usage stats:\n")
		for _, coreUsage := range usage.CPU {
			b.WriteString(fmt.Sprintf("`%s` \\- %d%%\n", coreUsage.Slug, coreUsage.Percentage))
		}
	}

	b.WriteString("\nCommands:\n")
	b.WriteString("📈 **/usage\\_stats** \\- returns information about server's hardware state")

	return d.sendMessage(ctx, sendMessageIn{
		ChatID:              d.cfg.AdminID,
		Text:                b.String(),
		ParseMode:           sendMessageParseModeMarkdownV2,
		DisableNotification: true,
	})
}

func (d *Domain) sendHelpMessage(ctx context.Context) (*dtoMessage, error) {
	var b strings.Builder

	b.WriteString("🤖 Welcome\\!\n\n")
	b.WriteString("\nCommands:\n")
	b.WriteString("📈 **/usage\\_stats** \\- returns information about server's hardware state")

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
