package event

import "time"

// Type identifies a domain event (ADR-0003 v0.1).
type Type string

const (
	TypeMeetingCreated         Type = "MeetingCreated"
	TypeParticipantInvited     Type = "ParticipantInvited"
	TypeRoundStarted           Type = "RoundStarted"
	TypeParticipantResponded   Type = "ParticipantResponded"
	TypeRoundCompleted         Type = "RoundCompleted"
	TypeModeratorSummarized    Type = "ModeratorSummarized"
	TypeFreeDialogueStarted    Type = "FreeDialogueStarted"
	TypeFreeDialogueQuestion   Type = "FreeDialogueQuestionAsked"
	TypeFreeDialogueAnswer     Type = "FreeDialogueAnswered"
	TypeFreeDialogueCompleted  Type = "FreeDialogueCompleted"
	TypeConsensusReached       Type = "ConsensusReached"
	TypeConsensusVetoed        Type = "ConsensusVetoed"
	TypeConsensusForced        Type = "ConsensusForced"
	TypeConfirmationPrepared   Type = "ConfirmationPrepared"
	TypeConfirmationPresented  Type = "ConfirmationPresented"
	TypeConfirmationApproved   Type = "ConfirmationApproved"
	TypeConfirmationRejected   Type = "ConfirmationRejected"
	TypeConfirmationSkipped    Type = "ConfirmationSkipped"
	TypeConfirmationForced     Type = "ConfirmationForced"
	TypeMeetingPaused          Type = "MeetingPaused"
	TypeMeetingResumed         Type = "MeetingResumed"
	TypeMeetingFinished        Type = "MeetingFinished"
	TypeArtifactProduced       Type = "ArtifactProduced"
	TypeActionItemCreated      Type = "ActionItemCreated"
)

// Actor is the envelope actor field (ADR-0003).
type Actor string

const (
	ActorPrincipal   Actor = "principal"
	ActorModerator   Actor = "moderator"
	ActorParticipant Actor = "participant"
	ActorSystem      Actor = "system"
)

// Stance is a participant's structured agreement signal (ADR-0002).
type Stance string

const (
	StanceAgree   Stance = "agree"
	StanceObject  Stance = "object"
	StanceAbstain Stance = "abstain"
	StanceNone    Stance = "none"
)

// Envelope is the immutable event record (ADR-0003).
type Envelope struct {
	ID         string
	MeetingID  string
	Sequence   int
	Type       Type
	Version    int
	Payload    []byte
	OccurredAt time.Time
	Actor      Actor
}
