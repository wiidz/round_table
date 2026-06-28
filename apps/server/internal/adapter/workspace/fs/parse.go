package fs

import (
	"strconv"
	"strings"
	"unicode"

	"round_table/apps/server/internal/adapter/workspace"
)

// EnrichFromMeetingDoc fills index fields from MEETING.md when present.
func EnrichFromMeetingDoc(idx *workspace.MeetingIndex, doc string) {
	if idx.Topic == "" {
		idx.Topic = parseMeetingTopic(doc)
	}
	if s := parseMeetingStatus(doc); s != "" {
		idx.Status = s
	}
	if mode := parseMeetingTableField(doc, "会议模式"); mode != "" {
		idx.Mode = mode
		idx.ModeKind = parseMeetingModeKind(mode)
	}
	if started := parseMeetingTableField(doc, "会议时间"); started != "" {
		idx.StartedAt = started
	}
	if n := parseMeetingMaxRounds(doc); n > 0 {
		idx.MaxRounds = n
	}
	idx.FreeDialogue = parseMeetingFreeDialogue(doc)
	if n := parseMeetingParticipantCount(doc); n > 0 {
		idx.ParticipantCount = n
	}
}

func (s *Store) EnrichMeetingIndex(idx workspace.MeetingIndex) workspace.MeetingIndex {
	if data, err := s.Read(idx.ID, workspace.FileMeeting); err == nil {
		EnrichFromMeetingDoc(&idx, string(data))
	}
	return idx
}

func parseMeetingTableField(doc, label string) string {
	prefix := "| " + label + " |"
	for _, line := range strings.Split(doc, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
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

func parseMeetingModeKind(mode string) string {
	lower := strings.ToLower(mode)
	switch {
	case strings.Contains(mode, "研讨") || strings.Contains(lower, "deliberation"):
		return "deliberation"
	case strings.Contains(mode, "裁决") || strings.Contains(lower, "decision"):
		return "decision"
	default:
		return ""
	}
}

func parseMeetingMaxRounds(doc string) int {
	raw := parseMeetingTableField(doc, "辩论轮次上限")
	if raw == "" {
		return 0
	}
	n := 0
	for _, r := range raw {
		if unicode.IsDigit(r) {
			n = n*10 + int(r-'0')
		} else if n > 0 {
			break
		}
	}
	return n
}

func parseMeetingFreeDialogue(doc string) bool {
	raw := parseMeetingTableField(doc, "Round 1 后自由对话")
	if raw == "" {
		return false
	}
	raw = strings.TrimSpace(raw)
	if raw == "0" || strings.HasPrefix(raw, "0 ") {
		return false
	}
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return !unicode.IsDigit(r)
	}) {
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err == nil && n > 0 {
			return true
		}
	}
	return strings.Contains(raw, "每人最多") || strings.Contains(raw, "轮提问")
}

func parseMeetingParticipantCount(doc string) int {
	const marker = "## 参会人员"
	i := strings.Index(doc, marker)
	if i < 0 {
		return 0
	}
	rest := doc[i+len(marker):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}

	count := 0
	passedSep := false
	for _, line := range strings.Split(rest, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		if strings.Contains(line, "---") {
			passedSep = true
			continue
		}
		if !passedSep {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		name := strings.TrimSpace(parts[1])
		if name == "" || name == "参会者" || strings.HasPrefix(name, "_") {
			continue
		}
		count++
	}
	return count
}
