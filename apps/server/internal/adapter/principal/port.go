package principal

import (
	"context"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

// Decision is the Principal's confirmation response (ADR-0004).
type Decision string

const (
	DecisionApproved Decision = "approved"
	DecisionRejected Decision = "rejected"

	// Limit fallback choices (ADR-0004 §6) when max_confirmation_cycles is reached.
	DecisionLimitForceApprove Decision = "limit_force_approve"
	DecisionLimitContinue     Decision = "limit_continue"
	DecisionLimitAbort        Decision = "limit_abort"
)

// Response is the Principal's answer to a Confirmation Brief.
type Response struct {
	Decision  Decision
	Feedback  string
	ItemNotes map[int]string
}

// RunningInterventionKind is a Principal action at Participant turn boundaries (ADR-0005 §6).
type RunningInterventionKind string

const (
	RunningInterventionNone           RunningInterventionKind = ""
	RunningInterventionForceConsensus RunningInterventionKind = "force_consensus"
	RunningInterventionForceSynthesis RunningInterventionKind = "force_synthesis"
	RunningInterventionPause          RunningInterventionKind = "pause"
	RunningInterventionAbort          RunningInterventionKind = "abort"
	RunningInterventionResume         RunningInterventionKind = "resume"
)

// RunningIntervention is a Principal request during Status=Running.
type RunningIntervention struct {
	Kind   RunningInterventionKind
	Reason string
}

// FreeDialogueQuestionRequest is a Principal question queued for free dialogue.
type FreeDialogueQuestionRequest struct {
	Question   string
	AnswererID string // optional; empty uses the current pair's answerer
}

// Port represents the Principal at Confirmation and optional Running turn boundaries.
type Port interface {
	Confirm(ctx context.Context, meetingID string, brief event.ConfirmationBrief, cycle int) (Response, error)
	// RunningAction returns a turn-boundary intervention while Status=Running. Default: none.
	RunningAction(ctx context.Context, meetingID string, s meeting.State) (RunningIntervention, error)
	// PausedAction returns resume or abort while Status=Paused. Default: resume immediately.
	PausedAction(ctx context.Context, meetingID string, s meeting.State) (RunningIntervention, error)
	// FreeDialogueQuestion returns a Principal question queued at a free-dialogue turn boundary.
	// ok=false when none is queued.
	FreeDialogueQuestion(ctx context.Context, meetingID string, s meeting.State) (FreeDialogueQuestionRequest, bool, error)
}
