package engine

import (
	"errors"
	"fmt"

	"round_table/apps/server/internal/domain/meeting"
)

var (
	errMeetingIDRequired     = errors.New("engine: meeting id required")
	errTopicRequired         = errors.New("engine: topic required")
	errNoParticipants        = errors.New("engine: at least one participant required")
	errParticipantIDRequired = errors.New("engine: participant id required")
	errPrincipalRequired     = errors.New("engine: principal port required for confirmation")
	errNoConfirmationBrief   = errors.New("engine: confirmation brief missing")
	errUnknownPrincipalDecision = errors.New("engine: unknown principal decision")
)

func errStatusNotRunnable(st meeting.Status) error {
	return fmt.Errorf("engine: cannot run in status %s", st)
}
