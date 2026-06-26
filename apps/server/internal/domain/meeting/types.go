package meeting

import "round_table/apps/server/internal/domain/event"

const (
	ConfirmationModeRequired = "required"
	ConfirmationModeSkip     = "skip"

	OutcomeCompleted = "completed"
	OutcomeAborted   = "aborted"
)

// ParticipantState is the folded projection of an invited participant.
type ParticipantState struct {
	ID        string
	Role      string
	Expertise string
	Goal      string
}

// RoundResponse stores one participant turn in a round.
type RoundResponse struct {
	Content      string
	Stance       event.Stance
	ObjectReason string
}

// ConsensusState is set when Consensus is reached.
type ConsensusState struct {
	Strategy   string
	ResolvedBy string
	Dissent    []event.DissentingOpinion
}

// ConfirmationState tracks the Principal confirmation gate.
type ConfirmationState struct {
	Cycle      int
	Brief      event.ConfirmationBrief
	Approved   bool
	ItemNotes  map[int]string
}

// RoundSummary is one round entry in Minutes draft.
type RoundSummary struct {
	RoundNumber int
	Summary     string
}

// MinutesDraft accumulates structured minutes during the Meeting.
type MinutesDraft struct {
	Rounds []RoundSummary
}

// ArtifactRef references a produced artifact.
type ArtifactRef struct {
	ID   string
	Type string
	Ref  string
}

// ActionItem is a follow-up task from the Meeting.
type ActionItem struct {
	ID          string
	Assignee    string
	Description string
}
