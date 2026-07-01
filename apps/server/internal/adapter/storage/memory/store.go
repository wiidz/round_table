package memory

import (
	"context"
	"fmt"
	"sync"

	"round_table/apps/server/internal/adapter/storage"
	"round_table/apps/server/internal/domain/event"
)

// Store is an in-memory append-only event store for tests and v0.1 dev.
type Store struct {
	mu     sync.Mutex
	events map[string][]event.Envelope
}

// New returns an empty memory store.
func New() *Store {
	return &Store{events: make(map[string][]event.Envelope)}
}

var _ storage.Store = (*Store)(nil)
var _ storage.MeetingDeleter = (*Store)(nil)

// Append implements storage.Store.
func (s *Store) Append(ctx context.Context, env event.Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.events[env.MeetingID]
	for _, e := range list {
		if e.Sequence == env.Sequence {
			return fmt.Errorf("duplicate sequence %d for meeting %s", env.Sequence, env.MeetingID)
		}
	}
	s.events[env.MeetingID] = append(list, env)
	return nil
}

// List implements storage.Store.
func (s *Store) List(ctx context.Context, meetingID string) ([]event.Envelope, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := append([]event.Envelope(nil), s.events[meetingID]...)
	return out, nil
}

// DeleteMeeting implements storage.MeetingDeleter.
func (s *Store) DeleteMeeting(ctx context.Context, meetingID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.events, meetingID)
	return nil
}
