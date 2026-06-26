package discord

import "errors"

var (
	errNoParticipants      = errors.New("discord: at least one participant required")
	errInvalidParticipant  = errors.New("discord: invalid participant, want id:Role[:Expertise]")
	errMeetTopicRequired   = errors.New("discord: meet topic required")
	errMeetModeFlag        = errors.New("discord: -mode requires a value")
	errNotBoundPrincipal   = errors.New("discord: principal bind required")
	errNotScopePrincipal   = errors.New("discord: only the bound principal can start a meeting")
	errChannelMeetingBusy  = errors.New("discord: a meeting is already running in this channel")
	errSetupPending        = errors.New("discord: meet setup already pending in this channel")
	errSetupReplyEmpty     = errors.New("discord: empty setup reply")
	errSetupReplyUnrecognized = errors.New("discord: unrecognized setup reply")
	errSetupInvalidChoice     = errors.New("discord: invalid menu choice")
	errSetupInvalidMode    = errors.New("discord: mode must be decision or deliberation")
	errSetupInvalidConfirmation = errors.New("discord: confirmation must be skip or required")
	errSetupInvalidRounds  = errors.New("discord: rounds must be a positive integer")
	errSetupInvalidMinRounds = errors.New("discord: min-rounds must be a positive integer")
	errSetupInvalidFree    = errors.New("discord: free must be a non-negative integer")
)
