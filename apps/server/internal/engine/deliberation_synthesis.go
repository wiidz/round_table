package engine

import (
	"fmt"
	"regexp"
	"strings"

	"round_table/apps/server/internal/domain/meeting"
)

var (
	openQuestionSectionRe = regexp.MustCompile(`(?i)^(\*{0,2}\s*(待决问题|开放问题|未决事项)\s*\*{0,2}\s*[:：]?|\d+[.)]\s*(待决问题|开放问题|问题\d*))`)
	openQuestionLineRe    = regexp.MustCompile(`[？?]|是否|能否|要不要|该不该|如何设计`)
	bulletLineRe          = regexp.MustCompile(`^(\s*[-*•]|\s*\d+[.)])\s+`)
	openQuestionPrefixRe  = regexp.MustCompile(`^([①②③④⑤⑥⑦⑧⑨⑩]|Q\d*|Q|问题\d*)\s*[:：]?\s*`)
	participantMetaRe     = regexp.MustCompile(`^[\(*（]*[^)）]*[\)）]\s*[:：]\s*`)
	perspectiveOpenerRe   = regexp.MustCompile(`^从.{1,12}角度`)
	sectionLabelRe        = regexp.MustCompile(`^(\d+[.)]\s*)?(\*{1,2})?[^*]{1,40}(\*{1,2})?\s*[:：]\s*`)
	weakDecisionSuffixRe  = regexp.MustCompile(`(：|\)|）)\s*支持[。.]?$`)
	freeDialogueAnswerRe  = regexp.MustCompile(`^A\s*[:：]\s*`)
)

// Generic deliberation phrasing — not tied to any domain or participant role.
var decisionLineMarkers = []string{
	"最终倾向", "最终结论", "正式确认", "综合结论", "明确结论",
	"确定采用", "决定采用", "达成共识", "一致同意", "一致支持",
	"强烈支持", "推荐采用", "统一为", "统一限制为", "明确为",
	"请按此锁定", "我明确坚持", "已采纳", "最终方案",
	"必须与", "我认同", "技术上可行", "可复用现有", "不再采用", "不采用",
}

var revisionAnchorMarkers = []string{
	"收束", "修订", "调整如下", "细化方案", "更新方案", "更新框架",
	"基于上一轮反馈", "基于上一轮", "收束核心", "收束框架",
}

var decisionSplitMarkers = []string{
	"，但需确认", "，需确认", "；需确认", "，尚需确认", "，待确认",
}

var introNoiseMarkers = []string{
	"结合前几轮", "基于上一轮", "基于讨论", "我建议进一步",
	"待决问题", "开放问题", "未决事项",
}

var deliberationStopTokens = map[string]bool{
	"如果": true, "是否": true, "能否": true, "或者": true, "以及": true,
	"需要": true, "建议": true, "可以": true, "进行": true, "一个": true,
}

func moderatorSynthesizeFinal(s meeting.State) (summary string, openQuestions []string, agenda *synthesisAgendaOutput) {
	decisions, spillover := collectDeliberationDecisions(s)
	openQuestions = collectDeliberationOpenQuestions(s, decisions, spillover)
	coreScheme := summarizeCoreScheme(s)
	coreBullets := schemeBulletsFromText(coreScheme)
	decisions = dedupeDecisionsAgainstCoreScheme(coreBullets, decisions)
	if hasAgendaForSynthesis(s) {
		out := ruleBasedAgendaSynthesis(s, coreBullets, decisions, openQuestions)
		verdict, keyDecisions := deriveExecutiveVerdict(s.Topic, coreScheme, decisions, openQuestions)
		out.ExecutiveVerdict = verdict
		out.KeyDecisions = keyDecisions
		summary, openQuestions = assembleDesignDraftFromAgenda(s, out)
		return summary, openQuestions, &out
	}
	verdict, keyDecisions := deriveExecutiveVerdict(s.Topic, coreScheme, decisions, openQuestions)
	return assembleDesignDraft(s, verdict, keyDecisions, coreScheme, decisions, openQuestions), openQuestions, nil
}

func assembleDesignDraft(s meeting.State, executiveVerdict string, keyDecisions []string, coreScheme string, decisions, openQuestions []string) string {
	var b strings.Builder
	b.WriteString("# 方案草案\n\n")
	writeDeliberationExecutiveSummary(&b, s, executiveVerdict, keyDecisions, coreScheme, decisions, openQuestions)

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

	return strings.TrimSpace(b.String())
}

