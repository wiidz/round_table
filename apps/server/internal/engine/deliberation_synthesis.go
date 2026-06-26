package engine

import (
	"fmt"
	"regexp"
	"strings"

	"round_table/apps/server/internal/domain/meeting"
)

var (
	openQuestionSectionRe = regexp.MustCompile(`(?i)^(\*{0,2}\s*待决问题\s*\*{0,2}\s*[:：]?|\d+[.)]\s*待决问题)`)
	openQuestionLineRe    = regexp.MustCompile(`[？?]|是否|能否|要不要|该不该`)
	bulletLineRe          = regexp.MustCompile(`^(\s*[-*•]|\s*\d+[.)])\s+`)
)

var openQuestionSkipPrefixes = []string{
	"待决问题", "待决意见", "待决：", "开放问题", "未决事项",
	"从工程角度", "从运营角度", "基于上一轮", "**待决",
}

var resolvedLineMarkers = []string{
	"最终倾向", "请按此锁定", "明确坚持", "正式确认", "资源系统确认",
	"初期不上", "初期禁止", "预留接口但初期", "综合结论",
	"我明确坚持", "锁定资源规格",
}

var decisionLineMarkers = []string{
	"最终倾向", "请按此锁定", "正式确认", "资源系统确认", "明确坚持",
	"初期不上", "初期禁止", "预留接口但初期禁止", "综合结论",
	"我明确坚持——", "锁定资源规格", "建议走保守",
}

func moderatorSynthesizeFinal(s meeting.State) (summary string, openQuestions []string) {
	openQuestions = collectDeliberationOpenQuestions(s)
	decisions := collectDeliberationDecisions(s)
	coreScheme := summarizeCoreScheme(s)

	var b strings.Builder
	b.WriteString("# 方案草案\n\n")
	writeDeliberationExecutiveSummary(&b, s, coreScheme, decisions, openQuestions)

	b.WriteString("\n\n## 详细记录\n\n")
	b.WriteString("### 主题\n\n")
	b.WriteString(s.Topic)
	if s.Goal != "" {
		b.WriteString("\n\n### 目标\n\n")
		b.WriteString(s.Goal)
	}

	if s.PreMeetingSummary != "" {
		b.WriteString("\n\n### 初始视角（Pre-meeting）\n\n")
		b.WriteString(strings.TrimSpace(s.PreMeetingSummary))
	}

	b.WriteString("\n\n### 各轮贡献汇总\n\n")
	for _, r := range s.Minutes.Rounds {
		if r.RoundNumber <= 0 {
			continue
		}
		fmt.Fprintf(&b, "#### Round %d\n\n%s\n\n", r.RoundNumber, strings.TrimSpace(r.Summary))
		if mod, ok := s.ModeratorSummaries[r.RoundNumber]; ok {
			fmt.Fprintf(&b, "##### Moderator 提炼\n\n%s\n\n", strings.TrimSpace(mod))
		}
	}

	if s.FreeDialogueCompleted && s.FreeDialogueSummary != "" {
		b.WriteString("### 自由对话要点\n\n")
		b.WriteString(strings.TrimSpace(s.FreeDialogueSummary))
		b.WriteString("\n\n")
	}

	return strings.TrimSpace(b.String()), openQuestions
}

func writeDeliberationExecutiveSummary(b *strings.Builder, s meeting.State, coreScheme string, decisions, openQuestions []string) {
	b.WriteString("## Executive Summary\n\n")
	fmt.Fprintf(b, "**主题**：%s\n\n", s.Topic)
	if s.Goal != "" {
		fmt.Fprintf(b, "**目标**：%s\n\n", s.Goal)
	}

	b.WriteString("### 核心方案（摘要）\n\n")
	if coreScheme != "" {
		b.WriteString(coreScheme)
		b.WriteByte('\n')
	} else {
		b.WriteString("- （见各轮详细记录）\n")
	}

	b.WriteString("\n### 已决要点\n\n")
	if len(decisions) == 0 {
		b.WriteString("- （讨论中未形成明确决议表述 — 见各轮发言）\n")
	} else {
		for _, d := range decisions {
			fmt.Fprintf(b, "- %s\n", d)
		}
	}

	b.WriteString("\n### 待决事项\n\n")
	if len(openQuestions) == 0 {
		b.WriteString("- （无显式开放问题 — Principal 可指定下一版需深化的模块）\n")
	} else {
		for _, q := range openQuestions {
			fmt.Fprintf(b, "- %s\n", q)
		}
	}

	b.WriteString("\n> 完整发言与 Q&A 见下方「详细记录」。\n")
}

func summarizeCoreScheme(s meeting.State) string {
	// Prefer latest moderator summary bullets; fallback to last round designer/key points.
	if mod, ok := s.ModeratorSummaries[s.CurrentRound]; ok {
		points := extractSchemePoints(mod)
		if len(points) > 0 {
			return formatBulletList(points, 6)
		}
	}
	for round := s.CurrentRound; round >= 1; round-- {
		for _, id := range s.RoundOrder {
			r, ok := s.RoundResponses[round][id]
			if !ok {
				continue
			}
			points := extractSchemePoints(r.Content)
			if len(points) >= 2 {
				return formatBulletList(points, 5)
			}
		}
	}
	return ""
}

