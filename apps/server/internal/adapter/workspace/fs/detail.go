package fs

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"round_table/apps/server/internal/adapter/workspace"
)

const (
	maxMeetingFileBytes = 512 << 10
	maxMeetingFiles     = 64
)

// ReadMeetingDetail loads markdown files for one meeting workspace.
func (s *Store) ReadMeetingDetail(id string) (workspace.MeetingDetail, error) {
	dir, err := s.meetingDir(id)
	if err != nil {
		return workspace.MeetingDetail{}, err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return workspace.MeetingDetail{}, workspace.ErrNotFound
	}

	files := make(map[string]string)
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		if len(files) >= maxMeetingFiles {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "" || strings.Contains(rel, "..") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() > maxMeetingFileBytes {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = string(data)
		return nil
	})
	if err != nil {
		return workspace.MeetingDetail{}, err
	}

	detail := workspace.MeetingDetail{
		MeetingIndex: workspace.MeetingIndex{
			ID:        id,
			UpdatedAt: dirUpdatedAt(dir),
		},
		Files: files,
	}
	if doc, ok := files[workspace.FileMeeting]; ok {
		EnrichFromMeetingDoc(&detail.MeetingIndex, doc)
	}
	if doc, ok := files[workspace.FileUsageSummary]; ok {
		EnrichFromUsageSummary(&detail.MeetingIndex, doc)
	}

	if len(files) == 0 {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return workspace.MeetingDetail{}, workspace.ErrNotFound
		}
	}

	sortMeetingFileKeys(files)
	return detail, nil
}

func sortMeetingFileKeys(files map[string]string) {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	ordered := make(map[string]string, len(files))
	for _, name := range names {
		ordered[name] = files[name]
	}
	for k := range files {
		delete(files, k)
	}
	for k, v := range ordered {
		files[k] = v
	}
}
