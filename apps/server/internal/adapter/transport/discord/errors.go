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
)
