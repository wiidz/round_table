package discord

import "strings"

// Discord counts message length in Unicode code points (roughly one CJK glyph = one).
const maxMessageRunes = 2000

// clipMessageRunes trims content to Discord's limit without splitting multibyte runes.
func clipMessageRunes(content string) string {
	content = strings.TrimSpace(content)
	runes := []rune(content)
	if len(runes) <= maxMessageRunes {
		return content
	}
	if maxMessageRunes <= 1 {
		return string(runes[:maxMessageRunes])
	}
	return string(runes[:maxMessageRunes-1]) + "…"
}

func splitDiscordMessages(content string, maxRunes int) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	if maxRunes <= 0 {
		maxRunes = maxMessageRunes
	}

	runes := []rune(content)
	if len(runes) <= maxRunes {
		return []string{content}
	}

	var parts []string
	remaining := runes
	for len(remaining) > maxRunes {
		cut := maxRunes
		window := string(remaining[:cut])
		if idx := strings.LastIndex(window, "\n\n"); idx >= 0 {
			if n := len([]rune(window[:idx])); n > maxRunes/3 {
				cut = n
			}
		} else if idx := strings.LastIndex(window, "\n"); idx >= 0 {
			if n := len([]rune(window[:idx])); n > maxRunes/3 {
				cut = n
			}
		}
		part := strings.TrimSpace(string(remaining[:cut]))
		if part != "" {
			parts = append(parts, part)
		}
		remaining = trimRunesLeft(remaining[cut:])
	}
	if tail := strings.TrimSpace(string(remaining)); tail != "" {
		parts = append(parts, tail)
	}
	return parts
}

func trimRunesLeft(runes []rune) []rune {
	for len(runes) > 0 {
		switch runes[0] {
		case ' ', '\t', '\n', '\r':
			runes = runes[1:]
		default:
			return runes
		}
	}
	return runes
}
