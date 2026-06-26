package discord

import (
	"context"
	"fmt"
	"strings"
)

const moderatorSummaryPrefix = "◆ moderator summary round="

// channelProgress forwards selected Engine progress lines to a Discord channel.
type channelProgress struct {
	pool      *BotPool
	channelID string
	loc       Locale
}

func (p *channelProgress) Logf(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	if strings.HasPrefix(line, moderatorSummaryPrefix) {
		return
	}
	if !shouldPostProgress(line) {
		return
	}
	sender := p.pool.Default
	if sender == nil {
		return
	}
	_ = sender.Send(context.Background(), p.channelID, localizeProgressLine(line, p.loc))
}

func parseModeratorSummaryLine(line string) (round int, body string, ok bool) {
	if !strings.HasPrefix(line, moderatorSummaryPrefix) {
		return 0, "", false
	}
	rest := line[len(moderatorSummaryPrefix):]
	idx := strings.IndexByte(rest, '\n')
	if idx < 0 {
		return 0, "", false
	}
	if _, err := fmt.Sscanf(strings.TrimSpace(rest[:idx]), "%d", &round); err != nil || round < 0 {
		return 0, "", false
	}
	body = strings.TrimSpace(rest[idx+1:])
	if body == "" {
		return 0, "", false
	}
	return round, body, true
}

func shouldPostProgress(line string) bool {
	if strings.HasPrefix(line, moderatorSummaryPrefix) {
		return false
	}
	if strings.HasPrefix(line, "… waiting") {
		return false
	}
	if strings.HasPrefix(line, "… LLM") {
		return false
	}
	// Token usage is appended to each speaker's own message.
	if strings.HasPrefix(line, "✓ LLM") {
		return false
	}
	if strings.Contains(line, "generating moderator summary") ||
		strings.Contains(line, "generating deliberation summary") {
		return false
	}
	// Internal LLM status — stream output covers readiness/synthesis body.
	if strings.HasPrefix(line, "◆ LLM") ||
		strings.HasPrefix(line, "◆ readiness") ||
		strings.HasPrefix(line, "◆ synthesis readiness round=") ||
		strings.HasPrefix(line, "◆ synthesis completed (") {
		return false
	}
	for _, mark := range []string{"▶", "■", "★", "◇", "⏸", "↩"} {
		if strings.Contains(line, mark) {
			return true
		}
	}
	return strings.Contains(line, "meeting finished")
}
