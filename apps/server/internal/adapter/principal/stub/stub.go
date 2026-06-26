package stub

import (
	"context"

	"round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

// Principal is a test double for Confirmation and Running interventions.
type Principal struct {
	// RejectUntilCycle rejects when cycle < RejectUntilCycle (1-based cycle from engine).
	RejectUntilCycle int
	Feedback         string

	// LimitFallbackDecision is returned when brief.LimitFallback is set (ADR-0004 §6).
	LimitFallbackDecision principal.Decision

	// ForceConsensus triggers ConsensusForced at the next debate turn boundary (decision mode).
	ForceConsensus bool
	// ForceSynthesisWhenRoundGTE triggers SynthesisForced when CurrentRound >= this value (deliberation).
	ForceSynthesisWhenRoundGTE int
	ForceSynthesisReason       string

	// AbortWhenRoundGTE triggers MeetingFinished(outcome=aborted) at the next turn boundary.
	AbortWhenRoundGTE int
	AbortReason       string
	// PauseWhenRoundGTE triggers MeetingPaused once at the next turn boundary when CurrentRound >= N.
	PauseWhenRoundGTE int
	PauseReason       string

	pauseTriggered bool
}

var _ principal.Port = (*Principal)(nil)

// Confirm implements principal.Port.
func (p *Principal) Confirm(ctx context.Context, _ string, brief event.ConfirmationBrief, cycle int) (principal.Response, error) {
	if err := ctx.Err(); err != nil {
		return principal.Response{}, err
	}
	if brief.LimitFallback {
		decision := p.LimitFallbackDecision
		if decision == "" {
			decision = principal.DecisionLimitForceApprove
		}
		fb := p.Feedback
		return principal.Response{Decision: decision, Feedback: fb}, nil
	}
	if p.RejectUntilCycle > 0 && cycle < p.RejectUntilCycle {
		fb := p.Feedback
		if fb == "" {
			fb = "需要更多细节"
		}
		return principal.Response{Decision: principal.DecisionRejected, Feedback: fb}, nil
	}
	return principal.Response{Decision: principal.DecisionApproved}, nil
}

// RunningAction implements principal.Port.
func (p *Principal) RunningAction(_ context.Context, _ string, s meeting.State) (principal.RunningIntervention, error) {
	if s.Status != meeting.StatusRunning || s.CurrentRound <= 0 {
		return principal.RunningIntervention{}, nil
	}
	if p.AbortWhenRoundGTE > 0 && s.CurrentRound >= p.AbortWhenRoundGTE {
		reason := p.AbortReason
		if reason == "" {
			reason = "Principal 终止会议"
		}
		return principal.RunningIntervention{Kind: principal.RunningInterventionAbort, Reason: reason}, nil
	}
	if p.PauseWhenRoundGTE > 0 && s.CurrentRound >= p.PauseWhenRoundGTE && !p.pauseTriggered {
		reason := p.PauseReason
		if reason == "" {
			reason = "Principal 暂停会议"
		}
		p.pauseTriggered = true
		return principal.RunningIntervention{Kind: principal.RunningInterventionPause, Reason: reason}, nil
	}
	if s.IsDeliberation() {
		if p.ForceSynthesisWhenRoundGTE > 0 && s.CurrentRound >= p.ForceSynthesisWhenRoundGTE {
			reason := p.ForceSynthesisReason
			if reason == "" {
				reason = "Principal 要求立即合成草案"
			}
			return principal.RunningIntervention{
				Kind:   principal.RunningInterventionForceSynthesis,
				Reason: reason,
			}, nil
		}
		return principal.RunningIntervention{}, nil
	}
	if p.ForceConsensus {
		return principal.RunningIntervention{
			Kind:   principal.RunningInterventionForceConsensus,
			Reason: "Principal 强制宣布共识",
		}, nil
	}
	return principal.RunningIntervention{}, nil
}

// PausedAction implements principal.Port.
func (p *Principal) PausedAction(_ context.Context, _ string, s meeting.State) (principal.RunningIntervention, error) {
	if s.Status != meeting.StatusPaused {
		return principal.RunningIntervention{}, nil
	}
	return principal.RunningIntervention{Kind: principal.RunningInterventionResume}, nil
}
