package discord

import (
	"context"
	"fmt"
	"strings"
)

const moderatorSummaryPrefix = "◆ moderator summary round="
const executiveRecapPrefix = "◆ executive recap"

// channelProgress forwards selected Engine progress lines to a Discord channel.
type channelProgress struct {
	pool               *BotPool
	channelID          string
	loc                Locale
	principal          *ChannelPrincipal
	pendingEngineStart string
}

func (p *channelProgress) Logf(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	if round, body, ok := parseModeratorSummaryLine(line); ok {
		p.postModeratorContent(formatModeratorRoundSummaryDiscord(round, body, p.loc))
		return
	}
	if body, ok := parseExecutiveRecapLine(line); ok {
		p.postModeratorContent(formatExecutiveRecapDiscord(body, p.loc))
		return
	}
	if p.principal != nil {
		if strings.HasPrefix(line, "▶ free dialogue after round") {
			p.principal.MarkFreeDialogue(p.channelID, true)
		} else if strings.HasPrefix(line, "■ free dialogue completed") {
			p.principal.MarkFreeDialogue(p.channelID, false)
		}
	}
	if strings.HasPrefix(line, "▶ engine run started") {
		p.pendingEngineStart = line
		return
	}
	if p.pendingEngineStart != "" && strings.HasPrefix(line, "▶ pre-meeting started") {
		p.postProgress(mergeMeetingStartProgress(p.pendingEngineStart, line, p.loc))
		p.pendingEngineStart = ""
		return
	}
	if p.pendingEngineStart != "" {
		p.postProgress(localizeProgressLine(p.pendingEngineStart, p.loc))
		p.pendingEngineStart = ""
	}
	if !shouldPostProgress(line) {
		return
	}
	p.postProgress(localizeProgressLine(line, p.loc))
}

func (p *channelProgress) postProgress(content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	sender := p.pool.Default
	if sender == nil {
		return
	}
	_ = sender.Send(context.Background(), p.channelID, content)
}

func (p *channelProgress) postModeratorContent(content string) {
	content = strings.TrimSpace(content)
	if content == "" || p.pool.Default == nil {
		return
	}
	SendLong(p.pool.Default, context.Background(), p.channelID, content)
}

func formatModeratorRoundSummaryDiscord(round int, body string, loc Locale) string {
	body = stripDuplicateModeratorHeading(body)
	if loc == LocaleZH {
		return fmt.Sprintf("📝 **主持人 · 第 %d 轮摘要**\n\n%s", round, body)
	}
	return fmt.Sprintf("📝 **Moderator · Round %d summary**\n\n%s", round, body)
}

func formatExecutiveRecapDiscord(body string, loc Locale) string {
	body = stripDuplicateModeratorHeading(body)
	if loc == LocaleZH {
		return "📖 **会议回顾 · Executive Recap**\n\n" + body
	}
	return "📖 **Executive Recap**\n\n" + body
}

func stripDuplicateModeratorHeading(body string) string {
	body = strings.TrimSpace(body)
	for _, prefix := range []string{"## 会议回顾", "## Executive Recap"} {
		if strings.HasPrefix(body, prefix) {
			body = strings.TrimPrefix(body, prefix)
			return strings.TrimLeft(body, "\n")
		}
	}
	return body
}

func parseExecutiveRecapLine(line string) (body string, ok bool) {
	if !strings.HasPrefix(line, executiveRecapPrefix) {
		return "", false
	}
	rest := strings.TrimPrefix(line, executiveRecapPrefix)
	rest = strings.TrimLeft(rest, "\n")
	body = strings.TrimSpace(rest)
	return body, body != ""
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
	// Free-dialogue Q&A body is posted by participant stream; skip progress duplicates.
	// Principal questions are already shown in the queued ack — no Moderator relay line.
	if strings.HasPrefix(line, "◆ free dialogue question ") ||
		strings.HasPrefix(line, "◆ free dialogue answer ") {
		return false
	}
	// Internal LLM status — stream output covers readiness/synthesis body.
	if strings.HasPrefix(line, "◆ LLM") ||
		strings.HasPrefix(line, "◆ readiness") ||
		strings.HasPrefix(line, "◆ executive recap failed") ||
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
