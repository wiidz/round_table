package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"round_table/apps/server/internal/stream"
)

const discordStreamMaxLen = 1900

// channelStream posts turn headers and formatted LLM output to a Discord channel.
type channelStream struct {
	sender    ChannelSender
	channelID string
	buf       strings.Builder
}

func (s *channelStream) Start(meta stream.Meta) {
	s.buf.Reset()
	line := fmt.Sprintf("↳ %s · %s", meta.ParticipantID, meta.Phase)
	if meta.Detail != "" {
		line += " · " + meta.Detail
	}
	_ = s.sender.Send(context.Background(), s.channelID, line)
}

func (s *channelStream) Delta(delta string) {
	s.buf.WriteString(delta)
}

func (s *channelStream) End() {
	raw := strings.TrimSpace(s.buf.String())
	s.buf.Reset()
	if raw == "" {
		return
	}
	body := formatStreamForDiscord(raw)
	if body == "" {
		body = raw
	}
	body = truncateDiscord(body, discordStreamMaxLen)
	_ = s.sender.Send(context.Background(), s.channelID, body)
}

func formatStreamForDiscord(raw string) string {
	if text := formatParticipantStream(raw); text != "" {
		return text
	}
	if text := formatReadinessStream(raw); text != "" {
		return text
	}
	if text := formatSynthesisStream(raw); text != "" {
		return text
	}
	return ""
}

func formatParticipantStream(raw string) string {
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
		fmt.Fprintf(&b, "\n\n_立场: %s_", part.Stance)
	}
	if strings.TrimSpace(part.ObjectReason) != "" {
		fmt.Fprintf(&b, "\n_反对理由: %s_", strings.TrimSpace(part.ObjectReason))
	}
	return b.String()
}

func formatReadinessStream(raw string) string {
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
		b.WriteString("**研讨就绪** ✓\n")
	} else {
		b.WriteString("**继续研讨**\n")
	}
	if ready.Rationale != "" {
		fmt.Fprintf(&b, "%s\n", ready.Rationale)
	}
	if len(ready.Gaps) > 0 {
		b.WriteString("\n**待补缺口**\n")
		for _, gap := range ready.Gaps {
			fmt.Fprintf(&b, "- %s\n", gap)
		}
	}
	return strings.TrimSpace(b.String())
}

func formatSynthesisStream(raw string) string {
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
	b.WriteString("**设计草案合成**\n")
	if syn.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", syn.Summary)
	}
	writeBulletSection(&b, "方案要点", syn.CoreScheme)
	writeBulletSection(&b, "已决事项", syn.Decisions)
	writeBulletSection(&b, "开放问题", syn.OpenQuestions)
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

func truncateDiscord(text string, max int) string {
	if max <= 0 || len(text) <= max {
		return text
	}
	if max <= 1 {
		return text[:max]
	}
	return text[:max-1] + "…"
}
