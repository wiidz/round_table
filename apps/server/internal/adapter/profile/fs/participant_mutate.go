package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"round_table/apps/server/internal/platform/config"
)

// RenameParticipant moves profiles/participants/{old} → {new}.
func (s *Store) RenameParticipant(oldID, newID string) error {
	if err := validateProfileID(oldID); err != nil {
		return err
	}
	if err := config.ValidateParticipantID(newID); err != nil {
		return err
	}
	oldPath := filepath.Join(s.Root, "participants", oldID)
	newPath := filepath.Join(s.Root, "participants", newID)
	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("participant profile %q not found", oldID)
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("participant profile %q already exists", newID)
	}
	return os.Rename(oldPath, newPath)
}

// DeleteParticipant removes profiles/participants/{id}.
func (s *Store) DeleteParticipant(id string) error {
	if err := validateProfileID(id); err != nil {
		return err
	}
	dir := filepath.Join(s.Root, "participants", id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}
