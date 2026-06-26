package engine

import (
	"errors"
	"fmt"

	"round_table/apps/server/internal/domain/meeting"
)

var (
	errMeetingIDRequired      = errors.New("engine: meeting id required")
	errTopicRequired          = errors.New("engine: topic required")
	errNoParticipants         = errors.New("engine: at least one participant required")
	errParticipantIDRequired  = errors.New("engine: participant id required")
	errStatusNotRunnable      = func(st meeting.Status) error {
		return fmt.Errorf("engine: cannot run in status %s", st)
	}
)
