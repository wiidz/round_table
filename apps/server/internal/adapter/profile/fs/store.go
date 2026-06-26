package fs

import (
	"io"
	"os"
	"path/filepath"

	"round_table/apps/server/internal/adapter/profile"
)

// Store implements profile.Port on the local filesystem.
type Store struct {
	Root      string
	Templates string
}

func NewStore(root, templates string) *Store {
	return &Store{Root: root, Templates: templates}
}

func (s *Store) EnsureParticipant(id string) error {
	dir := filepath.Join(s.Root, "participants", id)
	return s.ensureFromTemplates(dir, filepath.Join(s.Templates, "participants"))
}

func (s *Store) EnsurePrincipal(id string) error {
	dir := filepath.Join(s.Root, "principals", id)
	return s.ensureFromTemplates(dir, filepath.Join(s.Templates, "principals"))
}

func (s *Store) EnsureModerator() error {
	dir := filepath.Join(s.Root, "moderator")
	return s.ensureFromTemplates(dir, filepath.Join(s.Templates, "moderator"))
}

func (s *Store) ReadParticipant(id, filename string) ([]byte, error) {
	return s.read(filepath.Join(s.Root, "participants", id, filename))
}

func (s *Store) WriteParticipant(id, filename string, data []byte) error {
	dir := filepath.Join(s.Root, "participants", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, filename), data, 0o644)
}

func (s *Store) ReadPrincipal(id, filename string) ([]byte, error) {
	return s.read(filepath.Join(s.Root, "principals", id, filename))
}

func (s *Store) WritePrincipal(id, filename string, data []byte) error {
	dir := filepath.Join(s.Root, "principals", id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, filename), data, 0o644)
}

func (s *Store) ReadModerator(filename string) ([]byte, error) {
	return s.read(filepath.Join(s.Root, "moderator", filename))
}

func (s *Store) ensureFromTemplates(destDir, templateDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(templateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		target := filepath.Join(destDir, name)
		if _, err := os.Stat(target); err == nil {
			continue
		}
		if err := copyFile(filepath.Join(templateDir, name), target); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, profile.ErrNotFound
	}
	return data, err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
