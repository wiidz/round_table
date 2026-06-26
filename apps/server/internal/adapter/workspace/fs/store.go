package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"round_table/apps/server/internal/adapter/workspace"
)

// Store implements workspace.Port on the local filesystem.
type Store struct {
	Root string
}

func NewStore(root string) *Store {
	return &Store{Root: root}
}

func (s *Store) EnsureMeeting(meetingID, topic string) error {
	dir, err := s.meetingDir(meetingID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, sub := range []string{workspace.DirArtifacts, workspace.DirRounds} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return err
		}
	}
	meetingFile := filepath.Join(dir, workspace.FileMeeting)
	if _, err := os.Stat(meetingFile); os.IsNotExist(err) {
		body := fmt.Sprintf("# Meeting\n\n%s\n", topic)
		return os.WriteFile(meetingFile, []byte(body), 0o644)
	}
	return nil
}

func (s *Store) Read(meetingID, relPath string) ([]byte, error) {
	abs, err := s.Resolve(meetingID, relPath)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if os.IsNotExist(err) {
		return nil, workspace.ErrNotFound
	}
	return data, err
}

func (s *Store) Write(meetingID, relPath string, data []byte) error {
	abs, err := s.Resolve(meetingID, relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, data, 0o644)
}

func (s *Store) List(meetingID, relPath string) ([]string, error) {
	abs, err := s.Resolve(meetingID, relPath)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(abs)
	if os.IsNotExist(err) {
		return nil, workspace.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}

func (s *Store) Resolve(meetingID, relPath string) (string, error) {
	root, err := s.meetingDir(meetingID)
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(relPath)
	if clean == "." {
		return root, nil
	}
	abs := filepath.Join(root, clean)
	if !pathWithin(root, abs) {
		return "", workspace.ErrOutsideRoot
	}
	return abs, nil
}

func (s *Store) meetingDir(meetingID string) (string, error) {
	if meetingID == "" || strings.Contains(meetingID, "..") || strings.ContainsAny(meetingID, `/\`) {
		return "", fmt.Errorf("workspace: invalid meeting id")
	}
	root := filepath.Clean(s.Root)
	abs := filepath.Join(root, meetingID)
	if !pathWithin(root, abs) {
		return "", workspace.ErrOutsideRoot
	}
	return abs, nil
}

func pathWithin(root, target string) bool {
	root = filepath.Clean(root)
	target = filepath.Clean(target)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
