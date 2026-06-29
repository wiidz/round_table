package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

// DefaultGameClassAgenda is the generic deliberation agenda for the game-class-design scenario template.
var DefaultGameClassAgenda = []event.AgendaItem{
	{ID: "skills", Title: "核心技能与资源机制"},
	{ID: "positioning", Title: "职业定位与差异化"},
	{ID: "monetization", Title: "商业化与活动联动"},
	{ID: "engineering", Title: "工程实现与平衡约束"},
}

type synthesisAgendaSection struct {
	AgendaID      string   `json:"agenda_id"`
	Summary       []string `json:"summary"`
	Decisions     []string `json:"decisions"`
	OpenQuestions []string `json:"open_questions"`
}

type synthesisAgendaCrossCutting struct {
	Decisions     []string `json:"decisions"`
	OpenQuestions []string `json:"open_questions"`
}

type synthesisAgendaOutput struct {
	ExecutiveVerdict string                      `json:"executive_verdict"`
	KeyDecisions     []string                    `json:"key_decisions"`
	Sections         []synthesisAgendaSection    `json:"sections"`
	CrossCutting     synthesisAgendaCrossCutting `json:"cross_cutting"`
}

func hasAgendaForSynthesis(s meeting.State) bool {
	return len(s.Agenda) > 0
}

func writeDeliberationAgendaBlock(b *strings.Builder, s meeting.State) {
	if len(s.Agenda) == 0 {
		return
	}
	b.WriteString("## Agenda（按子项组织合成输出）\n\n")
	for _, item := range s.Agenda {
		id := agendaItemID(item)
		fmt.Fprintf(b, "- **%s** (agenda_id=%s)\n", item.Title, id)
	}
	b.WriteByte('\n')
}

func agendaItemID(item event.AgendaItem) string {
	if item.ID != "" {
		return item.ID
	}
	return normalizeQuestionKey(item.Title)
}

func agendaTitleByID(s meeting.State, id string) string {
	for _, item := range s.Agenda {
		if agendaItemID(item) == id {
			return item.Title
		}
	}
	return id
}

func buildAgendaSynthesisSchema(items []event.AgendaItem) string {
	var b strings.Builder
	b.WriteString(`Respond ONLY with a JSON object (no markdown fences).
Do NOT use ASCII double quotes (") inside string values — use 「」 for emphasis if needed.
{
  "executive_verdict": "1-3 sentences for Principal: recommended direction and what remains open",
  "key_decisions": ["up to 3 must-acknowledge items for Principal; may be []"],
  "sections": [
`)
	for i, item := range items {
		id := agendaItemID(item)
		comma := ","
		if i == len(items)-1 {
			comma = ""
		}
		fmt.Fprintf(&b, `    {
      "agenda_id": %q,
      "summary": ["2-4 bullets for this agenda item only"],
      "decisions": ["incremental agreements for this item not already in summary; may be []"],
      "open_questions": ["unresolved for this item; may be []"]
    }%s
`, id, comma)
	}
	b.WriteString(`  ],
  "cross_cutting": {
    "decisions": ["agreements spanning multiple agenda items; may be []"],
    "open_questions": ["global or unmapped open items; may be []"]
  }
}`)
	return b.String()
}

func parseAgendaSynthesisOutput(raw string, items []event.AgendaItem) (synthesisAgendaOutput, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	tryParse := func(data []byte) (synthesisAgendaOutput, error) {
		var out synthesisAgendaOutput
		if err := json.Unmarshal(data, &out); err != nil {
			return out, err
		}
		if !agendaSynthesisOutputNonEmpty(out) {
			return out, fmt.Errorf("empty agenda synthesis JSON")
		}
		return normalizeAgendaSynthesisOutput(out, items), nil
	}

	if out, err := tryParse([]byte(raw)); err == nil {
		return out, nil
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return tryParse([]byte(raw[start : end+1]))
	}
	return synthesisAgendaOutput{}, fmt.Errorf("invalid agenda synthesis JSON")
}

