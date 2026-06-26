package engine

import (
	"context"
	"fmt"

	"round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/domain/meeting"
)

func (e *Engine) advancePaused(ctx context.Context, s meeting.State) (meeting.State, error) {
	if e.Principal == nil {
		return s, fmt.Errorf("engine: meeting %s paused with no principal port to resume", s.ID)
	}
	action, err := e.Principal.PausedAction(ctx, s.ID, s)
	if err != nil {
		return s, err
	}
	switch action.Kind {
	case principal.RunningInterventionResume:
		e.logf("▶ principal resumed meeting")
		return e.append(ctx, s, eventMeetingResumed())
	case principal.RunningInterventionAbort:
		e.logf("■ principal abort while paused (%s)", action.Reason)
		return e.append(ctx, s, eventMeetingAborted(action.Reason))
	default:
		return s, fmt.Errorf("engine: meeting %s paused without principal resume or abort", s.ID)
	}
}
