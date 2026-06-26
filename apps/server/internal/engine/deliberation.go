package engine

import (
	"context"
	"fmt"
	"strings"

	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/event"
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

	atMax := s.CurrentRound >= s.MaxRoundsPerSegment
	ready := false

	if s.CurrentRound >= s.MinRoundsBeforeSynthesis {
		result, assessErr := e.assessDeliberationReadiness(ctx, s)
		if assessErr != nil {
			return s, assessErr
		}
		ready = result.Ready
		s, err = e.append(ctx, s, eventDeliberationReadinessChecked(
			s.CurrentRound, result.Ready, result.Rationale, result.Gaps, result.Usage,
		))
		if err != nil {
			return s, err
		}
		if ready {
			e.logf("◇ synthesis readiness: ready (%s)", result.Rationale)
		} else if len(result.Gaps) > 0 {
			e.logf("◇ synthesis readiness: not ready — %s", strings.Join(result.Gaps, "; "))
		} else {
			e.logf("◇ synthesis readiness: not ready (%s)", result.Rationale)
		}
	}

	if ready || atMax {
		resolvedBy := synthesisResolvedBy(s.CurrentRound, s.MaxRoundsPerSegment, ready)
		if atMax && !ready {
			e.logf("◇ max deliberation rounds reached (%d) — synthesizing design draft", s.MaxRoundsPerSegment)
		} else if ready {
			e.logf("◇ deliberation ready at round %d — synthesizing design draft", s.CurrentRound)
		}
		return e.completeDeliberation(ctx, s, resolvedBy)
	}
	return e.startRound(ctx, s)
}

func (e *Engine) completeDeliberation(ctx context.Context, s meeting.State, resolvedBy string) (meeting.State, error) {
	summary, openQuestions, usage, agenda, err := e.synthesizeDeliberationFinal(ctx, s)
	if err != nil {
		return s, err
	}
	e.logf("◆ synthesis completed (%d open questions)", len(openQuestions))
	var sections []event.SynthesisAgendaSectionPayload
	var cross *event.SynthesisCrossCuttingPayload
	if agenda != nil {
		sections, cross = synthesisAgendaOutputToEvent(*agenda)
	}
	return e.append(ctx, s, eventSynthesisCompleted(summary, openQuestions, resolvedBy, usage, sections, cross))
}

func (e *Engine) buildDeliberationPrompt(s meeting.State, participantID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n%s\nRound: %d\nYou are %s (%s).\n",
		s.Topic, PhaseDeliberation, s.CurrentRound, participantID, s.Participants[participantID].Role)
	b.WriteString("\nThis is a **deliberation** meeting — contribute design ideas, constraints, and trade-offs from your role.\n")
	b.WriteString("Do NOT vote approve/reject; focus on building a shared scheme.\n")
	if len(s.Agenda) > 0 {
		b.WriteString("\nAgenda items (contribute to relevant items from your role):\n")
		for _, item := range s.Agenda {
			fmt.Fprintf(&b, "- %s\n", item.Title)
		}
		b.WriteByte('\n')
	}
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
