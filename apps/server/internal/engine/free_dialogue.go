package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/stream"
)

func (e *Engine) startFreeDialogue(ctx context.Context, s meeting.State) (meeting.State, error) {
	s, err := e.append(ctx, s, eventFreeDialogueStarted(1, s.FreeDialogueMaxQuestions))
	if err != nil {
		return s, err
	}
	return e.advanceFreeDialogue(ctx, s)
}

func (e *Engine) advanceFreeDialogue(ctx context.Context, s meeting.State) (meeting.State, error) {
	if s.PendingFreeDialogue != nil {
		return e.inviteFreeDialogueAnswer(ctx, s)
	}
	if freeDialogueTotal(s) <= 0 || s.FreeDialogueQuestionIndex >= freeDialogueTotal(s) {
		return e.completeFreeDialogue(ctx, s)
	}
	return e.inviteFreeDialogueAsk(ctx, s)
}

func (e *Engine) inviteFreeDialogueAsk(ctx context.Context, s meeting.State) (meeting.State, error) {
	askerID, answererID := freeDialoguePair(s)
	prompt := e.buildFreeDialogueAskPrompt(s, askerID, answererID)
	detail := freeDialogueTurnLabel(s) + " ask " + askerID + "→" + answererID
	e.logLLMWaiting("free-dialogue-ask", askerID, detail)
	ctx = e.withStreamCtx(ctx, stream.Meta{
		ParticipantID: askerID,
		Phase:         "free-dialogue-ask",
		Detail:        freeDialogueTurnLabel(s) + " ask → " + answererID,
	})
	start := time.Now()
	resp, err := e.Participant.Respond(ctx, s.ID, askerID, prompt)
	elapsed := time.Since(start)
	if err != nil {
		return s, err
	}
	e.logLLMDone("free-dialogue-ask", askerID, "none", resp, elapsed)
	return e.append(ctx, s, eventFreeDialogueQuestionAsked(
		askerID, answererID, s.FreeDialogueQuestionIndex, resp.Content,
		tokenUsageFromResponse(PhaseFreeDialogueAsk, askerID, 1, s.FreeDialogueQuestionIndex, resp),
	))
}

func (e *Engine) inviteFreeDialogueAnswer(ctx context.Context, s meeting.State) (meeting.State, error) {
	pending := s.PendingFreeDialogue
	prompt := e.buildFreeDialogueAnswerPrompt(s, pending.AnswererID, pending.Question)
	detail := freeDialogueTurnLabel(s) + " answer for " + pending.AskerID
	e.logLLMWaiting("free-dialogue-answer", pending.AnswererID, detail)
	ctx = e.withStreamCtx(ctx, stream.Meta{
		ParticipantID: pending.AnswererID,
		Phase:         "free-dialogue-answer",
		Detail:        freeDialogueTurnLabel(s) + " answer ← " + pending.AskerID,
	})
	start := time.Now()
	resp, err := e.Participant.Respond(ctx, s.ID, pending.AnswererID, prompt)
	elapsed := time.Since(start)
	if err != nil {
		return s, err
	}
	e.logLLMDone("free-dialogue-answer", pending.AnswererID, "none", resp, elapsed)
	return e.append(ctx, s, eventFreeDialogueAnswered(
		pending.AskerID, pending.AnswererID, pending.QuestionIndex,
		pending.Question, resp.Content,
		tokenUsageFromResponse(PhaseFreeDialogueAnswer, pending.AnswererID, 1, pending.QuestionIndex, resp),
	))
}

func (e *Engine) completeFreeDialogue(ctx context.Context, s meeting.State) (meeting.State, error) {
	summary := summarizeFreeDialogue(s)
	s, err := e.append(ctx, s, eventFreeDialogueCompleted(1, summary))
	if err != nil {
		return s, err
	}
	return e.continueAfterDebateRound(ctx, s)
}

func freeDialogueTotal(s meeting.State) int {
	return s.FreeDialogueMaxQuestions * len(s.ParticipantOrder)
}

