package discord

import (
	"fmt"
	"strconv"
	"strings"

	"round_table/apps/server/internal/platform/config"
)

type rosterLine struct {
	index   int
	id      string
	display string
}

func buildRosterLines(raw string) []rosterLine {
	meta := config.ParseMeetParticipants(raw)
	ids := config.MeetParticipantIDs(raw)
	out := make([]rosterLine, 0, len(ids))
	for i, id := range ids {
		display := id
		if entry, ok := meta[id]; ok && strings.TrimSpace(entry.DisplayName) != "" {
			display = strings.TrimSpace(entry.DisplayName)
		}
		out = append(out, rosterLine{index: i + 1, id: id, display: display})
	}
	return out
}

func summarizeParticipantIDs(raw string, ids []string) string {
	lines := buildRosterLines(raw)
	if len(ids) == 0 {
		return summarizeParticipants(raw)
	}
	allowed := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		allowed[id] = struct{}{}
	}
	var parts []string
	for _, line := range lines {
		if _, ok := allowed[line.id]; ok {
			parts = append(parts, line.id+"·"+line.display)
		}
	}
	if len(parts) == 0 {
		return summarizeParticipants(raw)
	}
	return strings.Join(parts, ", ")
}

func resolveRosterPick(input string, raw string) ([]string, error) {
	return pickParticipants(input, raw, nil, false)
}

func resolveParticipantPick(input string, raw string, casts []config.MeetCastConfig) ([]string, error) {
	if ids, ok := matchMeetCast(input, casts); ok {
		return ids, nil
	}
	return pickParticipants(input, raw, casts, true)
}

func pickParticipants(input string, raw string, casts []config.MeetCastConfig, allowCast bool) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errSetupReplyEmpty
	}
	lower := strings.ToLower(normalizeASCIIForms(input))
	if lower == "0" || lower == "all" || input == "全员" || input == "全部" {
		return nil, nil
	}

	lines := buildRosterLines(raw)
	if len(lines) == 0 {
		return nil, errNoParticipants
	}
	byIndex := make(map[int]string, len(lines))
	byID := make(map[string]string, len(lines))
	for _, line := range lines {
		byIndex[line.index] = line.id
		byID[strings.ToLower(line.id)] = line.id
	}

	tokens := splitParticipantPickTokens(input)
	if len(tokens) == 0 {
		return nil, errSetupInvalidParticipants
	}

	seen := make(map[string]struct{})
	var out []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if id, err := resolveRosterToken(token, lines, byIndex, byID); err != nil {
			return nil, err
		} else if id != "" {
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
			continue
		}
		if allowCast {
			if ids, ok := matchMeetCast(token, casts); ok {
				for _, id := range ids {
					if _, dup := seen[id]; dup {
						continue
					}
					seen[id] = struct{}{}
					out = append(out, id)
				}
				continue
			}
		}
		return nil, fmt.Errorf("%w: %q", errSetupInvalidParticipants, token)
	}
	if len(out) == 0 {
		return nil, errSetupInvalidParticipants
	}
	return out, nil
}

func resolveRosterToken(token string, lines []rosterLine, byIndex map[int]string, byID map[string]string) (string, error) {
	if n, err := strconv.Atoi(token); err == nil {
		id, ok := byIndex[n]
		if !ok {
			return "", fmt.Errorf("%w: index %d", errSetupInvalidParticipants, n)
		}
		return id, nil
	}
	if id, ok := byID[strings.ToLower(token)]; ok {
		return id, nil
	}
	matches := matchDisplayNames(token, lines)
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("%w: ambiguous %q", errSetupInvalidParticipants, token)
	}
	return "", nil
}

func matchMeetCast(token string, casts []config.MeetCastConfig) ([]string, bool) {
	token = strings.TrimSpace(token)
	if token == "" || len(casts) == 0 {
		return nil, false
	}
	norm := strings.ToLower(normalizeASCIIForms(token))
	castID := ""
	if strings.HasPrefix(norm, "c") && len(norm) > 1 {
		castID = strings.TrimPrefix(norm, "c")
	}
	for _, cast := range casts {
		idNorm := strings.ToLower(strings.TrimSpace(cast.ID))
		if castID != "" && castID == idNorm {
			return append([]string(nil), cast.ParticipantIDs...), true
		}
		if strings.EqualFold(token, cast.NameZH) || strings.EqualFold(token, cast.NameEN) {
			return append([]string(nil), cast.ParticipantIDs...), true
		}
	}
	return nil, false
}

func matchDisplayNames(token string, lines []rosterLine) []string {
	token = strings.ToLower(strings.TrimSpace(token))
	if token == "" {
		return nil
	}
	var exact []string
	var partial []string
	for _, line := range lines {
		display := strings.ToLower(line.display)
		id := strings.ToLower(line.id)
		if display == token || id == token {
			exact = append(exact, line.id)
			continue
		}
		if strings.Contains(display, token) || strings.Contains(id, token) {
			partial = append(partial, line.id)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	return partial
}

func splitParticipantPickTokens(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	repl := strings.NewReplacer("，", ",", "、", ",", ";", ",", "；", ",", "|", ",")
	input = repl.Replace(input)
	var tokens []string
	for _, part := range strings.Split(input, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, " ") {
			for _, sub := range strings.Fields(part) {
				if s := strings.TrimSpace(sub); s != "" {
					tokens = append(tokens, s)
				}
			}
			continue
		}
		tokens = append(tokens, part)
	}
	return tokens
}

func formatPickParticipantsPrompt(loc Locale, raw string, casts []config.MeetCastConfig) string {
	lines := buildRosterLines(raw)
	var b strings.Builder
	if loc == LocaleZH {
		b.WriteString("🎙️ **请选择参会专家**\n\n")
		if len(casts) > 0 {
			b.WriteString("**阵容**（可直接发编号或名称）：\n")
			for _, cast := range casts {
				b.WriteString(fmt.Sprintf("- **C%s** · %s — %s\n", cast.ID, cast.NameZH, strings.Join(cast.ParticipantIDs, ", ")))
			}
			b.WriteString("\n")
		}
		b.WriteString("**或按编号 / id / 显示名**（逗号分隔）：\n")
		for _, line := range lines {
			b.WriteString(fmt.Sprintf("%d · %s (`%s`)\n", line.index, line.display, line.id))
		}
		b.WriteString("\n**0** — 全员参会\n**取消会议** — 放弃")
		return b.String()
	}
	b.WriteString("🎙️ **Choose participants**\n\n")
	if len(casts) > 0 {
		b.WriteString("**Casts**:\n")
		for _, cast := range casts {
			b.WriteString(fmt.Sprintf("- **C%s** · %s — %s\n", cast.ID, cast.NameEN, strings.Join(cast.ParticipantIDs, ", ")))
		}
		b.WriteString("\n")
	}
	b.WriteString("**By index / id / display name** (comma-separated):\n")
	for _, line := range lines {
		b.WriteString(fmt.Sprintf("%d · %s (`%s`)\n", line.index, line.display, line.id))
	}
	b.WriteString("\n**0** — all roster participants\n**取消会议** — cancel")
	return b.String()
}
