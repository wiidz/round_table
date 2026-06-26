package engine

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

var numberedLine = regexp.MustCompile(`(?m)^\s*\d+[.)]\s*`)

func summarizePreMeeting(s meeting.State) string {
	var b strings.Builder
	b.WriteString("Pre-meeting perspectives\n\n")
	for _, id := range s.RoundOrder {
		r := s.RoundResponses[0][id]
		role := s.Participants[id].Role
		fmt.Fprintf(&b, "- **%s** (%s): %s\n", id, role, r.Content)
	}
	return b.String()
}

func moderatorSummarizeRound(s meeting.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Round %d 摘要\n\n", s.CurrentRound)

	var objectors, supporters []participantTurn
	for _, id := range s.RoundOrder {
		r := s.RoundResponses[s.CurrentRound][id]
		turn := participantTurn{
			id:   id,
			role: s.Participants[id].Role,
			stance: string(r.Stance),
			reason: r.ObjectReason,
			points: extractKeyPoints(r.Content),
		}
		if r.Stance == event.StanceObject {
			objectors = append(objectors, turn)
		} else {
			supporters = append(supporters, turn)
		}
	}

	if len(objectors) > 0 {
		b.WriteString("### 未解决的分歧\n\n")
		for _, o := range objectors {
			fmt.Fprintf(&b, "- **%s** (%s) _object_\n", o.id, o.role)
			if o.reason != "" {
				fmt.Fprintf(&b, "  - 核心理由：%s\n", o.reason)
			}
			for _, p := range o.points {
				fmt.Fprintf(&b, "  - %s\n", p)
			}
		}
		b.WriteByte('\n')
	}

	if len(supporters) > 0 {
		b.WriteString("### 提出的方案与缓解措施\n\n")
		for _, p := range supporters {
			fmt.Fprintf(&b, "- **%s** (%s) _%s_\n", p.id, p.role, p.stance)
			if len(p.points) == 0 {
				b.WriteString("  - （未列出具体措施）\n")
				continue
			}
			for _, pt := range p.points {
				fmt.Fprintf(&b, "  - %s\n", pt)
			}
		}
		b.WriteByte('\n')
	}

	b.WriteString("### 共识状态\n\n")
	if len(objectors) > 0 {
		fmt.Fprintf(&b, "本轮 **未达成共识**（%d 位 object）。下一轮需针对上述分歧给出可验证的补救或设计承诺。\n", len(objectors))
	} else {
		b.WriteString("本轮 **无 object**，可进入共识判定。\n")
	}

	return strings.TrimSpace(b.String())
}

func moderatorSummarizeDeliberationRound(s meeting.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Round %d 研讨摘要\n\n", s.CurrentRound)

	b.WriteString("### 各角色贡献\n\n")
	for _, id := range s.RoundOrder {
		r := s.RoundResponses[s.CurrentRound][id]
		role := s.Participants[id].Role
		fmt.Fprintf(&b, "- **%s** (%s)\n", id, role)
		points := extractKeyPoints(r.Content)
		if len(points) == 0 {
			text := strings.TrimSpace(r.Content)
			if text == "" {
				b.WriteString("  - （无内容）\n")
				continue
			}
			fmt.Fprintf(&b, "  - %s\n", truncateRunes(text, 500))
			continue
		}
		for _, p := range points {
			fmt.Fprintf(&b, "  - %s\n", p)
		}
	}

	if s.CurrentRound > 1 {
		b.WriteString("\n### 与上轮衔接\n\n")
		fmt.Fprintf(&b, "承接 Round %d 讨论；完整发言见上文 Round %d。\n", s.CurrentRound-1, s.CurrentRound-1)
	}

	b.WriteString("\n### 研讨状态\n\n")
	if s.CurrentRound >= s.MaxRoundsPerSegment {
		b.WriteString("本轮为最后一轮，接下来将合成 **design-draft**。\n")
	} else {
		b.WriteString("继续下一轮以补全方案要素；最后一轮后将合成 **design-draft**。\n")
	}
	return strings.TrimSpace(b.String())
}

type participantTurn struct {
	id, role, stance, reason string
	points                     []string
}

func extractKeyPoints(content string) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	if numberedLine.MatchString(content) {
		var points []string
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || !numberedLine.MatchString(line) {
				continue
			}
			line = numberedLine.ReplaceAllString(line, "")
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// drop markdown bold headers like **JWT泄露面**：
			line = strings.TrimPrefix(line, "**")
			if idx := strings.Index(line, "**"); idx > 0 {
				line = strings.TrimSpace(line[idx+2:])
			}
			line = strings.TrimLeft(line, "：:")
			points = append(points, truncateRunes(line, 200))
		}
		if len(points) > 0 {
			return points
		}
	}
	// fallback: first sentence or truncated paragraph
	first := content
	if idx := strings.IndexAny(content, "。.\n"); idx > 0 {
		first = content[:idx+1]
	}
	return []string{truncateRunes(first, 200)}
}

func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "…"
}
