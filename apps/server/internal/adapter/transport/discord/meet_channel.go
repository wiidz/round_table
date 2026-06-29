package discord

import "strings"

// meetChannelContext shares per-meeting Discord channel state between progress and stream.
type meetChannelContext struct {
	executiveRecapPosted bool
	stopTyping           func()
}

func (c *meetChannelContext) beginTyping(pool *BotPool, participantID, channelID string) {
	if c == nil || pool == nil || channelID == "" {
		return
	}
	c.endTyping()
	sender := pool.SenderFor(participantID)
	if sender == nil {
		sender = pool.Default
	}
	typer, ok := sender.(TypingSender)
	if !ok && pool.Default != nil {
		typer, ok = pool.Default.(TypingSender)
	}
	if !ok {
		return
	}
	c.stopTyping = typer.StartTyping(channelID)
}

func (c *meetChannelContext) endTyping() {
	if c == nil || c.stopTyping == nil {
		return
	}
	c.stopTyping()
	c.stopTyping = nil
}

func isModeratorGenerationProgress(line string) bool {
	return strings.Contains(line, "generating moderator summary") ||
		strings.Contains(line, "generating deliberation summary") ||
		strings.HasPrefix(line, "◆ LLM executive recap") ||
		strings.HasPrefix(line, "◆ LLM round summary") ||
		strings.HasPrefix(line, "◆ LLM synthesis readiness check")
}
