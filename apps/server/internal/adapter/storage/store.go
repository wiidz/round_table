package storage

import (
	"context"

	"round_table/apps/server/internal/domain/event"
)

// Store append-only event persistence (ADR-0003).
type Store interface {
	Append(ctx context.Context, env event.Envelope) error
	List(ctx context.Context, meetingID string) ([]event.Envelope, error)
}
