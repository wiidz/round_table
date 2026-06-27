package fs

import (
	"os"
	"path/filepath"

	"round_table/apps/server/internal/platform/config"
)

// PruneMisplacedBotProfiles removes participant profile dirs keyed by Discord application ids.
func (s *Store) PruneMisplacedBotProfiles() error {
	root := filepath.Join(s.Root, "participants")
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		if !config.IsMisplacedBotProfileID(id) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, id)); err != nil {
			return err
		}
	}
	return nil
}
