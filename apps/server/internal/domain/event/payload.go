package event

// MeetingCreatedPayload is the v1 payload for MeetingCreated.
type MeetingCreatedPayload struct {
	Topic                 string       `json:"topic"`
	Agenda                []AgendaItem `json:"agenda,omitempty"`
	ConsensusStrategy     string       `json:"consensus_strategy,omitempty"`
	ConfirmationMode      string       `json:"confirmation_mode"`
	MaxRoundsPerSegment   int          `json:"max_rounds_per_segment"`
	MaxConfirmationCycles int          `json:"max_confirmation_cycles"`
}

// AgendaItem is a discussion objective entry.
type AgendaItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ParticipantInvitedPayload is the v1 payload for ParticipantInvited.
type ParticipantInvitedPayload struct {
	ParticipantID string `json:"participant_id"`
	Role          string `json:"role"`
	Expertise     string `json:"expertise,omitempty"`
	Goal          string `json:"goal,omitempty"`
}

// RoundStartedPayload is the v1 payload for RoundStarted.
type RoundStartedPayload struct {
	RoundNumber int      `json:"round_number"`
	Order       []string `json:"order"`
}

// ParticipantRespondedPayload is the v1 payload for ParticipantResponded.
type ParticipantRespondedPayload struct {
	ParticipantID string `json:"participant_id"`
	RoundNumber   int    `json:"round_number"`
	Content       string `json:"content"`
	Stance        Stance `json:"stance"`
	ObjectReason  string `json:"object_reason,omitempty"`
}

// RoundCompletedPayload is the v1 payload for RoundCompleted.
type RoundCompletedPayload struct {
	RoundNumber int    `json:"round_number"`
	Summary     string `json:"summary"`
}

// DissentingOpinion records minority objection when consensus still passes.
type DissentingOpinion struct {
	ParticipantID string `json:"participant_id"`
	Reason        string `json:"reason"`
}

// ConsensusReachedPayload is the v1 payload for ConsensusReached.
type ConsensusReachedPayload struct {
	Strategy   string              `json:"strategy"`
	Dissent    []DissentingOpinion `json:"dissent,omitempty"`
	ResolvedBy string              `json:"resolved_by"` // strategy | moderator | principal
}

// ConsensusVetoedPayload is the v1 payload for ConsensusVetoed.
type ConsensusVetoedPayload struct {
	Reason string `json:"reason"`
}

// ConsensusForcedPayload is the v1 payload for ConsensusForced.
type ConsensusForcedPayload struct {
	Reason string `json:"reason"`
}

// ConfirmationItem is one numbered item in a Confirmation Brief.
type ConfirmationItem struct {
	Index       int    `json:"index"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"`
}

// ConfirmationBrief is presented to the Principal for review.
type ConfirmationBrief struct {
	ExecutiveSummary string             `json:"executive_summary"`
	Items            []ConfirmationItem `json:"items"`
}

// ConfirmationPreparedPayload is the v1 payload for ConfirmationPrepared.
type ConfirmationPreparedPayload struct {
	Cycle int               `json:"cycle"`
	Brief ConfirmationBrief `json:"brief"`
}

// ConfirmationPresentedPayload is the v1 payload for ConfirmationPresented.
type ConfirmationPresentedPayload struct {
	Cycle int `json:"cycle"`
}

// ConfirmationApprovedPayload is the v1 payload for ConfirmationApproved.
type ConfirmationApprovedPayload struct {
	Cycle     int              `json:"cycle"`
	ItemNotes map[int]string   `json:"item_notes,omitempty"`
}

// ConfirmationRejectedPayload is the v1 payload for ConfirmationRejected.
type ConfirmationRejectedPayload struct {
	Cycle     int            `json:"cycle"`
	Feedback  string         `json:"feedback"`
	ItemNotes map[int]string `json:"item_notes,omitempty"`
}

// ConfirmationSkippedPayload is the v1 payload for ConfirmationSkipped.
type ConfirmationSkippedPayload struct {
	Reason string `json:"reason"`
}

// ConfirmationForcedPayload is the v1 payload for ConfirmationForced.
type ConfirmationForcedPayload struct {
	Cycle  int    `json:"cycle"`
	Reason string `json:"reason"`
}

// MeetingPausedPayload is the v1 payload for MeetingPaused.
type MeetingPausedPayload struct {
	Reason string `json:"reason,omitempty"`
}

// MeetingFinishedPayload is the v1 payload for MeetingFinished.
type MeetingFinishedPayload struct {
	Outcome string `json:"outcome,omitempty"` // completed | aborted
}

// ArtifactProducedPayload is the v1 payload for ArtifactProduced.
type ArtifactProducedPayload struct {
	ArtifactID string `json:"artifact_id"`
	Type       string `json:"type"`
	Ref        string `json:"ref"`
}

// ActionItemCreatedPayload is the v1 payload for ActionItemCreated.
type ActionItemCreatedPayload struct {
	ActionItemID string `json:"action_item_id"`
	Assignee     string `json:"assignee,omitempty"`
	Description  string `json:"description"`
}
