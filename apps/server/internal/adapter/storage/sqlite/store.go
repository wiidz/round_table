package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"round_table/apps/server/internal/adapter/storage"
	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

// Store is a SQLite-backed append-only event store (ADR-0003).
type Store struct {
	db *sql.DB
}

// Open opens (or creates) a SQLite database at path and runs migrations.
func Open(path string) (*Store, error) {
	path = filepath.Clean(path)
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("sqlite: mkdir %s: %w", dir, err)
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := Migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

var _ storage.Store = (*Store)(nil)
var _ storage.MeetingCatalog = (*Store)(nil)

// Append implements storage.Store.
func (s *Store) Append(ctx context.Context, env event.Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	err = tx.QueryRowContext(ctx,
		`SELECT 1 FROM events WHERE meeting_id = ? AND sequence = ?`,
		env.MeetingID, env.Sequence,
	).Scan(&exists)
	if err == nil {
		return fmt.Errorf("duplicate sequence %d for meeting %s", env.Sequence, env.MeetingID)
	}
	if err != sql.ErrNoRows {
		return err
	}

	occurred := env.OccurredAt.UTC().Format(time.RFC3339Nano)
	_, err = tx.ExecContext(ctx, `
INSERT INTO events (id, meeting_id, sequence, type, version, payload, occurred_at, actor)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		env.ID, env.MeetingID, env.Sequence, string(env.Type), env.Version, env.Payload, occurred, string(env.Actor),
	)
	if err != nil {
		return err
	}

	if err := upsertMeetingIndex(ctx, tx, env, occurred); err != nil {
		return err
	}

	return tx.Commit()
}

// List implements storage.Store.
func (s *Store) List(ctx context.Context, meetingID string) ([]event.Envelope, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, meeting_id, sequence, type, version, payload, occurred_at, actor
FROM events WHERE meeting_id = ? ORDER BY sequence ASC`, meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []event.Envelope
	for rows.Next() {
		var env event.Envelope
		var typ, actor, occurred string
		if err := rows.Scan(&env.ID, &env.MeetingID, &env.Sequence, &typ, &env.Version, &env.Payload, &occurred, &actor); err != nil {
			return nil, err
		}
		env.Type = event.Type(typ)
		env.Actor = event.Actor(actor)
		t, err := time.Parse(time.RFC3339Nano, occurred)
		if err != nil {
			t, err = time.Parse(time.RFC3339, occurred)
			if err != nil {
				return nil, fmt.Errorf("parse occurred_at %q: %w", occurred, err)
			}
		}
		env.OccurredAt = t
		out = append(out, env)
	}
	return out, rows.Err()
}

// ListMeetingsPage implements storage.MeetingCatalog.
func (s *Store) ListMeetingsPage(ctx context.Context, page, pageSize int) (workspace.PaginatedMeetings, error) {
	if err := ctx.Err(); err != nil {
		return workspace.PaginatedMeetings{}, err
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}

	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM meeting_index`).Scan(&total); err != nil {
		return workspace.PaginatedMeetings{}, err
	}

	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx, `
SELECT meeting_id, topic, status, updated_at
FROM meeting_index
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return workspace.PaginatedMeetings{}, err
	}
	defer rows.Close()

	meetings := make([]workspace.MeetingIndex, 0, pageSize)
	for rows.Next() {
		var idx workspace.MeetingIndex
		var updated string
		if err := rows.Scan(&idx.ID, &idx.Topic, &idx.Status, &updated); err != nil {
			return workspace.PaginatedMeetings{}, err
		}
		idx.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		if idx.UpdatedAt.IsZero() {
			idx.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		}
		meetings = append(meetings, idx)
	}
	if err := rows.Err(); err != nil {
		return workspace.PaginatedMeetings{}, err
	}

	return workspace.PaginatedMeetings{
		Meetings: meetings,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func upsertMeetingIndex(ctx context.Context, tx *sql.Tx, env event.Envelope, occurred string) error {
	switch env.Type {
	case event.TypeMeetingCreated:
		var p event.MeetingCreatedPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("MeetingCreated payload: %w", err)
		}
		_, err := tx.ExecContext(ctx, `
INSERT INTO meeting_index (meeting_id, topic, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(meeting_id) DO UPDATE SET
  topic = excluded.topic,
  updated_at = excluded.updated_at`,
			env.MeetingID, p.Topic, string(meeting.StatusPreparing), occurred, occurred,
		)
		return err
	case event.TypeMeetingFinished:
		status := string(meeting.StatusCompleted)
		var p event.MeetingFinishedPayload
		if err := json.Unmarshal(env.Payload, &p); err == nil && strings.EqualFold(p.Outcome, meeting.OutcomeAborted) {
			status = string(meeting.StatusCompleted)
		}
		_, err := tx.ExecContext(ctx, `
UPDATE meeting_index SET status = ?, updated_at = ? WHERE meeting_id = ?`,
			status, occurred, env.MeetingID,
		)
		return err
	case event.TypeMeetingPaused:
		_, err := tx.ExecContext(ctx, `
UPDATE meeting_index SET status = ?, updated_at = ? WHERE meeting_id = ?`,
			string(meeting.StatusPaused), occurred, env.MeetingID,
		)
		return err
	case event.TypeMeetingResumed:
		_, err := tx.ExecContext(ctx, `
UPDATE meeting_index SET status = ?, updated_at = ? WHERE meeting_id = ?`,
			string(meeting.StatusRunning), occurred, env.MeetingID,
		)
		return err
	case event.TypeRoundStarted:
		_, err := tx.ExecContext(ctx, `
UPDATE meeting_index SET status = ?, updated_at = ? WHERE meeting_id = ?`,
			string(meeting.StatusRunning), occurred, env.MeetingID,
		)
		return err
	default:
		_, err := tx.ExecContext(ctx, `
UPDATE meeting_index SET updated_at = ? WHERE meeting_id = ?`, occurred, env.MeetingID)
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
}
