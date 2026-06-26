package fs

import (
	"os"
	"path/filepath"
	"strings"

	"round_table/apps/server/internal/adapter/knowledge"
)

// Store implements knowledge.Port on the local filesystem.
type Store struct {
	Root      string
	Templates string
}

func NewStore(root, templates string) *Store {
	return &Store{Root: root, Templates: templates}
}

func (s *Store) Ensure(scope knowledge.Scope, ownerID string) error {
	dir, err := s.scopeDir(scope, ownerID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, knowledge.DirMemoryLogs), 0o755); err != nil {
		return err
	}
	memPath := filepath.Join(dir, knowledge.FileMemory)
	if _, err := os.Stat(memPath); os.IsNotExist(err) {
		if err := copyIfExists(filepath.Join(s.Templates, knowledge.FileMemory), memPath); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ReadMemory(scope knowledge.Scope, ownerID string) ([]byte, error) {
	dir, err := s.scopeDir(scope, ownerID)
	if err != nil {
		return nil, err
	}
	return s.read(filepath.Join(dir, knowledge.FileMemory))
}

func (s *Store) WriteMemory(scope knowledge.Scope, ownerID string, data []byte) error {
	if err := s.Ensure(scope, ownerID); err != nil {
		return err
	}
	dir, err := s.scopeDir(scope, ownerID)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, knowledge.FileMemory), data, 0o644)
}

func (s *Store) AppendDailyLog(scope knowledge.Scope, ownerID, date string, data []byte) error {
	if err := s.Ensure(scope, ownerID); err != nil {
		return err
	}
	dir, err := s.scopeDir(scope, ownerID)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, knowledge.DirMemoryLogs, date+".md")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func (s *Store) ReadDailyLog(scope knowledge.Scope, ownerID, date string) ([]byte, error) {
	dir, err := s.scopeDir(scope, ownerID)
	if err != nil {
		return nil, err
	}
	return s.read(filepath.Join(dir, knowledge.DirMemoryLogs, date+".md"))
}

func (s *Store) ListDailyLogs(scope knowledge.Scope, ownerID string) ([]string, error) {
	dir, err := s.scopeDir(scope, ownerID)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(dir, knowledge.DirMemoryLogs))
	if os.IsNotExist(err) {
		return nil, knowledge.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func (s *Store) scopeDir(scope knowledge.Scope, ownerID string) (string, error) {
	switch scope {
	case knowledge.ScopeShared:
		return filepath.Join(s.Root, string(scope)), nil
	case knowledge.ScopeParticipant, knowledge.ScopePrincipal:
		if ownerID == "" || strings.Contains(ownerID, "..") || strings.ContainsAny(ownerID, `/\`) {
			return "", knowledge.ErrInvalidOwner
		}
		return filepath.Join(s.Root, string(scope), ownerID), nil
	default:
		return "", knowledge.ErrInvalidScope
	}
}

func (s *Store) read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, knowledge.ErrNotFound
	}
	return data, err
}

func copyIfExists(src, dst string) error {
	data, err := os.ReadFile(src)
	if os.IsNotExist(err) {
		return os.WriteFile(dst, []byte("# MEMORY\n"), 0o644)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