func freeDialoguePair(s meeting.State) (askerID, answererID string) {
	n := len(s.ParticipantOrder)
	if n == 0 {
		return "", ""
	}
	askerIdx := s.FreeDialogueAskerIndex % n
	answererIdx := (askerIdx + 1) % n
	return s.ParticipantOrder[askerIdx], s.ParticipantOrder[answererIdx]
}

func (e *Engine) buildFreeDialogueAskPrompt(s meeting.State, askerID, answererID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n%s\nFree dialogue after Round 1\nYou are %s (%s).\n",
		s.Topic, PhaseFreeDialogueAsk, askerID, s.Participants[askerID].Role)
	fmt.Fprintf(&b, "Ask one focused question to **%s** (%s) to clarify their position or probe assumptions.\n",
		answererID, s.Participants[answererID].Role)
	if ctx := formatFreeDialogueContext(s); ctx != "" {
		b.WriteString("\n--- Discussion so far ---\n")
		b.WriteString(ctx)
		b.WriteByte('\n')
	}
	if e.Workspace != nil {
		if data, err := e.Workspace.Read(s.ID, "rounds/round-001.md"); err == nil {
			b.WriteString("\n--- Round 1 ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (e *Engine) buildFreeDialogueAnswerPrompt(s meeting.State, answererID, question string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n%s\nFree dialogue after Round 1\nYou are %s (%s).\n",
		s.Topic, PhaseFreeDialogueAnswer, answererID, s.Participants[answererID].Role)
	b.WriteString("Answer the question below directly and substantively.\n\n")
	fmt.Fprintf(&b, "**Question:** %s\n", question)
	if ctx := formatFreeDialogueContext(s); ctx != "" {
		b.WriteString("\n--- Discussion so far ---\n")
		b.WriteString(ctx)
		b.WriteByte('\n')
	}
	return b.String()
}

func formatFreeDialogueContext(s meeting.State) string {
	var b strings.Builder
	if s.PreMeetingSummary != "" {
		b.WriteString("## Pre-meeting (Round 0)\n\n")
		b.WriteString(strings.TrimSpace(s.PreMeetingSummary))
		b.WriteString("\n\n")
	}
	for _, r := range s.Minutes.Rounds {
		if r.RoundNumber == 1 {
			fmt.Fprintf(&b, "## Round 1\n\n%s\n\n", strings.TrimSpace(r.Summary))
		}
	}
	for _, ex := range s.FreeDialogueExchanges {
		askRole := s.Participants[ex.AskerID].Role
		ansRole := s.Participants[ex.AnswererID].Role
		fmt.Fprintf(&b, "**%s** (%s) asked **%s** (%s): %s\n",
			ex.AskerID, askRole, ex.AnswererID, ansRole, ex.Question)
		fmt.Fprintf(&b, "**%s** answered: %s\n\n", ex.AnswererID, ex.Answer)
	}
	if pending := s.PendingFreeDialogue; pending != nil {
		askRole := s.Participants[pending.AskerID].Role
		fmt.Fprintf(&b, "**%s** (%s) asked: %s _(awaiting answer)_\n",
			pending.AskerID, askRole, pending.Question)
	}
	return strings.TrimSpace(b.String())
}

func summarizeFreeDialogue(s meeting.State) string {
	var b strings.Builder
	b.WriteString("Free dialogue after Round 1\n\n")
	for _, ex := range s.FreeDialogueExchanges {
		askRole := s.Participants[ex.AskerID].Role
		ansRole := s.Participants[ex.AnswererID].Role
		fmt.Fprintf(&b, "**%s** (%s) → **%s** (%s)\n\n", ex.AskerID, askRole, ex.AnswererID, ansRole)
		fmt.Fprintf(&b, "Q: %s\n\nA: %s\n\n", ex.Question, ex.Answer)
	}
	return strings.TrimSpace(b.String())
}
