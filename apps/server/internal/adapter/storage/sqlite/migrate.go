package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

const currentSchemaVersion = 2

// Migrate applies pending schema migrations.
func Migrate(ctx context.Context, db *sql.DB) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return fmt.Errorf("sqlite migrate: init table: %w", err)
	}

	var version int
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version); err != nil {
		return err
	}
	if version >= currentSchemaVersion {
		return nil
	}

	for v := version + 1; v <= currentSchemaVersion; v++ {
		switch v {
		case 1:
			if _, err := db.ExecContext(ctx, schemaV1); err != nil {
				return fmt.Errorf("sqlite migrate v1: %w", err)
			}
		case 2:
			if _, err := db.ExecContext(ctx, schemaV2); err != nil {
				return fmt.Errorf("sqlite migrate v2: %w", err)
			}
		default:
			return fmt.Errorf("sqlite migrate: unknown version %d", v)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES (?)`, v); err != nil {
			return err
		}
	}
	return nil
}
