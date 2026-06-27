package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"round_table/apps/server/internal/platform/config"
)

var _ config.SettingsStore = (*Store)(nil)

// GetAllSettings implements config.SettingsStore.
func (s *Store) GetAllSettings(ctx context.Context) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM app_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var key, val string
		if err := rows.Scan(&key, &val); err != nil {
			return nil, err
		}
		out[key] = val
	}
	return out, rows.Err()
}

// SetSettings implements config.SettingsStore.
func (s *Store) SetSettings(ctx context.Context, updates map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for key, val := range updates {
		if !config.IsPersistableSettingKey(key) {
			continue
		}
		_, err := tx.ExecContext(ctx, `
INSERT INTO app_settings (key, value, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
			key, val, now,
		)
		if err != nil {
			return fmt.Errorf("set setting %q: %w", key, err)
		}
	}
	return tx.Commit()
}

// DeleteSettings removes keys (used in tests).
func (s *Store) DeleteSettings(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	for _, key := range keys {
		if _, err := s.db.ExecContext(ctx, `DELETE FROM app_settings WHERE key = ?`, key); err != nil {
			return err
		}
	}
	return nil
}

// DB exposes the underlying handle for migrate CLI tests.
func (s *Store) DB() *sql.DB {
	return s.db
}