func agendaSynthesisOutputNonEmpty(out synthesisAgendaOutput) bool {
	if strings.TrimSpace(out.ExecutiveVerdict) != "" || len(out.KeyDecisions) > 0 {
		return true
	}
	for _, sec := range out.Sections {
		if len(sec.Summary)+len(sec.Decisions)+len(sec.OpenQuestions) > 0 {
			return true
		}
	}
	return len(out.CrossCutting.Decisions)+len(out.CrossCutting.OpenQuestions) > 0
}

func normalizeAgendaSynthesisOutput(out synthesisAgendaOutput, items []event.AgendaItem) synthesisAgendaOutput {
	byID := make(map[string]synthesisAgendaSection, len(out.Sections))
	for _, sec := range out.Sections {
		id := strings.TrimSpace(sec.AgendaID)
		if id == "" {
			continue
		}
		byID[id] = sec
	}
	out.Sections = out.Sections[:0]
	for _, item := range items {
		id := agendaItemID(item)
		sec, ok := byID[id]
		if !ok {
			sec = synthesisAgendaSection{AgendaID: id}
		}
		sec.AgendaID = id
		sec.Summary = normalizeSynthesisStrings(sec.Summary, 6)
		sec.Decisions = normalizeSynthesisStrings(sec.Decisions, 8)
		sec.OpenQuestions = normalizeSynthesisStrings(sec.OpenQuestions, 6)
		sec.Decisions = dedupeDecisionsAgainstCoreScheme(sec.Summary, sec.Decisions)
		firm, open := splitTentativeDecisions(sec.Decisions, sec.OpenQuestions)
		sec.Decisions = firm
		sec.OpenQuestions = open
		out.Sections = append(out.Sections, sec)
	}
	out.CrossCutting.Decisions = normalizeSynthesisStrings(out.CrossCutting.Decisions, 8)
	out.CrossCutting.OpenQuestions = normalizeSynthesisStrings(out.CrossCutting.OpenQuestions, 8)
	firm, open := splitTentativeDecisions(out.CrossCutting.Decisions, out.CrossCutting.OpenQuestions)
	out.CrossCutting.Decisions = firm
	out.CrossCutting.OpenQuestions = open
	out.ExecutiveVerdict = strings.TrimSpace(out.ExecutiveVerdict)
	out.KeyDecisions = normalizeSynthesisStrings(out.KeyDecisions, 3)
	return out
}

func assembleDesignDraftFromAgenda(s meeting.State, out synthesisAgendaOutput) (summary string, openQuestions []string) {
	var b strings.Builder
	b.WriteString("# 方案草案\n\n")
	writeAgendaExecutiveSummary(&b, s, out)

	openQuestions = collectAgendaOpenQuestions(out)
	return strings.TrimSpace(b.String()), openQuestions
}

func synthesisAgendaOutputToEvent(out synthesisAgendaOutput) ([]event.SynthesisAgendaSectionPayload, *event.SynthesisCrossCuttingPayload) {
	sections := make([]event.SynthesisAgendaSectionPayload, 0, len(out.Sections))
	for _, sec := range out.Sections {
		sections = append(sections, event.SynthesisAgendaSectionPayload{
			AgendaID:      sec.AgendaID,
			Summary:       append([]string(nil), sec.Summary...),
			Decisions:     append([]string(nil), sec.Decisions...),
			OpenQuestions: append([]string(nil), sec.OpenQuestions...),
		})
	}
	if len(out.CrossCutting.Decisions) == 0 && len(out.CrossCutting.OpenQuestions) == 0 {
		return sections, nil
	}
	return sections, &event.SynthesisCrossCuttingPayload{
		Decisions:     append([]string(nil), out.CrossCutting.Decisions...),
		OpenQuestions: append([]string(nil), out.CrossCutting.OpenQuestions...),
	}
}

