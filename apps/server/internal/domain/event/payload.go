package event

// MeetingCreatedPayload is the v1 payload for MeetingCreated.
type MeetingCreatedPayload struct {
	Topic                    string       `json:"topic"`
	Goal                     string       `json:"goal,omitempty"`
	MeetingMode              string       `json:"meeting_mode,omitempty"`
	Agenda                   []AgendaItem `json:"agenda,omitempty"`
	ConsensusStrategy     string       `json:"consensus_strategy,omitempty"`
	ConfirmationMode      string       `json:"confirmation_mode"`
	MaxRoundsPerSegment        int          `json:"max_rounds_per_segment"`
	MinRoundsBeforeSynthesis   *int         `json:"min_rounds_before_synthesis,omitempty"`
	MaxConfirmationCycles      int          `json:"max_confirmation_cycles"`
	FreeDialogueMaxQuestions   *int         `json:"free_dialogue_max_questions,omitempty"`
}

// AgendaItem is a discussion objective entry.
type AgendaItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// TokenUsage records LLM token consumption for one participant turn.
type TokenUsage struct {
	Model            string `json:"model,omitempty"`
	Phase            string `json:"phase"`
	ParticipantID    string `json:"participant_id"`
	RoundNumber      int    `json:"round_number,omitempty"`
	QuestionIndex    int    `json:"question_index,omitempty"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
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
	Stance        Stance     `json:"stance"`
	ObjectReason  string     `json:"object_reason,omitempty"`
	TokenUsage    *TokenUsage `json:"token_usage,omitempty"`
}

// RoundCompletedPayload is the v1 payload for RoundCompleted.
type RoundCompletedPayload struct {
	RoundNumber int    `json:"round_number"`
	Summary     string `json:"summary"`
}

// ModeratorSummarizedPayload is the v1 payload for ModeratorSummarized.
type ModeratorSummarizedPayload struct {
	RoundNumber int    `json:"round_number"`
	Summary     string `json:"summary"`
}

// DeliberationReadinessCheckedPayload records Moderator synthesis-readiness judgment (deliberation mode).
type DeliberationReadinessCheckedPayload struct {
	RoundNumber int         `json:"round_number"`
	Ready       bool        `json:"ready"`
	Rationale   string      `json:"rationale,omitempty"`
	Gaps        []string    `json:"gaps,omitempty"`
	TokenUsage  *TokenUsage `json:"token_usage,omitempty"`
}

// FreeDialogueStartedPayload marks Q&A after a debate round (fixed after Round 1).
type FreeDialogueStartedPayload struct {
	AfterRound   int `json:"after_round"`
	MaxQuestions int `json:"max_questions"`
}

// FreeDialogueQuestionAskedPayload records a question in free dialogue.
type FreeDialogueQuestionAskedPayload struct {
	AskerID       string `json:"asker_id"`
	AnswererID    string `json:"answerer_id"`
	QuestionIndex int        `json:"question_index"`
	Content       string     `json:"content"`
	TokenUsage    *TokenUsage `json:"token_usage,omitempty"`
}

// FreeDialogueAnsweredPayload records an answer in free dialogue.
type FreeDialogueAnsweredPayload struct {
	AskerID       string `json:"asker_id"`
	AnswererID    string `json:"answerer_id"`
	QuestionIndex int    `json:"question_index"`
	Question      string     `json:"question"`
	Answer        string     `json:"answer"`
	TokenUsage    *TokenUsage `json:"token_usage,omitempty"`
}

// FreeDialogueCompletedPayload closes the free dialogue segment.
type FreeDialogueCompletedPayload struct {
	AfterRound int    `json:"after_round"`
	Summary    string `json:"summary"`
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

// SynthesisCompletedPayload is the v1 payload for SynthesisCompleted (deliberation mode).
type SynthesisCompletedPayload struct {
	Summary       string                         `json:"summary"`
	OpenQuestions []string                       `json:"open_questions,omitempty"`
	ResolvedBy    string                         `json:"resolved_by,omitempty"` // readiness | synthesis | max_rounds | principal
	TokenUsage    *TokenUsage                    `json:"token_usage,omitempty"`
	Sections      []SynthesisAgendaSectionPayload  `json:"sections,omitempty"`
	CrossCutting  *SynthesisCrossCuttingPayload    `json:"cross_cutting,omitempty"`
}

// SynthesisAgendaSectionPayload is one agenda item's synthesized content.
type SynthesisAgendaSectionPayload struct {
	AgendaID      string   `json:"agenda_id"`
	Summary       []string `json:"summary,omitempty"`
	Decisions     []string `json:"decisions,omitempty"`
	OpenQuestions []string `json:"open_questions,omitempty"`
}

// SynthesisCrossCuttingPayload holds synthesis not tied to a single agenda item.
type SynthesisCrossCuttingPayload struct {
	Decisions     []string `json:"decisions,omitempty"`
	OpenQuestions []string `json:"open_questions,omitempty"`
}

// SynthesisForcedPayload records Principal intent to stop debate and synthesize now (deliberation).
type SynthesisForcedPayload struct {
	Reason string `json:"reason,omitempty"`
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
	ExecutiveSummary    string             `json:"executive_summary"`
	Items               []ConfirmationItem `json:"items"`
	LimitFallback       bool               `json:"limit_fallback,omitempty"`
	LimitRejectFeedback string             `json:"limit_reject_feedback,omitempty"`
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
	Cycle      int            `json:"cycle"`
	Feedback   string         `json:"feedback"`
	ItemNotes  map[int]string `json:"item_notes,omitempty"`
	ResetCycle bool           `json:"reset_cycle,omitempty"`
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
