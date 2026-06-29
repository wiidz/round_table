package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/workspace"
)

// ListMeetings scans workspace root for meeting directories.
func (s *Store) ListMeetings() ([]workspace.MeetingIndex, error) {
	root := filepath.Clean(s.Root)
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return []workspace.MeetingIndex{}, nil
	}
	if err != nil {
		return nil, err
	}

	out := make([]workspace.MeetingIndex, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		if id == "" || strings.Contains(id, "..") {
			continue
		}
		idx := workspace.MeetingIndex{
			ID:        id,
			Topic:     "",
			Status:    "",
			UpdatedAt: dirUpdatedAt(filepath.Join(root, id)),
		}
		if data, err := s.Read(id, workspace.FileMeeting); err == nil {
			EnrichFromMeetingDoc(&idx, string(data))
		}
		if data, err := s.Read(id, workspace.FileUsageSummary); err == nil {
			EnrichFromUsageSummary(&idx, string(data))
		}
		out = append(out, idx)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

// ListMeetingsPage returns one page (1-based) after sorting by UpdatedAt desc.
func (s *Store) ListMeetingsPage(page, pageSize int) (workspace.PaginatedMeetings, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}
	all, err := s.ListMeetings()
	if err != nil {
		return workspace.PaginatedMeetings{}, err
	}
	total := len(all)
	start := (page - 1) * pageSize
	if start >= total {
		return workspace.PaginatedMeetings{
			Meetings: []workspace.MeetingIndex{},
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return workspace.PaginatedMeetings{
		Meetings: all[start:end],
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func dirUpdatedAt(dir string) time.Time {
	info, err := os.Stat(dir)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func parseMeetingTopic(doc string) string {
	const marker = "## 会议主题"
	i := strings.Index(doc, marker)
	if i < 0 {
		return ""
	}
	rest := strings.TrimSpace(doc[i+len(marker):])
	if rest == "" {
		return ""
	}
	if rest[0] == '\n' {
		rest = strings.TrimSpace(rest[1:])
	}
	if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
		rest = rest[:nl]
	}
	return strings.TrimSpace(rest)
}

func parseMeetingStatus(doc string) string {
	for _, line := range strings.Split(doc, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "| 会议状态 |") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			return ""
		}
		return strings.TrimSpace(parts[2])
	}
	return ""
}
