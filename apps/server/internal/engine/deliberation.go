package engine

import (
	"context"
	"fmt"
	"strings"

	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/meeting"
)

func (e *Engine) continueAfterDebateRound(ctx context.Context, s meeting.State) (meeting.State, error) {
	if s.IsDeliberation() {
		return e.continueAfterDeliberationRound(ctx, s)
	}
	return e.continueAfterDecisionRound(ctx, s)
}

func (e *Engine) continueAfterDecisionRound(ctx context.Context, s meeting.State) (meeting.State, error) {
	e.logf("◇ consensus check after round %d", s.CurrentRound)
	result, err := e.Strategy.Evaluate(consensus.Context{Meeting: s})
	if err != nil {
		return s, err
	}
	if result.Reached {
		return e.append(ctx, s, eventConsensusReached(s.ConsensusStrategy, result))
	}

	if s.CurrentRound >= s.MaxRoundsPerSegment {
		e.logf("◇ max debate rounds reached (%d) — moderator decision", s.MaxRoundsPerSegment)
		return e.append(ctx, s, eventModeratorDecision(s.ConsensusStrategy))
	}

	e.logf("◆ generating moderator summary for round %d", s.CurrentRound)
	modSummary := moderatorSummarizeRound(s)
	s, err = e.append(ctx, s, eventModeratorSummarized(s.CurrentRound, modSummary))
	if err != nil {
		return s, err
	}
	return e.startRound(ctx, s)
}

func (e *Engine) continueAfterDeliberationRound(ctx context.Context, s meeting.State) (meeting.State, error) {
	e.logf("◆ generating deliberation summary for round %d", s.CurrentRound)
	modSummary := moderatorSummarizeDeliberationRound(s)
	s, err := e.append(ctx, s, eventModeratorSummarized(s.CurrentRound, modSummary))
	if err != nil {
		return s, err
	}

	if s.CurrentRound >= s.MaxRoundsPerSegment {
		e.logf("◇ max deliberation rounds reached (%d) — synthesizing design draft", s.MaxRoundsPerSegment)
		return e.completeDeliberation(ctx, s, "max_rounds")
	}
	return e.startRound(ctx, s)
}

func (e *Engine) completeDeliberation(ctx context.Context, s meeting.State, resolvedBy string) (meeting.State, error) {
	summary, openQuestions := moderatorSynthesizeFinal(s)
	e.logf("◆ synthesis completed (%d open questions)", len(openQuestions))
	return e.append(ctx, s, eventSynthesisCompleted(summary, openQuestions, resolvedBy))
}

func (e *Engine) buildDeliberationPrompt(s meeting.State, participantID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n%s\nRound: %d\nYou are %s (%s).\n",
		s.Topic, PhaseDeliberation, s.CurrentRound, participantID, s.Participants[participantID].Role)
	b.WriteString("\nThis is a **deliberation** meeting — contribute design ideas, constraints, and trade-offs from your role.\n")
	b.WriteString("Do NOT vote approve/reject; focus on building a shared scheme.\n")
	if s.PrincipalFeedback != "" {
		fmt.Fprintf(&b, "\nPrincipal feedback (address this round):\n%s\n", s.PrincipalFeedback)
	}
	if ctx := formatDiscussionContext(s, participantID); ctx != "" {
		b.WriteString("\n--- Discussion so far ---\n")
		b.WriteString(ctx)
		b.WriteByte('\n')
	}
	if e.Workspace != nil {
		if data, err := e.Workspace.Read(s.ID, workspace.FileMeeting); err == nil {
			b.WriteString("\n--- MEETING.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
