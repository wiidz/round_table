package discord

import (
	"context"
	"fmt"
	"strings"
)

const moderatorSummaryPrefix = "◆ moderator summary round="

// channelProgress forwards selected Engine progress lines to a Discord channel.
type channelProgress struct {
	sender    ChannelSender
	channelID string
}

func (p *channelProgress) Logf(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	if round, body, ok := parseModeratorSummaryLine(line); ok {
		msg := fmt.Sprintf("**第 %d 轮主持人摘要**\n%s", round, body)
		msg = truncateDiscord(msg, discordStreamMaxLen)
		_ = p.sender.Send(context.Background(), p.channelID, msg)
		return
	}
	if !shouldPostProgress(line) {
		return
	}
	_ = p.sender.Send(context.Background(), p.channelID, line)
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
	// Posted as a dedicated formatted message via parseModeratorSummaryLine.
	if strings.HasPrefix(line, moderatorSummaryPrefix) {
		return false
	}
	// Principal / confirmation waits — no Discord action yet.
	if strings.HasPrefix(line, "… waiting") {
		return false
	}
	// Turn start is posted via channelStream.Start (↳).
	if strings.HasPrefix(line, "… LLM") {
		return false
	}
	// Summary body follows in a dedicated message.
	if strings.Contains(line, "generating moderator summary") ||
		strings.Contains(line, "generating deliberation summary") {
		return false
	}
	for _, mark := range []string{"▶", "■", "★", "◆", "◇", "⏸", "↩", "✓"} {
		if strings.Contains(line, mark) {
			return true
		}
	}
	return strings.Contains(line, "meeting finished")
}
