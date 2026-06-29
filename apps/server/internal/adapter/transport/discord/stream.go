package discord

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	participantllm "round_table/apps/server/internal/adapter/participant/llm"
	"round_table/apps/server/internal/engine"
	"round_table/apps/server/internal/stream"
)

var reRoundSummaryHeading = regexp.MustCompile(`(?m)^##\s*Round\s+(\d+)\s`)

type pendingTurn struct {
	speaker string
	raw     string
}

// channelStream posts turn headers and formatted LLM output to a Discord channel.
type channelStream struct {
	pool       *BotPool
	channelID  string
	loc        Locale
	buf        strings.Builder
	speaker    string
	phase      string
	pending    pendingTurn
	stopTyping func()
}

func suppressDiscordStreamPost(phase string) bool {
	switch phase {
	case "moderator-executive-recap", "moderator-round-summary":
		return true
	default:
		return false
	}
}

func (s *channelStream) beginTyping(participantID string) {
	s.endTyping()
	if s.pool == nil {
		return
	}
	sender := s.pool.SenderFor(participantID)
	if sender == nil {
		return
	}
	typer, ok := sender.(TypingSender)
	if !ok {
		return
	}
	s.stopTyping = typer.StartTyping(s.channelID)
}

func (s *channelStream) endTyping() {
	if s.stopTyping != nil {
		s.stopTyping()
		s.stopTyping = nil
	}
}

func (s *channelStream) Start(meta stream.Meta) {
	s.buf.Reset()
	s.pending = pendingTurn{}
	s.speaker = meta.ParticipantID
	s.phase = meta.Phase

	if suppressDiscordStreamPost(meta.Phase) {
		return
	}

	s.beginTyping(meta.ParticipantID)

	skipHeader := meta.ParticipantID == "moderator" &&
		(meta.Phase == "deliberation-readiness" || meta.Phase == "deliberation-synthesis")
	// Dedicated participant bots show native Discord typing under their account name.
	if skipHeader || (s.pool != nil && s.pool.HasBot(meta.ParticipantID)) {
		return
	}
	line := formatStreamStart(streamMeta{
		ParticipantID: meta.ParticipantID,
		Phase:         meta.Phase,
		Detail:        meta.Detail,
	}, s.loc)
	sender := s.pool.SenderFor(s.speaker)
	if sender == nil {
		return
	}
	_ = sender.Send(context.Background(), s.channelID, line)
}

func (s *channelStream) Delta(delta string) {
	s.buf.WriteString(delta)
}

func (s *channelStream) End() {
	raw := strings.TrimSpace(s.buf.String())
	speaker := s.speaker
	phase := s.phase
	s.buf.Reset()
	s.speaker = ""
	s.phase = ""
	if raw == "" {
		s.endTyping()
		return
	}
	if suppressDiscordStreamPost(phase) {
		s.endTyping()
		return
	}
	if s.pool != nil && s.pool.HasBot(speaker) {
		s.pending = pendingTurn{speaker: speaker, raw: raw}
		return
	}
	s.endTyping()
	s.postTurn(speaker, raw, 0, 0)
}

func (s *channelStream) CompleteTurn(participantID string, tokens int, elapsed time.Duration) {
	if s.pending.speaker == "" || s.pending.speaker != participantID {
		return
	}
	s.endTyping()
	s.postTurn(s.pending.speaker, s.pending.raw, tokens, elapsed)
	s.pending = pendingTurn{}
}

func (s *channelStream) postTurn(speaker, raw string, tokens int, elapsed time.Duration) {
	body := formatStreamForDiscord(raw, s.loc)
	if body == "" {
		body = fallbackStreamBody(raw, s.loc)
	}
	if footer := formatTurnFooter(tokens, elapsed, s.loc); footer != "" {
		body += footer
	}
	sender := s.pool.SenderFor(speaker)
	if sender == nil {
		return
	}
	SendLong(sender, context.Background(), s.channelID, body)
}

func formatTurnFooter(tokens int, elapsed time.Duration, loc Locale) string {
	if tokens <= 0 && elapsed <= 0 {
		return ""
	}
	if loc == LocaleZH {
		if tokens > 0 && elapsed > 0 {
			return fmt.Sprintf("\n\n_⏱ %s · %d Token_", elapsed.Round(time.Millisecond), tokens)
		}
		if elapsed > 0 {
			return fmt.Sprintf("\n\n_⏱ %s_", elapsed.Round(time.Millisecond))
		}
		return fmt.Sprintf("\n\n_%d Token_", tokens)
	}
	if tokens > 0 && elapsed > 0 {
		return fmt.Sprintf("\n\n_⏱ %s · %d tokens_", elapsed.Round(time.Millisecond), tokens)
	}
	if elapsed > 0 {
		return fmt.Sprintf("\n\n_⏱ %s_", elapsed.Round(time.Millisecond))
	}
	return fmt.Sprintf("\n\n_%d tokens_", tokens)
}

