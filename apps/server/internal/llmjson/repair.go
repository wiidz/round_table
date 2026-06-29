package llmjson

import "strings"

// Clean strips markdown code fences from LLM output.
func Clean(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

// UnwrapObject returns the outermost {...} slice when present.
func UnwrapObject(raw string) string {
	raw = Clean(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return raw[start : end+1]
	}
	return raw
}

// RepairObject fixes common LLM JSON mistakes (missing brace, Chinese closing quotes).
func RepairObject(raw string) string {
	raw = UnwrapObject(raw)
	switch {
	case strings.HasSuffix(raw, "」}"):
		raw = raw[:len(raw)-len("」}")] + `"}` 
	case strings.HasSuffix(raw, "」",):
		raw = raw[:len(raw)-len("」",)] + `",`
	case strings.HasSuffix(raw, "」"):
		raw = raw[:len(raw)-len("」")] + `"`
	}
	if strings.HasPrefix(raw, "{") && !strings.HasSuffix(strings.TrimSpace(raw), "}") {
		raw += "}"
	}
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