func formatAgendaSectionBody(summary, decisions, openQuestions []string) string {
	var b strings.Builder
	b.WriteString("**方案要点**\n\n")
	if len(summary) > 0 {
		for _, line := range summary {
			fmt.Fprintf(&b, "- %s\n", line)
		}
	} else {
		b.WriteString("- （本议程项讨论不足 — 见 MINUTES.md）\n")
	}
	b.WriteByte('\n')

	b.WriteString("**已决要点**\n\n")
	if len(decisions) > 0 {
		for _, d := range decisions {
			fmt.Fprintf(&b, "- %s\n", d)
		}
	} else if len(summary) > 0 {
		b.WriteString("- （已并入方案要点 — 见各轮发言）\n")
	} else {
		b.WriteString("- （未形成明确决议 — 见各轮发言）\n")
	}
	if len(openQuestions) > 0 {
		b.WriteString("\n**待决事项**\n\n")
		for _, q := range openQuestions {
			fmt.Fprintf(&b, "- %s\n", q)
		}
	}
	return strings.TrimSpace(b.String())
}

func formatCrossCuttingSectionBody(decisions, openQuestions []string) string {
	var b strings.Builder
	if len(decisions) > 0 {
		b.WriteString("**已决要点**\n\n")
		for _, d := range decisions {
			fmt.Fprintf(&b, "- %s\n", d)
		}
	}
	if len(openQuestions) > 0 {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("**待决事项**\n\n")
		for _, q := range openQuestions {
			fmt.Fprintf(&b, "- %s\n", q)
		}
	}
	return strings.TrimSpace(b.String())
}

func writeAgendaExecutiveSummary(b *strings.Builder, s meeting.State, out synthesisAgendaOutput) {
	b.WriteString("## Executive Summary\n\n")
	fmt.Fprintf(b, "**主题**：%s\n\n", s.Topic)
	if s.Goal != "" {
		fmt.Fprintf(b, "**目标**：%s\n\n", s.Goal)
	}

	verdict, keyDecisions := out.ExecutiveVerdict, out.KeyDecisions
	if strings.TrimSpace(verdict) == "" && len(keyDecisions) == 0 {
		allDecisions, allOpen := collectAgendaDecisionsForVerdict(out)
		coreScheme := formatBulletList(collectAgendaSummaryBullets(out), 6)
		verdict, keyDecisions = deriveExecutiveVerdict(s.Topic, coreScheme, allDecisions, allOpen)
	}
	writeExecutiveVerdictBlock(b, verdict, keyDecisions)

	for _, sec := range out.Sections {
		title := agendaTitleByID(s, sec.AgendaID)
		fmt.Fprintf(b, "### %s\n\n", title)
		b.WriteString(formatAgendaSectionBody(sec.Summary, sec.Decisions, sec.OpenQuestions))
		b.WriteString("\n\n")
	}

	if len(out.CrossCutting.Decisions) > 0 || len(out.CrossCutting.OpenQuestions) > 0 {
		b.WriteString("### 跨议程事项\n\n")
		b.WriteString(formatCrossCuttingSectionBody(out.CrossCutting.Decisions, out.CrossCutting.OpenQuestions))
		b.WriteString("\n\n")
	}

	b.WriteString("\n" + designDraftRecordFooter)
}

func collectAgendaSummaryBullets(out synthesisAgendaOutput) []string {
	var bullets []string
	for _, sec := range out.Sections {
		bullets = append(bullets, sec.Summary...)
	}
	return bullets
}

func collectAgendaDecisionsForVerdict(out synthesisAgendaOutput) (decisions, openQuestions []string) {
	for _, sec := range out.Sections {
		decisions = append(decisions, sec.Decisions...)
		openQuestions = append(openQuestions, sec.OpenQuestions...)
	}
	decisions = append(decisions, out.CrossCutting.Decisions...)
	openQuestions = append(openQuestions, out.CrossCutting.OpenQuestions...)
	return decisions, openQuestions
}

