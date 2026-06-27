package bootstrap

import (
	"fmt"

	"round_table/apps/server/internal/adapter/storage"
	"round_table/apps/server/internal/adapter/storage/memory"
	"round_table/apps/server/internal/adapter/storage/sqlite"
	"round_table/apps/server/internal/platform/config"
)

// OpenStorage returns an event store for the configured driver.
func OpenStorage(cfg config.Storage) (storage.Store, error) {
	driver := cfg.Driver
	if driver == "" {
		driver = "sqlite"
	}
	switch driver {
	case "memory":
		return memory.New(), nil
	case "sqlite":
		if cfg.SQLitePath == "" {
			return nil, fmt.Errorf("storage: sqlite_path required")
		}
		return sqlite.Open(cfg.SQLitePath)
	default:
		return nil, fmt.Errorf("storage: unsupported driver %q", driver)
	}
}
