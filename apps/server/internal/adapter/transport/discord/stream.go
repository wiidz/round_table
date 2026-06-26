package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"round_table/apps/server/internal/stream"
)

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
	pending    pendingTurn
	stopTyping func()
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
	s.buf.Reset()
	s.speaker = ""
	if raw == "" {
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
		body = raw
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

func formatParticipantStream(raw string, loc Locale) string {
	var part struct {
		Content      string `json:"content"`
		Stance       string `json:"stance"`
		ObjectReason string `json:"object_reason"`
	}
	if err := json.Unmarshal([]byte(raw), &part); err != nil || part.Content == "" {
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

func formatReadinessStream(raw string, loc Locale) string {
	var ready struct {
		Ready     bool     `json:"ready"`
		Rationale string   `json:"rationale"`
		Gaps      []string `json:"gaps"`
	}
	if err := json.Unmarshal([]byte(raw), &ready); err != nil || ready.Rationale == "" && len(ready.Gaps) == 0 {
		return ""
	}
	var b strings.Builder
	if ready.Ready {
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
	if ready.Rationale != "" {
		fmt.Fprintf(&b, "%s\n", ready.Rationale)
	}
	if len(ready.Gaps) > 0 {
		if loc == LocaleZH {
			b.WriteString("\n**📌 待补缺口**\n")
		} else {
			b.WriteString("\n**📌 Gaps**\n")
		}
		for _, gap := range ready.Gaps {
			fmt.Fprintf(&b, "- %s\n", gap)
		}
	}
	return strings.TrimSpace(b.String())
}

func formatSynthesisStream(raw string, loc Locale) string {
	var syn struct {
		CoreScheme    []string `json:"core_scheme"`
		Decisions     []string `json:"decisions"`
		OpenQuestions []string `json:"open_questions"`
		Summary       string   `json:"summary"`
	}
	if err := json.Unmarshal([]byte(raw), &syn); err != nil {
		return ""
	}
	if syn.Summary == "" && len(syn.CoreScheme) == 0 && len(syn.Decisions) == 0 && len(syn.OpenQuestions) == 0 {
		return ""
	}
	var b strings.Builder
	if loc == LocaleZH {
		b.WriteString("📋 **设计草案合成**\n")
	} else {
		b.WriteString("📋 **Design draft synthesis**\n")
	}
	if syn.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", syn.Summary)
	}
	if loc == LocaleZH {
		writeBulletSection(&b, "💡 方案要点", syn.CoreScheme)
		writeBulletSection(&b, "✅ 已决事项", syn.Decisions)
		writeBulletSection(&b, "❓ 开放问题", syn.OpenQuestions)
	} else {
		writeBulletSection(&b, "💡 Core scheme", syn.CoreScheme)
		writeBulletSection(&b, "✅ Decisions", syn.Decisions)
		writeBulletSection(&b, "❓ Open questions", syn.OpenQuestions)
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
