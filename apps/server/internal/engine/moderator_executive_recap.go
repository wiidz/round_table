package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/stream"
)

const executiveRecapFormatHint = `Write an executive recap in Markdown (简体中文).
Use ## / ### headings. Do NOT wrap in code fences.
Synthesize the whole meeting arc — do NOT paste participant speeches verbatim.

Required sections:
## 会议回顾

### 目标与议程覆盖
Briefly restate Topic/Goal/Agenda and which items were substantively addressed.

### 过程脉络
Narrative across pre-meeting → rounds → free dialogue (if any): key shifts and why the discussion moved.

### 关键转折
2–5 bullets: pivotal exchanges or disagreements that shaped the direction.

### 当前态势
What is aligned, what remains disputed, what is intentionally left as open questions.

### 进入合成前提示
What the Principal should watch for in the upcoming design draft (decisions still fragile, not to re-litigate).`

func (e *Engine) moderatorExecutiveRecap(ctx context.Context, s meeting.State) string {
	if !e.LLMModeratorExecutiveRecap || e.Model == nil {
		return ""
	}
	recap, err := e.executiveRecapLLM(ctx, s)
	if err != nil {
		e.logf("◆ executive recap failed (%v) — skipped", err)
		return ""
	}
	return recap
}

func (e *Engine) executiveRecapLLM(ctx context.Context, s meeting.State) (string, error) {
	e.logf("◆ LLM executive recap")
	modelName := e.ModelName
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	prompt := e.buildExecutiveRecapPrompt(s)
	system, err := e.buildExecutiveRecapSystem(s)
	if err != nil {
		return "", err
	}

	phaseLabel := strings.TrimPrefix(PhaseModeratorExecutiveRecap, "Phase: ")
	ctx = e.withStreamCtx(ctx, stream.Meta{
		ParticipantID: "moderator",
		Phase:         phaseLabel,
		Detail:        "executive recap before synthesis",
	})

	start := time.Now()
	onDelta, onEnd := synthesisStreamHandlers(ctx)
	raw, err := e.Model.Complete(ctx, model.Request{
		Model: modelName,
		Messages: []model.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt + "\n\n" + executiveRecapFormatHint},
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
	e.logf("◆ LLM executive recap done in %s", time.Since(start).Round(time.Millisecond))

	recap := cleanRoundSummaryOutput(raw.Content)
	if recap == "" {
		return "", fmt.Errorf("empty executive recap")
	}
	return recap, nil
}

func (e *Engine) buildExecutiveRecapSystem(s meeting.State) (string, error) {
	var b strings.Builder
	b.WriteString("You are the RoundTable Moderator writing an executive recap for the Principal before design-draft synthesis.\n")
	b.WriteString("You orchestrate but do not add domain expertise — reflect the meeting record faithfully.\n")
	b.WriteString("Write in 简体中文. Be concise but complete.\n")
	if e.Profile != nil {
		if data, err := e.Profile.ReadModerator(profile.FileAgents); err == nil {
			b.WriteString("\n--- Moderator AGENTS.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
}

func (e *Engine) buildExecutiveRecapPrompt(s meeting.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", PhaseModeratorExecutiveRecap)
	fmt.Fprintf(&b, "Topic: %s\n", s.Topic)
	if s.Goal != "" {
		fmt.Fprintf(&b, "Goal: %s\n", s.Goal)
	}
	fmt.Fprintf(&b, "Rounds completed: %d\n", s.CurrentRound)
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
		b.WriteString("\n## Pre-meeting\n\n")
		b.WriteString(strings.TrimSpace(s.PreMeetingSummary))
		b.WriteString("\n\n")
	}
	if s.FreeDialogueCompleted && s.FreeDialogueSummary != "" {
		b.WriteString("\n## Free dialogue\n\n")
		b.WriteString(strings.TrimSpace(s.FreeDialogueSummary))
		b.WriteString("\n\n")
	}
	for _, r := range s.Minutes.Rounds {
		if r.RoundNumber <= 0 {
			continue
		}
		fmt.Fprintf(&b, "## Round %d transcript\n\n", r.RoundNumber)
		b.WriteString(strings.TrimSpace(r.Summary))
		b.WriteByte('\n')
		if mod, ok := s.ModeratorSummaries[r.RoundNumber]; ok {
			b.WriteString("\n### Moderator round summary\n\n")
			b.WriteString(strings.TrimSpace(mod))
		}
		b.WriteString("\n\n")
	}
	return b.String()
}

func renderMinutesWithRecap(s meeting.State, recap string) string {
	base := renderMinutes(s)
	recap = strings.TrimSpace(recap)
	if recap == "" {
		return base
	}
	marker := "## Synthesis"
	if idx := strings.Index(base, marker); idx >= 0 {
		return base[:idx] + "## Executive Recap\n\n" + recap + "\n\n" + base[idx:]
	}
	return base + "\n## Executive Recap\n\n" + recap + "\n"
}

const executiveRecapWorkspacePath = "moderator/executive-recap.md"

func writeExecutiveRecapFile(ws workspace.Port, meetingID, recap string) error {
	if ws == nil || strings.TrimSpace(recap) == "" {
		return nil
	}
	body := fmt.Sprintf("# Executive Recap\n\n%s\n", strings.TrimSpace(recap))
	return ws.Write(meetingID, executiveRecapWorkspacePath, []byte(body))
}
