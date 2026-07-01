package storage

import (
	"context"

	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/event"
)

// Store append-only event persistence (ADR-0003).
type Store interface {
	Append(ctx context.Context, env event.Envelope) error
	List(ctx context.Context, meetingID string) ([]event.Envelope, error)
}

// MeetingCatalog lists meetings for the web UI (SQLite-backed).
type MeetingCatalog interface {
	ListMeetingsPage(ctx context.Context, page, pageSize int) (workspace.PaginatedMeetings, error)
}

// MeetingDeleter removes persisted meeting events and catalog rows.
type MeetingDeleter interface {
	DeleteMeeting(ctx context.Context, meetingID string) error
}