func formatStreamForDiscord(raw string, loc Locale) string {
	if text := formatMarkdownModeratorStream(raw, loc); text != "" {
		return text
	}
	if text := formatParticipantStream(raw, loc); text != "" {
		return text
	}
	if text := formatReadinessStream(raw, loc); text != "" {
		return text
	}
	if text := formatSynthesisStream(raw, loc); text != "" {
		return text
	}
	return ""
}

func formatMarkdownModeratorStream(raw string, loc Locale) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "## 会议回顾") || strings.Contains(raw, "### 目标与议程覆盖") {
		return formatExecutiveRecapDiscord(raw, loc)
	}
	if m := reRoundSummaryHeading.FindStringSubmatch(raw); len(m) == 2 {
		round, err := strconv.Atoi(m[1])
		if err == nil {
			return formatModeratorRoundSummaryDiscord(round, raw, loc)
		}
	}
	return ""
}

func formatParticipantStream(raw string, loc Locale) string {
	part, err := participantllm.ParseOutput(raw)
	if err != nil || strings.TrimSpace(part.Content) == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(part.Content)
	if part.Stance != "" && part.Stance != "none" {
		if loc == LocaleZH {
			fmt.Fprintf(&b, "\n\n_立场：%s_", part.Stance)
		} else {
			fmt.Fprintf(&b, "\n\n_Stance: %s_", part.Stance)
		}
	}
	if strings.TrimSpace(part.ObjectReason) != "" {
		if loc == LocaleZH {
			fmt.Fprintf(&b, "\n_反对理由：%s_", strings.TrimSpace(part.ObjectReason))
		} else {
			fmt.Fprintf(&b, "\n_Objection: %s_", strings.TrimSpace(part.ObjectReason))
		}
	}
	return b.String()
}

func fallbackStreamBody(raw string, loc Locale) string {
	if text := formatMarkdownModeratorStream(raw, loc); text != "" {
		return text
	}
	if text := formatParticipantStream(raw, loc); text != "" {
		return text
	}
	if text := formatReadinessStream(raw, loc); text != "" {
		return text
	}
	if text := formatSynthesisStream(raw, loc); text != "" {
		return text
	}
	if looksLikeJSON(raw) {
		if loc == LocaleZH {
			return "（内容解析失败，完整记录已写入 workspace）"
		}
		return "(Parse failed; full record is in the workspace.)"
	}
	return raw
}

func looksLikeJSON(raw string) bool {
	raw = strings.TrimSpace(raw)
	return strings.HasPrefix(raw, "{") || strings.HasPrefix(raw, "```json") || strings.HasPrefix(raw, "```\n{")
}

func formatReadinessStream(raw string, loc Locale) string {
	out, err := engine.ParseReadinessOutput(raw)
	if err != nil || out.Rationale == "" && len(out.Gaps) == 0 {
		return ""
	}
	var b strings.Builder
	if out.Ready {
		if loc == LocaleZH {
			b.WriteString("✅ **研讨就绪**\n")
		} else {
			b.WriteString("✅ **Ready for synthesis**\n")
		}
	} else if loc == LocaleZH {
		b.WriteString("⏳ **继续研讨**\n")
	} else {
		b.WriteString("⏳ **Continue deliberation**\n")
	}
	if out.Rationale != "" {
		fmt.Fprintf(&b, "%s\n", out.Rationale)
	}
	if len(out.Gaps) > 0 {
		if loc == LocaleZH {
			b.WriteString("\n**📌 待补缺口**\n")
		} else {
			b.WriteString("\n**📌 Gaps**\n")
		}
		for _, gap := range out.Gaps {
			fmt.Fprintf(&b, "- %s\n", gap)
		}
	}
	return strings.TrimSpace(b.String())
}

func formatSynthesisStream(raw string, loc Locale) string {
	out, err := engine.ParseSynthesisOutput(raw)
	if err != nil {
		return ""
	}
	if out.ExecutiveVerdict == "" && len(out.KeyDecisions)+len(out.CoreScheme)+len(out.Decisions)+len(out.OpenQuestions) == 0 {
		return ""
	}
	var b strings.Builder
	if loc == LocaleZH {
		b.WriteString("📋 **设计草案合成**\n")
	} else {
		b.WriteString("📋 **Design draft synthesis**\n")
	}
	if out.ExecutiveVerdict != "" {
		fmt.Fprintf(&b, "%s\n\n", out.ExecutiveVerdict)
	}
	if loc == LocaleZH {
		writeBulletSection(&b, "📌 Principal 需知", out.KeyDecisions)
		writeBulletSection(&b, "💡 方案要点", out.CoreScheme)
		writeBulletSection(&b, "✅ 已决事项", out.Decisions)
		writeBulletSection(&b, "❓ 开放问题", out.OpenQuestions)
	} else {
		writeBulletSection(&b, "📌 Key decisions", out.KeyDecisions)
		writeBulletSection(&b, "💡 Core scheme", out.CoreScheme)
		writeBulletSection(&b, "✅ Decisions", out.Decisions)
		writeBulletSection(&b, "❓ Open questions", out.OpenQuestions)
	}
	return strings.TrimSpace(b.String())
}

func writeBulletSection(b *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "\n**%s**\n", title)
	for i, item := range items {
		fmt.Fprintf(b, "%d. %s\n", i+1, item)
	}
}
