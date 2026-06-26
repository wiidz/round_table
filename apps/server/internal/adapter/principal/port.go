package principal

import (
	"context"

	"round_table/apps/server/internal/domain/event"
)

// Decision is the Principal's confirmation response (ADR-0004).
type Decision string

const (
	DecisionApproved Decision = "approved"
	DecisionRejected Decision = "rejected"
)

// Response is the Principal's answer to a Confirmation Brief.
type Response struct {
	Decision  Decision
	Feedback  string
	ItemNotes map[int]string
}

// Port represents the Principal at the Confirmation gate.
type Port interface {
	Confirm(ctx context.Context, meetingID string, brief event.ConfirmationBrief, cycle int) (Response, error)
}
