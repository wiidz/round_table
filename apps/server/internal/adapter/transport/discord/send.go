package discord

import (
	"context"
)

// SendLong posts content, splitting across multiple Discord messages when needed.
func SendLong(sender ChannelSender, ctx context.Context, channelID, content string) {
	for _, part := range splitDiscordMessages(content, maxMessageRunes) {
		if part == "" {
			continue
		}
		_ = sender.Send(ctx, channelID, part)
	}
}
