package participant

import (
	"context"

	"round_table/apps/server/internal/adapter/model"
)

// Response is a participant's turn output (ADR-0003).
type Response struct {
	ParticipantID string
	Content       string
	Stance        string // agree | object | abstain | none
	ObjectReason  string
	Model         string
	Usage         model.Usage
}

// Port invokes a participant when invited by the Moderator.
type Port interface {
	Respond(ctx context.Context, meetingID, participantID string, prompt string) (Response, error)
}
