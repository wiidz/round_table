package engine

import (
	"context"

	"round_table/apps/server/internal/adapter/participant"
	"round_table/apps/server/internal/adapter/storage"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/meeting"
)

// Engine orchestrates Meeting lifecycle (Constitution step 5).
type Engine struct {
	Store     storage.Store
	Strategy  consensus.Strategy
	Participant participant.Port
}

// LoadState replays events for a meeting.
func (e *Engine) LoadState(ctx context.Context, meetingID string) (meeting.State, error) {
	events, err := e.Store.List(ctx, meetingID)
	if err != nil {
		return meeting.State{}, err
	}
	return meeting.Fold(meetingID, events)
}
