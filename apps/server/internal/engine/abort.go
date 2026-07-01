package engine

import (
	"context"
	"errors"

	"round_table/apps/server/internal/domain/meeting"
)

var (
	// ErrMeetingAlreadyFinished is returned when AbortMeeting targets a terminal meeting.
	ErrMeetingAlreadyFinished = errors.New("engine: meeting already finished")
	// ErrMeetingNotAbortable is returned when the meeting has no abortable lifecycle state.
	ErrMeetingNotAbortable = errors.New("engine: meeting cannot be aborted")
)

// IsAbortableStatus reports whether a folded meeting may be aborted.
func IsAbortableStatus(st meeting.Status) bool {
	switch st {
	case meeting.StatusPreparing, meeting.StatusRunning, meeting.StatusPaused,
		meeting.StatusConsensus, meeting.StatusConfirmation:
		return true
	default:
		return false
	}
}

// AbortMeeting appends partial minutes and MeetingFinished(outcome=aborted).
// Idempotent callers should treat ErrMeetingAlreadyFinished as success.
func (e *Engine) AbortMeeting(ctx context.Context, meetingID, reason string) (meeting.State, error) {
	if meetingID == "" {
		return meeting.State{}, errMeetingIDRequired
	}
	s, err := e.LoadState(ctx, meetingID)
	if err != nil {
		return s, err
	}
	if s.Status == meeting.StatusCompleted || s.Status == meeting.StatusArchived {
		return s, ErrMeetingAlreadyFinished
	}
	if !IsAbortableStatus(s.Status) {
		return s, ErrMeetingNotAbortable
	}
	if reason == "" {
		reason = "会议已中止"
	}
	e.logf("■ abort meeting (%s)", reason)
	return e.abortMeeting(ctx, s, reason)
}