func writeDeliberationExecutiveSummary(b *strings.Builder, s meeting.State, executiveVerdict string, keyDecisions []string, coreScheme string, decisions, openQuestions []string) {
	b.WriteString("## Executive Summary\n\n")
	fmt.Fprintf(b, "**主题**：%s\n\n", s.Topic)
	if s.Goal != "" {
		fmt.Fprintf(b, "**目标**：%s\n\n", s.Goal)
	}

	writeExecutiveVerdictBlock(b, executiveVerdict, keyDecisions)

	b.WriteString("### 核心方案（摘要）\n\n")
	if coreScheme != "" {
		b.WriteString(coreScheme)
		b.WriteByte('\n')
	} else {
		b.WriteString("- （见各轮详细记录）\n")
	}

	b.WriteString("\n### 已决要点\n\n")
	if len(decisions) == 0 {
		if strings.TrimSpace(coreScheme) != "" {
			b.WriteString("- （与核心方案一致，无额外增量决议 — 见各轮发言与自由对话）\n")
		} else {
			b.WriteString("- （讨论中未形成明确决议表述 — 见各轮发言）\n")
		}
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

func writeExecutiveVerdictBlock(b *strings.Builder, executiveVerdict string, keyDecisions []string) {
	verdict := strings.TrimSpace(executiveVerdict)
	if verdict == "" && len(keyDecisions) == 0 {
		return
	}
	b.WriteString("### 总括结论\n\n")
	if verdict != "" {
		b.WriteString(verdict)
		b.WriteString("\n\n")
	}
	if len(keyDecisions) > 0 {
		b.WriteString("**Principal 需知**\n\n")
		for _, item := range keyDecisions {
			fmt.Fprintf(b, "- %s\n", item)
		}
		b.WriteString("\n")
	}
}

func deriveExecutiveVerdict(topic, coreScheme string, decisions, openQuestions []string) (verdict string, keyDecisions []string) {
	keyDecisions = normalizeSynthesisStrings(decisions, 3)
	if len(keyDecisions) == 0 {
		keyDecisions = normalizeSynthesisStrings(schemeBulletsFromText(coreScheme), 3)
	}

	var parts []string
	if topic != "" {
		parts = append(parts, fmt.Sprintf("围绕「%s」", topic))
	}
	if len(keyDecisions) > 0 {
		parts = append(parts, fmt.Sprintf("建议采纳：%s", keyDecisions[0]))
	} else if coreBullets := schemeBulletsFromText(coreScheme); len(coreBullets) > 0 {
		parts = append(parts, fmt.Sprintf("建议方向：%s", coreBullets[0]))
	}
	if len(openQuestions) > 0 {
		parts = append(parts, fmt.Sprintf("仍有 %d 项待 Principal 或后续会议拍板", len(openQuestions)))
	} else if len(parts) > 0 {
		parts = append(parts, "主要分歧已收束")
	}
	if len(parts) == 0 {
		return "", keyDecisions
	}
	return truncateRunes(strings.Join(parts, "；")+"。", 600), keyDecisions
}

func schemeBulletsFromText(scheme string) []string {
	var items []string
	for _, line := range strings.Split(scheme, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
	}
	return items
}

func dedupeDecisionsAgainstCoreScheme(coreItems, decisions []string) []string {
	if len(coreItems) == 0 || len(decisions) == 0 {
		return decisions
	}
	return filterDecisionsAgainstCore(coreItems, decisions, strictCoreOverlapMin)
}

const strictCoreOverlapMin = 4

func filterDecisionsAgainstCore(coreItems, decisions []string, minOverlap int) []string {
	var out []string
	for _, d := range decisions {
		if overlapsCoreScheme(d, coreItems, minOverlap) {
			continue
		}
		out = append(out, d)
	}
	return out
}

func overlapsCoreScheme(decision string, coreItems []string, minOverlap int) bool {
	dKey := normalizeQuestionKey(decision)
	if dKey == "" {
		return false
	}
	for _, c := range coreItems {
		if coreSchemeNearDuplicate(decision, c) {
			return true
		}
		if deliberationTokenOverlap(decision, c) >= minOverlap {
			return true
		}
	}
	return false
}

func coreSchemeNearDuplicate(decision, core string) bool {
	dk := normalizeQuestionKey(decision)
	ck := normalizeQuestionKey(core)
	if dk == "" || ck == "" {
		return false
	}
	if dk == ck {
		return true
	}
	shorter, longer := dk, ck
	if len(shorter) > len(longer) {
		shorter, longer = longer, shorter
	}
	if len(shorter) >= 24 && strings.Contains(longer, shorter) {
		return true
	}
	return false
}

func summarizeCoreScheme(s meeting.State) string {
	var bestPoints []string
	bestScore := -1

	for round := s.CurrentRound; round >= 1; round-- {
		for idx, id := range s.RoundOrder {
			r, ok := s.RoundResponses[round][id]
			if !ok {
				continue
			}
			points := extractSchemePoints(r.Content)
			if len(points) < 2 {
				continue
			}
			score := schemeSourceScore(round, s.CurrentRound, idx, len(s.RoundOrder), r.Content)
			if score > bestScore {
				bestScore = score
				bestPoints = points
			}
		}
	}
	if len(bestPoints) >= 2 {
		return formatBulletList(bestPoints, 6)
	}

	for round := s.CurrentRound; round >= 1; round-- {
		if mod, ok := s.ModeratorSummaries[round]; ok {
			points := extractSchemePoints(mod)
			if len(points) >= 2 {
				return formatBulletList(points, 6)
			}
		}
	}
	for round := s.CurrentRound; round >= 1; round-- {
		for _, id := range s.RoundOrder {
			r, ok := s.RoundResponses[round][id]
			if !ok {
				continue
			}
			points := extractSchemePoints(r.Content)
			if len(points) > 0 {
				return formatBulletList(points, 5)
			}
		}
	}
	return ""
}

func schemeSourceScore(round, currentRound, speakerIdx, speakerCount int, content string) int {
	score := (speakerCount - speakerIdx) * 5
	if contentHasRevisionAnchor(content) {
		return 1000 + round*100 + score
	}
	// Without explicit revision language, prefer earlier substantive proposals.
	return (currentRound - round + 1) * 50 + score
}

func contentHasRevisionAnchor(text string) bool {
	for _, m := range revisionAnchorMarkers {
		if strings.Contains(text, m) {
			return true
		}
	}
	return false
}

func extractSchemePoints(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "——") || strings.HasPrefix(line, "--") {
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
		if isOpenQuestionLine(line) || isDecisionLine(line) || isIntroNoiseLine(line) {
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

func collectDeliberationOpenQuestions(s meeting.State, decisions, spillover []string) []string {
	sources := deliberationTextSources(s)

	seen := make(map[string]bool)
	var out []string
	addQuestion := func(q string) {
		q = strings.TrimSpace(q)
		if q == "" {
			return
		}
		key := normalizeQuestionKey(q)
		if key == "" || seen[key] || overlapsDecision(q, decisions) || overlapsOpenQuestion(q, out) {
			return
		}
		if isTentativeStatementNotQuestion(q) || isSuggestionNotQuestion(q) {
			return
		}
		seen[key] = true
		out = append(out, q)
	}

	for _, q := range spillover {
		addQuestion(q)
	}
	for _, src := range sources {
		for _, q := range extractOpenQuestionsFromText(src) {
			addQuestion(q)
		}
	}
	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

func deliberationTextSources(s meeting.State) []string {
	var sources []string
	if s.FreeDialogueSummary != "" {
		sources = append(sources, s.FreeDialogueSummary)
	}
	for round := s.CurrentRound; round >= 1; round-- {
		if mod, ok := s.ModeratorSummaries[round]; ok {
			sources = append(sources, mod)
		}
		for _, id := range s.RoundOrder {
			if r, ok := s.RoundResponses[round][id]; ok {
				sources = append(sources, r.Content)
			}
		}
	}
	return sources
}

func extractOpenQuestionsFromText(text string) []string {
	lines := strings.Split(text, "\n")
	inSection := false
	var out []string

	flushLine := func(line string) {
		line = cleanOpenQuestionLine(line)
		line = openQuestionPrefixRe.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line == "" || !isOpenQuestionLine(line) || isDecisionLine(line) || isOpenQuestionNoise(line) {
			return
		}
		if isOpenQuestionParagraphNoise(line) {
			return
		}
		out = append(out, truncateRunes(line, 160))
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

func collectDeliberationDecisions(s meeting.State) (decisions []string, spilloverQuestions []string) {
	sources := deliberationTextSources(s)

	seen := make(map[string]bool)
	addDecision := func(line string) {
		decision, trailing := finalizeDecisionLine(line)
		if decision == "" {
			return
		}
		key := normalizeQuestionKey(decision)
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		decisions = append(decisions, decision)
		if trailing != "" {
			spilloverQuestions = append(spilloverQuestions, trailing)
		}
	}

	for _, src := range sources {
		revisionAnchor := contentHasRevisionAnchor(src)
		for _, line := range strings.Split(src, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			line = freeDialogueAnswerRe.ReplaceAllString(line, "")
			if isDecisionLine(line) {
				addDecision(line)
			} else if revisionAnchor && isRevisionDecisionLine(line) {
				addDecision(line)
			}
		}
	}
	if len(decisions) > 10 {
		decisions = decisions[:10]
	}
	return decisions, spilloverQuestions
}

func finalizeDecisionLine(line string) (decision string, trailingQuestion string) {
	line = extractDecisionText(line)
	if line == "" {
		return "", ""
	}
	for _, marker := range decisionSplitMarkers {
		if idx := strings.Index(line, marker); idx > 0 {
			decision = strings.TrimSpace(line[:idx])
			trailingQuestion = strings.TrimSpace(line[idx+len(marker):])
			trailingQuestion = strings.TrimLeft(trailingQuestion, "：:")
			if trailingQuestion != "" && !strings.HasSuffix(trailingQuestion, "？") && !strings.HasSuffix(trailingQuestion, "?") {
				trailingQuestion += "？"
			}
			return truncateRunes(decision, 200), truncateRunes(trailingQuestion, 160)
		}
	}
	if qIdx := strings.IndexAny(line, "？?"); qIdx > 0 {
		head := strings.TrimSpace(strings.TrimRight(line[:qIdx], "，,；;"))
		if isDecisionLine(head) || strings.Contains(head, "已采纳") || strings.Contains(head, "暂定") || strings.Contains(head, "必须与") {
			return truncateRunes(head, 200), truncateRunes(strings.TrimSpace(line[qIdx:]), 160)
		}
	}
	return truncateRunes(line, 200), ""
}

func extractDecisionText(line string) string {
	line = cleanOpenQuestionLine(line)
	line = openQuestionPrefixRe.ReplaceAllString(line, "")
	line = participantMetaRe.ReplaceAllString(line, "")
	line = strings.TrimSpace(line)
	if line == "" || (isOpenQuestionLine(line) && !isDecisionLine(line)) {
		return ""
	}
	// Strip short section labels before the actual decision, e.g. "2. **议题 B**：强烈支持方案 A".
	if m := sectionLabelRe.FindStringIndex(line); m != nil && m[1] < 40 {
		rest := strings.TrimSpace(line[m[1]:])
		if rest != "" && isDecisionLine(rest) {
			line = rest
		}
	}
	return line
}

func isOpenQuestionLine(line string) bool {
	if openQuestionLineRe.MatchString(line) {
		return true
	}
	for _, p := range introNoiseMarkers {
		if strings.HasPrefix(line, p) && strings.ContainsAny(line, "？?") {
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
	if strings.Contains(line, "明确为") && !strings.ContainsAny(line, "？?") {
		return true
	}
	if strings.ContainsAny(line, "？?") {
		return false
	}
	if strings.Contains(line, "已采纳") || strings.Contains(line, "最终方案") {
		return true
	}
	if strings.Contains(line, "暂定") && strings.Contains(line, "支持") {
		return true
	}
	if strings.Contains(line, "采纳") {
		return true
	}
	if weakDecisionSuffixRe.MatchString(strings.TrimSpace(line)) {
		return true
	}
	return false
}

func isIntroNoiseLine(line string) bool {
	for _, p := range introNoiseMarkers {
		if strings.Contains(line, p) {
			return true
		}
	}
	if participantMetaRe.MatchString(line) {
		return true
	}
	if perspectiveOpenerRe.MatchString(line) {
		return true
	}
	if strings.HasSuffix(line, "：") && !strings.Contains(line, "明确为") && len([]rune(line)) < 24 {
		return true
	}
	return false
}

func isOpenQuestionNoise(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	for _, label := range []string{"待决问题", "开放问题", "未决事项"} {
		if trimmed == label+"：" || trimmed == "**"+label+"：**" {
			return true
		}
	}
	if perspectiveOpenerRe.MatchString(trimmed) {
		return true
	}
	for _, p := range introNoiseMarkers {
		if trimmed == p || trimmed == p+"：" || trimmed == p+":" {
			return true
		}
	}
	if strings.HasSuffix(trimmed, "：") && !strings.ContainsAny(trimmed, "？?") && len([]rune(trimmed)) < 20 {
		return true
	}
	if strings.Contains(trimmed, "待决问题列表") || strings.Contains(trimmed, "开放问题列表") {
		return true
	}
	return false
}

func isOpenQuestionParagraphNoise(line string) bool {
	runes := len([]rune(line))
	if runes <= 120 {
		return false
	}
	if strings.Count(line, "。") >= 2 || strings.Count(line, "；") >= 2 {
		return true
	}
	trimmed := strings.TrimSpace(line)
	if runes > 160 && !strings.HasSuffix(trimmed, "？") && !strings.HasSuffix(trimmed, "?") {
		return true
	}
	return false
}

func overlapsDecision(question string, decisions []string) bool {
	qKey := normalizeQuestionKey(question)
	if qKey == "" {
		return false
	}
	for _, d := range decisions {
		dk := normalizeQuestionKey(d)
		if dk == "" {
			continue
		}
		if strings.Contains(qKey, dk) || strings.Contains(dk, qKey) {
			return true
		}
		if deliberationTokenOverlap(question, d) >= 2 {
			return true
		}
	}
	return false
}

func deliberationTokenOverlap(a, b string) int {
	ka := chineseKeywords(a)
	count := 0
	for k := range chineseKeywords(b) {
		if ka[k] {
			count++
		}
	}
	return count
}

func chineseKeywords(s string) map[string]bool {
	runes := []rune(s)
	seen := make(map[string]bool)
	for i := 0; i < len(runes); i++ {
		if !isHan(runes[i]) {
			continue
		}
		for n := 2; n <= 3 && i+n <= len(runes); n++ {
			chunk := runes[i : i+n]
			if !allHan(chunk) {
				break
			}
			kw := string(chunk)
			if deliberationStopTokens[kw] {
				continue
			}
			seen[kw] = true
		}
	}
	return seen
}

func isHan(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF
}

func allHan(runes []rune) bool {
	for _, r := range runes {
		if !isHan(r) {
			return false
		}
	}
	return true
}

func isRevisionDecisionLine(line string) bool {
	if strings.ContainsAny(line, "？?") {
		return false
	}
	if !bulletLineRe.MatchString(line) && !numberedLine.MatchString(line) {
		return false
	}
	cleaned := cleanOpenQuestionLine(line)
	cleaned = openQuestionPrefixRe.ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)
	if len([]rune(cleaned)) < 12 {
		return false
	}
	if isOpenQuestionLine(cleaned) || isIntroNoiseLine(cleaned) {
		return false
	}
	for _, p := range []string{"待决问题", "开放问题", "补充以下", "补充几点", "针对待决"} {
		if strings.Contains(cleaned, p) {
			return false
		}
	}
	return true
}

func overlapsOpenQuestion(candidate string, existing []string) bool {
	for _, e := range existing {
		if normalizeQuestionKey(candidate) == normalizeQuestionKey(e) {
			return true
		}
		if deliberationTokenOverlap(candidate, e) >= 3 {
			return true
		}
	}
	return false
}

func isSuggestionNotQuestion(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.ContainsAny(trimmed, "？?") {
		return false
	}
	if strings.Contains(trimmed, "希望") && strings.Contains(trimmed, "评估") {
		return true
	}
	if strings.HasPrefix(trimmed, "建议") && !strings.Contains(trimmed, "是否") {
		return true
	}
	return false
}

func isTentativeStatementNotQuestion(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.Contains(trimmed, "倾向") {
		return false
	}
	// "我倾向 X，但需评估" — not a focused open question.
	if !strings.ContainsAny(trimmed, "？?") && strings.Contains(trimmed, "但需") {
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
