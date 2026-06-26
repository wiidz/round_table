package engine

import (
	"context"

	"round_table/apps/server/internal/domain/meeting"
)

// LoadState replays events for a meeting.
func (e *Engine) LoadState(ctx context.Context, meetingID string) (meeting.State, error) {
	events, err := e.Store.List(ctx, meetingID)
	if err != nil {
		return meeting.State{}, err
	}
	return meeting.Fold(meetingID, events)
}

// CreateMeetingInput seeds a new meeting through MeetingCreated + invites.
type CreateMeetingInput struct {
	MeetingID           string
	Topic               string
	Goal                string
	ConfirmationMode    string
	MaxRoundsPerSegment      int
	FreeDialogueMaxQuestions *int // nil = default (1); explicit 0 disables
	Participants             []ParticipantInput
}

// ParticipantInput registers one expert for the meeting.
type ParticipantInput struct {
	ID        string
	Role      string
	Expertise string
	Goal      string
}

// CreateMeeting appends MeetingCreated and ParticipantInvited events.
func (e *Engine) CreateMeeting(ctx context.Context, in CreateMeetingInput) (meeting.State, error) {
	if in.MeetingID == "" {
		return meeting.State{}, errMeetingIDRequired
	}
	if in.Topic == "" {
		return meeting.State{}, errTopicRequired
	}
	if len(in.Participants) == 0 {
		return meeting.State{}, errNoParticipants
	}

	mode := in.ConfirmationMode
	if mode == "" {
		mode = meeting.ConfirmationModeSkip
	}

	freeQuestions := 1
	if in.FreeDialogueMaxQuestions != nil {
		freeQuestions = *in.FreeDialogueMaxQuestions
	}

	s, err := e.append(ctx, meeting.NewState(in.MeetingID), eventMeetingCreated(in.Topic, in.Goal, mode, in.MaxRoundsPerSegment, freeQuestions))
	if err != nil {
		return s, err
	}

	for _, p := range in.Participants {
		if p.ID == "" {
			return s, errParticipantIDRequired
		}
		s, err = e.append(ctx, s, eventParticipantInvited(p))
		if err != nil {
			return s, err
		}
	}
	return s, nil
}

// Run drives the meeting from Preparing through Completed.
func (e *Engine) Run(ctx context.Context, meetingID string) (meeting.State, error) {
	s, err := e.LoadState(ctx, meetingID)
	if err != nil {
		return s, err
	}
	if s.Status == meeting.StatusCompleted || s.Status == meeting.StatusArchived {
		return s, nil
	}

	for {
		if err := ctx.Err(); err != nil {
			return s, err
		}
		switch s.Status {
		case meeting.StatusCompleted, meeting.StatusArchived:
			return s, nil
		case meeting.StatusPreparing:
			s, err = e.startRound(ctx, s)
		case meeting.StatusRunning:
			if s.InFreeDialogue {
				s, err = e.advanceFreeDialogue(ctx, s)
			} else {
				s, err = e.advanceRunning(ctx, s)
			}
		case meeting.StatusConsensus:
			s, err = e.afterConsensus(ctx, s)
		case meeting.StatusConfirmation:
			s, err = e.advanceConfirmation(ctx, s)
		default:
			return s, errStatusNotRunnable(s.Status)
		}
		if err != nil {
			return s, err
		}
	}
}