func collectAgendaOpenQuestions(out synthesisAgendaOutput) []string {
	seen := make(map[string]bool)
	var all []string
	appendUnique := func(items []string) {
		for _, q := range items {
			q = strings.TrimSpace(q)
			if q == "" {
				continue
			}
			if overlapsOpenQuestion(q, all) {
				continue
			}
			key := normalizeQuestionKey(q)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			all = append(all, q)
		}
	}
	for _, sec := range out.Sections {
		appendUnique(sec.OpenQuestions)
	}
	appendUnique(out.CrossCutting.OpenQuestions)
	if len(all) > 12 {
		all = all[:12]
	}
	return dedupeOpenQuestions(all)
}

func dedupeOpenQuestions(items []string) []string {
	var out []string
	for _, q := range items {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if overlapsOpenQuestion(q, out) {
			continue
		}
		out = append(out, q)
	}
	return out
}

func agendaReadinessSchemaHint(s meeting.State) string {
	if len(s.Agenda) == 0 {
		return ""
	}
	var ids []string
	for _, item := range s.Agenda {
		ids = append(ids, agendaItemID(item))
	}
	return fmt.Sprintf("\nWhen judging gaps, reference agenda coverage (%s). A missing agenda item with blocking conflict should yield ready=false.\n", strings.Join(ids, ", "))
}

func ruleBasedAgendaSynthesis(s meeting.State, coreBullets, decisions, openQuestions []string) synthesisAgendaOutput {
	var out synthesisAgendaOutput
	assignedDecisions := make(map[int]bool)
	assignedOpen := make(map[int]bool)
	assignedCore := make(map[int]bool)

	for _, item := range s.Agenda {
		sec := synthesisAgendaSection{AgendaID: agendaItemID(item)}
		for i, b := range coreBullets {
			if assignedCore[i] {
				continue
			}
			if textMatchesAgendaItem(b, item) {
				sec.Summary = append(sec.Summary, b)
				assignedCore[i] = true
			}
		}
		for i, d := range decisions {
			if assignedDecisions[i] {
				continue
			}
			if textMatchesAgendaItem(d, item) {
				sec.Decisions = append(sec.Decisions, d)
				assignedDecisions[i] = true
			}
		}
		for i, q := range openQuestions {
			if assignedOpen[i] {
				continue
			}
			if textMatchesAgendaItem(q, item) {
				sec.OpenQuestions = append(sec.OpenQuestions, q)
				assignedOpen[i] = true
			}
		}
		sec.Summary = normalizeSynthesisStrings(sec.Summary, 6)
		sec.Decisions = normalizeSynthesisStrings(sec.Decisions, 8)
		sec.OpenQuestions = normalizeSynthesisStrings(sec.OpenQuestions, 6)
		out.Sections = append(out.Sections, sec)
	}

	for i, d := range decisions {
		if !assignedDecisions[i] {
			out.CrossCutting.Decisions = append(out.CrossCutting.Decisions, d)
		}
	}
	for i, q := range openQuestions {
		if !assignedOpen[i] {
			out.CrossCutting.OpenQuestions = append(out.CrossCutting.OpenQuestions, q)
		}
	}
	for i, b := range coreBullets {
		if !assignedCore[i] {
			out.CrossCutting.Decisions = append(out.CrossCutting.Decisions, b)
		}
	}
	out.CrossCutting.Decisions = normalizeSynthesisStrings(out.CrossCutting.Decisions, 8)
	out.CrossCutting.OpenQuestions = normalizeSynthesisStrings(out.CrossCutting.OpenQuestions, 8)
	return out
}

func textMatchesAgendaItem(text string, item event.AgendaItem) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	if item.ID != "" && strings.Contains(strings.ToLower(text), strings.ToLower(item.ID)) {
		return true
	}
	if deliberationTokenOverlap(text, item.Title) >= 2 {
		return true
	}
	for _, part := range strings.FieldsFunc(item.Title, func(r rune) bool {
		return r == '与' || r == '及' || r == '、' || r == '/' || r == ' '
	}) {
		part = strings.TrimSpace(part)
		if len([]rune(part)) >= 2 && strings.Contains(text, part) {
			return true
		}
	}
	return false
}
