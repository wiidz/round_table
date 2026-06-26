package engine

import (
	"context"
	"fmt"

	"round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

func (e *Engine) afterConsensus(ctx context.Context, s meeting.State) (meeting.State, error) {
	if s.ConfirmationMode == meeting.ConfirmationModeSkip {
		return e.finishMeeting(ctx, s)
	}
	return e.beginConfirmation(ctx, s)
}

func (e *Engine) beginConfirmation(ctx context.Context, s meeting.State) (meeting.State, error) {
	if e.Principal == nil {
		return s, errPrincipalRequired
	}
	cycle := confirmationCycle(s)
	brief := prepareConfirmationBrief(s)
	s, err := e.append(ctx, s, eventConfirmationPrepared(cycle, brief))
	if err != nil {
		return s, err
	}
	return e.append(ctx, s, eventConfirmationPresented(cycle))
}

func (e *Engine) advanceConfirmation(ctx context.Context, s meeting.State) (meeting.State, error) {
	if e.Principal == nil {
		return s, errPrincipalRequired
	}
	if s.Confirmation == nil {
		return s, errNoConfirmationBrief
	}

	e.logf("… waiting for principal decision cycle=%d", s.Confirmation.Cycle)
	resp, err := e.Principal.Confirm(ctx, s.ID, s.Confirmation.Brief, s.Confirmation.Cycle)
	if err != nil {
		return s, err
	}

	switch resp.Decision {
	case principal.DecisionApproved:
		s, err = e.append(ctx, s, eventConfirmationApproved(s.Confirmation.Cycle, resp.ItemNotes))
		if err != nil {
			return s, err
		}
		return e.finishMeeting(ctx, s)
	case principal.DecisionRejected:
		if s.Confirmation.Cycle >= s.MaxConfirmationCycles {
			s, err = e.append(ctx, s, eventConfirmationForced(s.Confirmation.Cycle, "max confirmation cycles reached"))
			if err != nil {
				return s, err
			}
			return e.finishMeeting(ctx, s)
		}
		fb := resp.Feedback
		if fb == "" {
			fb = "需要修订"
		}
		return e.append(ctx, s, eventConfirmationRejected(s.Confirmation.Cycle, fb, resp.ItemNotes))
	default:
		return s, errUnknownPrincipalDecision
	}
}

func (e *Engine) finishMeeting(ctx context.Context, s meeting.State) (meeting.State, error) {
	ref := "artifacts/minutes.md"
	s, err := e.append(ctx, s, eventArtifactProduced("minutes-1", "markdown", ref))
	if err != nil {
		return s, err
	}
	if err := e.writeArtifactFile(s.ID, ref, []byte(renderMinutes(s))); err != nil {
		return s, err
	}
	return e.append(ctx, s, eventMeetingFinished(meeting.OutcomeCompleted))
}

func confirmationCycle(s meeting.State) int {
	return s.ConfirmationCycle + 1
}

func prepareConfirmationBrief(s meeting.State) event.ConfirmationBrief {
	summary := s.Topic
	source := "consensus"
	if n := len(s.Minutes.Rounds); n > 0 {
		last := s.Minutes.Rounds[n-1]
		summary = last.Summary
		source = fmt.Sprintf("Round %d summary", last.RoundNumber)
	}
	return event.ConfirmationBrief{
		ExecutiveSummary: fmt.Sprintf("%s — 专家团队已达成共识，请 Principal 审阅。", s.Topic),
		Items: []event.ConfirmationItem{{
			Index:       1,
			Title:       "会议结论",
			Description: summary,
			Source:      source,
		}},
	}
}
