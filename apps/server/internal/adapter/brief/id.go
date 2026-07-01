package brief

import (
	"fmt"
	"strings"
	"unicode"
)

// SlugTemplateID derives a filesystem-safe template id from a human title.
func SlugTemplateID(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(title)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			if b.Len() > 0 && b.String()[b.Len()-1] != '_' {
				b.WriteByte('_')
			}
		case unicode.Is(unicode.Han, r):
			// keep Han runes as-is for readable ids in zh titles
			b.WriteRune(r)
		}
	}
	s := strings.Trim(strings.TrimSpace(b.String()), "_")
	if s == "" {
		return "brief-template"
	}
	return s
}

// NextAvailableTemplateID picks base or base-2, base-3, … while taken returns true.
func NextAvailableTemplateID(base string, taken func(id string) bool) string {
	if base == "" {
		base = "brief-template"
	}
	if !taken(base) {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !taken(candidate) {
			return candidate
		}
	}
}
