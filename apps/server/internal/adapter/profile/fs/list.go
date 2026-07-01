package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/profile"
)

var allowedPrincipalFiles = map[string]bool{
	profile.FileUser:   true,
	profile.FileSoul:   true,
	profile.FileAgents: true,
	profile.FileTools:  true,
}

var allowedParticipantFiles = map[string]bool{
	profile.FileSoul:   true,
	profile.FileAgents: true,
	profile.FileTools:  true,
}

// ListPrincipals scans profiles/principals/ for Principal directories.
func (s *Store) ListPrincipals() ([]profile.PrincipalIndex, error) {
	root := filepath.Join(s.Root, "principals")
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return []profile.PrincipalIndex{}, nil
	}
	if err != nil {
		return nil, err
	}

	out := make([]profile.PrincipalIndex, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
			continue
		}
		dir := filepath.Join(root, id)
		files, err := listMarkdownFiles(dir, allowedPrincipalFiles)
		if err != nil {
			return nil, err
		}
		out = append(out, profile.PrincipalIndex{
			ID:        id,
			Files:     files,
			UpdatedAt: dirUpdatedAt(dir),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// ReadPrincipalDetail loads allowed markdown files for a Principal.
func (s *Store) ReadPrincipalDetail(id string) (profile.PrincipalDetail, error) {
	if err := validatePrincipalID(id); err != nil {
		return profile.PrincipalDetail{}, err
	}
	dir := filepath.Join(s.Root, "principals", id)
	files, err := listMarkdownFiles(dir, allowedPrincipalFiles)
	if err != nil {
		return profile.PrincipalDetail{}, err
	}
	if len(files) == 0 {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return profile.PrincipalDetail{}, profile.ErrNotFound
		}
	}

	content := make(map[string]string, len(files))
	for _, name := range files {
		data, err := s.ReadPrincipal(id, name)
		if err != nil {
			return profile.PrincipalDetail{}, err
		}
		content[name] = string(data)
	}
	return profile.PrincipalDetail{
		ID:    id,
		Files: content,
	}, nil
}

// WritePrincipalFile validates id/filename then writes content.
func (s *Store) WritePrincipalFile(id, filename string, data []byte) error {
	if err := validatePrincipalID(id); err != nil {
		return err
	}
	if err := validatePrincipalFilename(filename); err != nil {
		return err
	}
	if err := s.EnsurePrincipal(id); err != nil {
		return err
	}
	return s.WritePrincipal(id, filename, data)
}

func listMarkdownFiles(dir string, allowed map[string]bool) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, profile.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		if !allowed[name] && !isSafeProfileFilename(name) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func isSafeProfileFilename(name string) bool {
	if name == "" || strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return false
	}
	return strings.HasSuffix(strings.ToLower(name), ".md")
}

func validatePrincipalID(id string) error {
	return validateProfileID(id)
}

func validatePrincipalFilename(name string) error {
	if !isSafeProfileFilename(name) {
		return profile.ErrNotFound
	}
	if !allowedPrincipalFiles[name] {
		return profile.ErrNotFound
	}
	return nil
}

func dirUpdatedAt(dir string) time.Time {
	info, err := os.Stat(dir)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
