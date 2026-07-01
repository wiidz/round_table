package profile

import (
	"strings"
)

// UserProfile is the structured Principal USER.md payload (ADR-0010).
type UserProfile struct {
	Language     string `json:"language"`
	Confirmation string `json:"confirmation,omitempty"`
	Context      string `json:"context,omitempty"`
}

// DefaultUserProfile returns empty profile defaults.
func DefaultUserProfile() UserProfile {
	return UserProfile{Language: "zh-CN"}
}

// ParseUserMD extracts structured fields from USER.md.
func ParseUserMD(content string) UserProfile {
	out := DefaultUserProfile()
	content = strings.TrimSpace(content)
	if content == "" {
		return out
	}

	preferences := extractUserSection(content, "Preferences")
	for _, line := range strings.Split(preferences, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
		line = strings.TrimSpace(strings.TrimPrefix(line, "*"))
		if line == "" {
			continue
		}
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:colon]))
		value := strings.TrimSpace(line[colon+1:])
		switch key {
		case "language":
			if value != "" {
				out.Language = value
			}
		case "confirmation":
			out.Confirmation = value
		}
	}

	out.Context = extractUserSection(content, "Context")
	return out
}

// RenderUserMD composes USER.md from structured fields.
func RenderUserMD(p UserProfile) string {
	lang := strings.TrimSpace(p.Language)
	if lang == "" {
		lang = "zh-CN"
	}
	var b strings.Builder
	b.WriteString("# USER\n\n## Preferences\n\n")
	b.WriteString("- Language: ")
	b.WriteString(lang)
	b.WriteByte('\n')
	if c := strings.TrimSpace(p.Confirmation); c != "" {
		b.WriteString("- Confirmation: ")
		b.WriteString(c)
		b.WriteByte('\n')
	}
	b.WriteString("\n## Context\n\n")
	if ctx := strings.TrimSpace(p.Context); ctx != "" {
		b.WriteString(ctx)
		b.WriteByte('\n')
	}
	return b.String()
}

func extractUserSection(doc, heading string) string {
	marker := "## " + heading
	idx := strings.Index(doc, marker)
	if idx < 0 {
		return ""
	}
	rest := doc[idx+len(marker):]
	rest = strings.TrimLeft(rest, " \t\r\n")
	if next := strings.Index(rest, "\n## "); next >= 0 {
		rest = rest[:next]
	}
	return strings.TrimSpace(rest)
}
