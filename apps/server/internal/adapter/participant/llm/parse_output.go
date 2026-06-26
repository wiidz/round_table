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
	reContentEnd   = regexp.MustCompile(`"\s*,\s*"(?:stance|object_reason)"`)
)

func parseOutput(raw string) (llmOutput, error) {
	raw = cleanRaw(raw)
	raw = repairMalformedJSON(raw)

	var out llmOutput
	if err := json.Unmarshal([]byte(raw), &out); err == nil && out.Content != "" {
		out.Content = finalizeContent(out.Content)
		return out, nil
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(raw[start:end+1]), &out); err == nil && out.Content != "" {
			out.Content = finalizeContent(out.Content)
			return out, nil
		}
	}

	if content, ok := extractContentLoose(raw); ok && content != "" {
		out.Content = finalizeContent(content)
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

// repairMalformedJSON fixes common LLM JSON mistakes (Chinese closing quotes, missing ").
func repairMalformedJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	switch {
	case strings.HasSuffix(raw, "」}"):
		raw = raw[:len(raw)-len("」}")] + `"}` 
	case strings.HasSuffix(raw, "」",):
		raw = raw[:len(raw)-len("」",)] + `",`
	case strings.HasSuffix(raw, "」"):
		raw = raw[:len(raw)-len("」")] + `"`
	}
	// content-only object missing closing brace (e.g. free-dialogue answer ending with 特色。」)
	if strings.HasPrefix(raw, "{") && strings.Contains(raw, `"content"`) && !strings.HasSuffix(raw, "}") {
		raw += "}"
	}
	// content-only object: {"content":"..."} with missing closing quote before }
	if strings.HasSuffix(raw, "}") && strings.Contains(raw, `"content"`) {
		if !strings.HasSuffix(raw, `"}`) && !strings.Contains(raw, `","`) {
			if idx := strings.LastIndex(raw, "}"); idx > 0 {
				before := strings.TrimRight(raw[:idx], " \t\n\r")
				if before != "" && !strings.HasSuffix(before, `"`) {
					raw = before + `"}` 
				}
			}
		}
	}
	return raw
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

	if loc := reContentEnd.FindStringIndex(raw[valueStart:]); loc != nil {
		return unescapeJSONString(strings.TrimSpace(raw[valueStart : valueStart+loc[0]])), true
	}

	for _, delim := range []string{
		`","stance"`, `","object_reason"`,
		`", "stance"`, `", "object_reason"`,
		`,"stance"`, `,"object_reason"`,
	} {
		if pos := strings.Index(raw[valueStart:], delim); pos >= 0 {
			return unescapeJSONString(strings.TrimSpace(raw[valueStart : valueStart+pos])), true
		}
	}

	trimmed := strings.TrimSpace(raw)
	close := strings.LastIndex(trimmed, `"}`)
	if close > valueStart && !strings.Contains(raw[valueStart:close], `", "stance"`) && !strings.Contains(raw[valueStart:close], `","stance"`) {
		return unescapeJSONString(strings.TrimSpace(raw[valueStart:close])), true
	}
	// Trailing 」 or 」} (after repair pass may still miss edge cases)
	if strings.HasSuffix(trimmed, "}") || strings.HasSuffix(trimmed, "」") {
		end := len(trimmed)
		if strings.HasSuffix(trimmed, "}") {
			end = strings.LastIndex(trimmed, "}")
		}
		if end > valueStart {
			content := strings.TrimSpace(raw[valueStart:end])
			content = strings.TrimSuffix(content, "」")
			content = strings.TrimSuffix(content, `"`)
			if content != "" {
				return unescapeJSONString(content), true
			}
		}
	}
	return "", false
}

// finalizeContent strips JSON syntax leaked into parsed content values.
func finalizeContent(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "」")
	for {
		prev := s
		s = strings.TrimSpace(s)
		for _, suffix := range []string{`", "stance"`, `","stance"`, `", "object_reason"`, `","object_reason"`, `",`} {
			if strings.HasSuffix(s, suffix) {
				s = strings.TrimSuffix(s, suffix)
				break
			}
		}
		s = strings.TrimSuffix(strings.TrimSpace(s), `"`)
		if s == prev {
			break
		}
	}
	return strings.TrimSpace(s)
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
