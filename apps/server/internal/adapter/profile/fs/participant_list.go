package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"round_table/apps/server/internal/adapter/profile"
)

// ListParticipants scans profiles/participants/ for Participant directories.
func (s *Store) ListParticipants() ([]profile.ParticipantIndex, error) {
	root := filepath.Join(s.Root, "participants")
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return []profile.ParticipantIndex{}, nil
	}
	if err != nil {
		return nil, err
	}

	out := make([]profile.ParticipantIndex, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		if err := validateProfileID(id); err != nil {
			continue
		}
		dir := filepath.Join(root, id)
		files, err := listMarkdownFiles(dir, allowedParticipantFiles)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			if err := s.EnsureParticipant(id); err != nil {
				return nil, err
			}
			files, err = listMarkdownFiles(dir, allowedParticipantFiles)
			if err != nil {
				return nil, err
			}
		}
		out = append(out, profile.ParticipantIndex{
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

// ReadParticipantDetail loads markdown profile files for a Participant.
func (s *Store) ReadParticipantDetail(id string) (profile.ParticipantDetail, error) {
	if err := validateProfileID(id); err != nil {
		return profile.ParticipantDetail{}, err
	}
	if err := s.EnsureParticipant(id); err != nil {
		return profile.ParticipantDetail{}, err
	}
	dir := filepath.Join(s.Root, "participants", id)
	files, err := listMarkdownFiles(dir, allowedParticipantFiles)
	if err != nil {
		return profile.ParticipantDetail{}, err
	}
	if len(files) == 0 {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return profile.ParticipantDetail{}, profile.ErrNotFound
		}
	}

	content := make(map[string]string, len(files))
	for _, name := range files {
		data, err := s.ReadParticipant(id, name)
		if err != nil {
			return profile.ParticipantDetail{}, err
		}
		content[name] = string(data)
	}
	return profile.ParticipantDetail{
		ID:    id,
		Files: content,
	}, nil
}

// WriteParticipantFile validates id/filename then writes content.
func (s *Store) WriteParticipantFile(id, filename string, data []byte) error {
	if err := validateProfileID(id); err != nil {
		return err
	}
	if err := validateParticipantFilename(filename); err != nil {
		return err
	}
	return s.WriteParticipant(id, filename, data)
}

func validateParticipantFilename(name string) error {
	if !isSafeProfileFilename(name) {
		return profile.ErrNotFound
	}
	if !allowedParticipantFiles[name] {
		return profile.ErrNotFound
	}
	return nil
}

func validateProfileID(id string) error {
	if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return profile.ErrNotFound
	}
	return nil
}