func extractSchemePoints(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !bulletLineRe.MatchString(line) && !numberedLine.MatchString(line) {
			continue
		}
		line = numberedLine.ReplaceAllString(line, "")
		line = bulletLineRe.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "**")
		if idx := strings.Index(line, "**"); idx > 0 {
			line = strings.TrimSpace(line[idx+2:])
		}
		line = strings.TrimLeft(line, "：:")
		if len([]rune(line)) < 8 {
			continue
		}
		if isOpenQuestionLine(line) || isResolvedLine(line) {
			continue
		}
		out = append(out, truncateRunes(line, 160))
	}
	return out
}

func formatBulletList(items []string, max int) string {
	if len(items) > max {
		items = items[:max]
	}
	var b strings.Builder
	for _, item := range items {
		fmt.Fprintf(&b, "- %s\n", item)
	}
	return strings.TrimSpace(b.String())
}

func collectDeliberationOpenQuestions(s meeting.State) []string {
	var sources []string
	if s.FreeDialogueSummary != "" {
		sources = append(sources, s.FreeDialogueSummary)
	}
	for round := s.CurrentRound; round >= 1; round-- {
		for _, id := range s.RoundOrder {
			if r, ok := s.RoundResponses[round][id]; ok {
				sources = append(sources, r.Content)
			}
		}
	}

	seen := make(map[string]bool)
	var out []string
	for _, src := range sources {
		for _, q := range extractOpenQuestionsFromText(src) {
			key := normalizeQuestionKey(q)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, q)
		}
	}
	if len(out) > 12 {
		out = out[:12]
	}
	return out
}

func extractOpenQuestionsFromText(text string) []string {
	lines := strings.Split(text, "\n")
	inSection := false
	var out []string

	flushLine := func(line string) {
		line = cleanOpenQuestionLine(line)
		if line == "" || !isOpenQuestionLine(line) || isResolvedLine(line) || isOpenQuestionNoise(line) {
			return
		}
		out = append(out, truncateRunes(line, 220))
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			if inSection {
				inSection = false
			}
			continue
		}
		if openQuestionSectionRe.MatchString(line) {
			inSection = true
			if isOpenQuestionLine(line) && !isOpenQuestionNoise(line) {
				flushLine(line)
			}
			continue
		}
		if inSection {
			if bulletLineRe.MatchString(line) || numberedLine.MatchString(line) || isOpenQuestionLine(line) {
				flushLine(line)
				continue
			}
			if strings.HasPrefix(line, "###") || strings.HasPrefix(line, "##") {
				inSection = false
			}
		}
		if isOpenQuestionLine(line) && (strings.Contains(line, "？") || strings.Contains(line, "?")) {
			flushLine(line)
		}
	}
	return out
}

func collectDeliberationDecisions(s meeting.State) []string {
	var sources []string
	if s.FreeDialogueSummary != "" {
		sources = append(sources, s.FreeDialogueSummary)
	}
	for round := s.CurrentRound; round >= 1; round-- {
		for _, id := range s.RoundOrder {
			if r, ok := s.RoundResponses[round][id]; ok {
				sources = append(sources, r.Content)
			}
		}
	}

	seen := make(map[string]bool)
	var out []string
	for _, src := range sources {
		for _, line := range strings.Split(src, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || isOpenQuestionLine(line) {
				continue
			}
			if !isDecisionLine(line) {
				continue
			}
			line = cleanOpenQuestionLine(line)
			if line == "" {
				continue
			}
			key := normalizeQuestionKey(line)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, truncateRunes(line, 200))
		}
	}
	if len(out) > 10 {
		out = out[:10]
	}
	return out
}

func isOpenQuestionLine(line string) bool {
	if openQuestionLineRe.MatchString(line) {
		return true
	}
	for _, p := range openQuestionSkipPrefixes {
		if strings.HasPrefix(line, p) && strings.ContainsAny(line, "？?") {
			return true
		}
	}
	return false
}

func isResolvedLine(line string) bool {
	for _, m := range resolvedLineMarkers {
		if strings.Contains(line, m) {
			return true
		}
	}
	return false
}

func isDecisionLine(line string) bool {
	for _, m := range decisionLineMarkers {
		if strings.Contains(line, m) {
			return true
		}
	}
	return false
}

func isOpenQuestionNoise(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || trimmed == "待决问题：" || trimmed == "**待决问题：**" {
		return true
	}
	for _, p := range openQuestionSkipPrefixes {
		if trimmed == p || trimmed == p+"：" || trimmed == p+":" {
			return true
		}
	}
	// Section headers without substance
	if strings.HasSuffix(trimmed, "：") && !strings.ContainsAny(trimmed, "？?") && len([]rune(trimmed)) < 20 {
		return true
	}
	if strings.Contains(trimmed, "待决问题列表") {
		return true
	}
	return false
}

func cleanOpenQuestionLine(line string) string {
	line = strings.TrimSpace(line)
	line = numberedLine.ReplaceAllString(line, "")
	line = bulletLineRe.ReplaceAllString(line, "")
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "**")
	if idx := strings.Index(line, "**"); idx > 0 {
		line = strings.TrimSpace(line[idx+2:])
	}
	line = strings.TrimLeft(line, "：:")
	return strings.TrimSpace(line)
}

func normalizeQuestionKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "　", "")
	return s
}
