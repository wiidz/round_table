package llm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	reStance       = regexp.MustCompile(`"stance"\s*:\s*"(agree|object|abstain|none)"`)
	reObjectReason = regexp.MustCompile(`"object_reason"\s*:\s*"((?:\\.|[^"\\])*)"`)
)

func parseOutput(raw string) (llmOutput, error) {
	raw = cleanRaw(raw)

	var out llmOutput
	if err := json.Unmarshal([]byte(raw), &out); err == nil && out.Content != "" {
		return out, nil
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(raw[start:end+1]), &out); err == nil && out.Content != "" {
			return out, nil
		}
	}

	if content, ok := extractContentLoose(raw); ok && content != "" {
		out.Content = content
		if m := reStance.FindStringSubmatch(raw); len(m) > 1 {
			out.Stance = m[1]
		}
		if m := reObjectReason.FindStringSubmatch(raw); len(m) > 1 {
			out.ObjectReason = unescapeJSONString(m[1])
		}
		return out, nil
	}
	return llmOutput{}, fmt.Errorf("invalid JSON: %q", truncateForError(raw, 240))
}

func cleanRaw(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

func extractContentLoose(raw string) (string, bool) {
	keyPos := strings.Index(raw, `"content"`)
	if keyPos < 0 {
		return "", false
	}
	after := strings.TrimSpace(raw[keyPos+len(`"content"`):])
	if !strings.HasPrefix(after, ":") {
		return "", false
	}
	after = strings.TrimSpace(after[1:])
	if len(after) == 0 || after[0] != '"' {
		return "", false
	}
	valueStart := keyPos + strings.Index(raw[keyPos:], `:"`) + len(`:"`)
	if raw[valueStart-1] != '"' {
		// tolerate `"content": "`
		valueStart = keyPos + strings.Index(raw[keyPos:], `: "`) + len(`: "`)
	}

	for _, delim := range []string{`","stance"`, `","object_reason"`, `,"stance"`, `,"object_reason"`} {
		if pos := strings.Index(raw[valueStart:], delim); pos >= 0 {
			return strings.TrimSpace(raw[valueStart : valueStart+pos]), true
		}
	}

	trimmed := strings.TrimSpace(raw)
	close := strings.LastIndex(trimmed, `"}`)
	if close > valueStart {
		return strings.TrimSpace(raw[valueStart:close]), true
	}
	return "", false
}

func unescapeJSONString(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			out.WriteByte(s[i])
			continue
		}
		switch s[i+1] {
		case '"', '\\', '/':
			out.WriteByte(s[i+1])
			i++
		case 'n':
			out.WriteByte('\n')
			i++
		case 't':
			out.WriteByte('\t')
			i++
		default:
			out.WriteByte(s[i])
		}
	}
	return out.String()
}

func truncateForError(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
