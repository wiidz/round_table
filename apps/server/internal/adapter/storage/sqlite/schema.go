package sqlite

const schemaV1 = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS events (
	id TEXT NOT NULL,
	meeting_id TEXT NOT NULL,
	sequence INTEGER NOT NULL,
	type TEXT NOT NULL,
	version INTEGER NOT NULL,
	payload BLOB NOT NULL,
	occurred_at TEXT NOT NULL,
	actor TEXT NOT NULL,
	PRIMARY KEY (meeting_id, sequence)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_events_id ON events(id);
CREATE INDEX IF NOT EXISTS idx_events_meeting_seq ON events(meeting_id, sequence);

CREATE TABLE IF NOT EXISTS meeting_index (
	meeting_id TEXT PRIMARY KEY,
	topic TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_meeting_index_updated ON meeting_index(updated_at DESC);
`

const schemaV2 = `
CREATE TABLE IF NOT EXISTS app_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
`
