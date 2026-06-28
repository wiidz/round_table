package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/stream"
)

const roundSummaryFormatHint = `Write the round summary in Markdown (简体中文).
Use ## / ### headings. Do NOT wrap in code fences.
Do NOT paste participant speeches verbatim — synthesize across roles.

Required sections (use exact headings):
## Round <N> 研讨摘要   (decision mode: ## Round <N> 摘要)

### 本轮进展
2–4 sentences: what advanced this round relative to Topic/Goal/Agenda.

### 已趋同
Bullet list of concrete agreements (may be empty).

### 仍存分歧
Bullet list of unresolved conflicts between participants (may be empty).

### 待决与下轮焦点
Open questions + what the next round should tackle (if any rounds remain; otherwise note synthesis is next).`

func (e *Engine) moderatorRoundSummary(ctx context.Context, s meeting.State) string {
	if e.LLMModeratorRoundSummary && e.Model != nil {
		summary, err := e.summarizeRoundLLM(ctx, s)
		if err != nil {
			e.logf("◆ LLM round summary failed (%v) — rule fallback", err)
		} else if strings.TrimSpace(summary) != "" {
			return summary
		}
	}
	if s.IsDeliberation() {
		return moderatorSummarizeDeliberationRound(s)
	}
	return moderatorSummarizeRound(s)
}

func (e *Engine) summarizeRoundLLM(ctx context.Context, s meeting.State) (string, error) {
	e.logf("◆ LLM round summary (round %d)", s.CurrentRound)
	modelName := e.ModelName
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	prompt := e.buildModeratorRoundSummaryPrompt(s)
	system, err := e.buildModeratorRoundSummarySystem(s)
	if err != nil {
		return "", err
	}

	phaseLabel := strings.TrimPrefix(PhaseModeratorRoundSummary, "Phase: ")
	ctx = e.withStreamCtx(ctx, stream.Meta{
		ParticipantID: "moderator",
		Phase:         phaseLabel,
		Detail:        fmt.Sprintf("round %d summary", s.CurrentRound),
	})

	start := time.Now()
	onDelta, onEnd := synthesisStreamHandlers(ctx)
	raw, err := e.Model.Complete(ctx, model.Request{
		Model: modelName,
		Messages: []model.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt + "\n\n" + roundSummaryFormatHint},
		},
		Temperature: 0.3,
		OnDelta:     onDelta,
	})
	if err != nil {
		return "", err
	}
	if onEnd != nil {
		onEnd()
	}
	e.logf("◆ LLM round summary done in %s", time.Since(start).Round(time.Millisecond))

	summary := cleanRoundSummaryOutput(raw.Content)
	if summary == "" {
		return "", fmt.Errorf("empty round summary")
	}
	return summary, nil
}

func (e *Engine) buildModeratorRoundSummarySystem(s meeting.State) (string, error) {
	var b strings.Builder
	b.WriteString("You are the RoundTable Moderator. Summarize the completed debate round for the Principal and participants.\n")
	b.WriteString("You orchestrate but do not add domain expertise — reflect what was said, where roles agree or clash, and what remains open.\n")
	b.WriteString("Write in 简体中文. Be concise; no filler.\n\n")
	if s.IsDeliberation() {
		b.WriteString("Meeting mode: deliberation (design exploration — no vote stances).\n")
	} else {
		b.WriteString("Meeting mode: decision (note agree/object stances and objections).\n")
	}
	if e.Profile != nil {
		if data, err := e.Profile.ReadModerator(profile.FileAgents); err == nil {
			b.WriteString("\n--- Moderator AGENTS.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
}

func (e *Engine) buildModeratorRoundSummaryPrompt(s meeting.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", PhaseModeratorRoundSummary)
	fmt.Fprintf(&b, "Topic: %s\n", s.Topic)
	if s.Goal != "" {
		fmt.Fprintf(&b, "Goal: %s\n", s.Goal)
	}
	fmt.Fprintf(&b, "Round completed: %d / %d max\n", s.CurrentRound, s.MaxRoundsPerSegment)
	if len(s.Agenda) > 0 {
		b.WriteString("\nAgenda:\n")
		for _, item := range s.Agenda {
			fmt.Fprintf(&b, "- [%s] %s\n", item.ID, item.Title)
		}
	}
	if e.Workspace != nil {
		if data, err := e.Workspace.Read(s.ID, workspace.FileMeeting); err == nil {
			b.WriteString("\n--- MEETING.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}

	if s.PreMeetingSummary != "" {
		b.WriteString("\n## Pre-meeting (Round 0)\n\n")
		b.WriteString(strings.TrimSpace(s.PreMeetingSummary))
		b.WriteString("\n\n")
	}

	for _, r := range s.Minutes.Rounds {
		if r.RoundNumber <= 0 || r.RoundNumber >= s.CurrentRound {
			continue
		}
		fmt.Fprintf(&b, "## Prior Round %d transcript summary\n\n", r.RoundNumber)
		b.WriteString(strings.TrimSpace(r.Summary))
		b.WriteByte('\n')
		if mod, ok := s.ModeratorSummaries[r.RoundNumber]; ok {
			b.WriteString("\n### Prior Moderator summary\n\n")
			b.WriteString(strings.TrimSpace(mod))
		}
		b.WriteByte('\n')
	}

	if s.CurrentRound == 1 && s.FreeDialogueCompleted && s.FreeDialogueSummary != "" {
		b.WriteString("\n## Free dialogue (after Round 1)\n\n")
		b.WriteString(strings.TrimSpace(s.FreeDialogueSummary))
		b.WriteString("\n\n")
	}

	fmt.Fprintf(&b, "\n## Round %d — full participant turns (summarize these)\n\n", s.CurrentRound)
	for _, id := range s.RoundOrder {
		resp, ok := s.RoundResponses[s.CurrentRound][id]
		if !ok {
			continue
		}
		role := s.Participants[id].Role
		fmt.Fprintf(&b, "### %s (%s)\n\n", id, role)
		if !s.IsDeliberation() && resp.Stance != "" && resp.Stance != event.StanceNone {
			fmt.Fprintf(&b, "Stance: %s\n", resp.Stance)
			if resp.ObjectReason != "" {
				fmt.Fprintf(&b, "Object reason: %s\n", resp.ObjectReason)
			}
		}
		b.WriteString(strings.TrimSpace(resp.Content))
		b.WriteString("\n\n")
	}

	return b.String()
}

func cleanRoundSummaryOutput(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```markdown")
	raw = strings.TrimPrefix(raw, "```md")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "{") && strings.Contains(raw, `"content"`) {
		return ""
	}
	if len([]rune(raw)) < 40 {
		return ""
	}
	return raw
}
