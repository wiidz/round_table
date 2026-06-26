package engine

import (
	"context"

	"round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/domain/meeting"
)

func (e *Engine) maybePrincipalRunningAction(ctx context.Context, s meeting.State) (meeting.State, bool, error) {
	if e.Principal == nil || s.CurrentRound <= 0 {
		return s, false, nil
	}
	action, err := e.Principal.RunningAction(ctx, s.ID, s)
	if err != nil {
		return s, false, err
	}
	switch action.Kind {
	case principal.RunningInterventionForceConsensus:
		if s.IsDeliberation() {
			return s, false, nil
		}
		e.logf("◇ principal force consensus (%s)", action.Reason)
		ns, err := e.append(ctx, s, eventConsensusForced(action.Reason))
		return ns, true, err
	case principal.RunningInterventionForceSynthesis:
		if !s.IsDeliberation() {
			return s, false, nil
		}
		e.logf("◇ principal force synthesis (%s)", action.Reason)
		ns, err := e.append(ctx, s, eventSynthesisForced(action.Reason))
		if err != nil {
			return ns, true, err
		}
		ns, err = e.completeDeliberation(ctx, ns, "principal")
		return ns, true, err
	case principal.RunningInterventionPause:
		e.logf("⏸ principal pause (%s)", action.Reason)
		ns, err := e.append(ctx, s, eventMeetingPaused(action.Reason))
		return ns, true, err
	case principal.RunningInterventionAbort:
		e.logf("■ principal abort (%s)", action.Reason)
		ns, err := e.append(ctx, s, eventMeetingAborted(action.Reason))
		return ns, true, err
	default:
		return s, false, nil
	}
}
