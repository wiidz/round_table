package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/stream"
)

const synthesisSchema = `Respond ONLY with a JSON object (no markdown fences).
Do NOT use ASCII double quotes (") inside string values — use 「」 for emphasis if needed.
{
  "core_scheme": ["3-6 bullets summarizing the latest substantive proposal, not engineering feasibility alone"],
  "decisions": ["incremental agreements only — items not already covered in core_scheme; use [] if none"],
  "open_questions": ["unresolved or disputed items, including anything still needing confirmation, max 8 items"]
}`

type synthesisLLMOutput struct {
	CoreScheme    []string `json:"core_scheme"`
	Decisions     []string `json:"decisions"`
	OpenQuestions []string `json:"open_questions"`
}

func (e *Engine) synthesizeDeliberationFinal(ctx context.Context, s meeting.State) (summary string, openQuestions []string, usage *event.TokenUsage, err error) {
	if e.Model == nil {
		summary, openQuestions = moderatorSynthesizeFinal(s)
		return summary, openQuestions, nil, nil
	}

	e.logf("◆ LLM synthesis (moderator)")
	modelName := e.ModelName
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	prompt := buildDeliberationSynthesisPrompt(s)
	system, err := e.buildModeratorSynthesisSystem()
	if err != nil {
		return "", nil, nil, err
	}

	phaseLabel := strings.TrimPrefix(PhaseDeliberationSynthesis, "Phase: ")
	ctx = e.withStreamCtx(ctx, stream.Meta{
		ParticipantID: "moderator",
		Phase:         phaseLabel,
		Detail:        "design-draft synthesis",
	})

	start := time.Now()
	onDelta, onEnd := synthesisStreamHandlers(ctx)
	raw, err := e.Model.Complete(ctx, model.Request{
		Model: modelName,
		Messages: []model.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt + "\n\n" + synthesisSchema},
		},
		Temperature: 0.3,
		OnDelta:     onDelta,
	})
	if err != nil {
		e.logf("◆ LLM synthesis failed (%v) — falling back to rule-based", err)
		summary, openQuestions = moderatorSynthesizeFinal(s)
		return summary, openQuestions, nil, nil
	}
	if onEnd != nil {
		onEnd()
	}
	e.logf("◆ LLM synthesis done in %s", time.Since(start).Round(time.Millisecond))

	out, parseErr := parseSynthesisOutput(raw.Content)
	if parseErr != nil {
		e.logf("◆ LLM synthesis parse failed (%v) — falling back to rule-based", parseErr)
		summary, openQuestions = moderatorSynthesizeFinal(s)
		return summary, openQuestions, tokenUsageFromModel(phaseLabel, modelName, s.CurrentRound, raw.Usage), nil
	}

	coreItems := normalizeSynthesisStrings(out.CoreScheme, 6)
	decisions, openQuestions := splitTentativeDecisions(
		normalizeSynthesisStrings(out.Decisions, 10),
		normalizeSynthesisStrings(out.OpenQuestions, 8),
	)
	decisions = dedupeDecisionsAgainstCoreScheme(coreItems, decisions)
	coreScheme := formatBulletList(coreItems, 6)
	summary = assembleDesignDraft(s, coreScheme, decisions, openQuestions)

	usage = tokenUsageFromModel(phaseLabel, modelName, s.CurrentRound, raw.Usage)
	return summary, openQuestions, usage, nil
}

func (e *Engine) buildModeratorSynthesisSystem() (string, error) {
	var b strings.Builder
	b.WriteString("You are the RoundTable Moderator synthesizing a deliberation meeting into a design draft.\n")
	b.WriteString("Output JSON only. Write content in Chinese (简体中文).\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- core_scheme: design snapshot — what the final scheme looks like (structure, resources, mechanics).\n")
	b.WriteString("- decisions: incremental agreements from the record that are NOT already in core_scheme. May be empty.\n")
	b.WriteString("- decisions: exclude items with 留待讨论/待确认/未表态 (move those to open_questions).\n")
	b.WriteString("- open_questions: unresolved, disputed, or needs-confirmation items; move tentative decisions here.\n")
	b.WriteString("- Do not invent facts absent from the meeting record.\n\n")
	if e.Profile != nil {
		if data, err := e.Profile.ReadModerator(profile.FileAgents); err == nil {
			b.WriteString("--- Moderator AGENTS.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
}

func buildDeliberationSynthesisPrompt(s meeting.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", PhaseDeliberationSynthesis)
	fmt.Fprintf(&b, "Topic: %s\n", s.Topic)
	if s.Goal != "" {
		fmt.Fprintf(&b, "Goal: %s\n", s.Goal)
	}
	fmt.Fprintf(&b, "Rounds completed: %d\n\n", s.CurrentRound)

	if s.PreMeetingSummary != "" {
		b.WriteString("## Pre-meeting\n\n")
		b.WriteString(strings.TrimSpace(s.PreMeetingSummary))
		b.WriteString("\n\n")
	}

	for _, r := range s.Minutes.Rounds {
		if r.RoundNumber <= 0 {
			continue
		}
		fmt.Fprintf(&b, "## Round %d transcript\n\n", r.RoundNumber)
		b.WriteString(strings.TrimSpace(r.Summary))
		b.WriteByte('\n')
		if mod, ok := s.ModeratorSummaries[r.RoundNumber]; ok {
			b.WriteString("\n### Moderator summary\n\n")
			b.WriteString(strings.TrimSpace(mod))
		}
		b.WriteString("\n\n")
		for _, id := range s.RoundOrder {
			resp, ok := s.RoundResponses[r.RoundNumber][id]
			if !ok {
				continue
			}
			role := s.Participants[id].Role
			fmt.Fprintf(&b, "### %s (%s)\n\n%s\n\n", id, role, strings.TrimSpace(resp.Content))
		}
	}

	if s.FreeDialogueCompleted && s.FreeDialogueSummary != "" {
		b.WriteString("## Free dialogue\n\n")
		b.WriteString(strings.TrimSpace(s.FreeDialogueSummary))
		b.WriteString("\n\n")
	}

	b.WriteString("Synthesize the design draft executive summary fields from the record above.\n")
	return b.String()
}

func parseSynthesisOutput(raw string) (synthesisLLMOutput, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var out synthesisLLMOutput
	if err := json.Unmarshal([]byte(raw), &out); err == nil && synthesisOutputNonEmpty(out) {
		return out, nil
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(raw[start:end+1]), &out); err == nil && synthesisOutputNonEmpty(out) {
			return out, nil
		}
	}
	return synthesisLLMOutput{}, fmt.Errorf("invalid synthesis JSON")
}

func synthesisOutputNonEmpty(out synthesisLLMOutput) bool {
	return len(out.CoreScheme)+len(out.Decisions)+len(out.OpenQuestions) > 0
}

var tentativeDecisionMarkers = []string{
	"留待讨论", "待讨论", "待确认", "需确认", "尚需确认", "待最终确认",
	"未明确表态", "未表态", "需讨论定夺", "需讨论",
}

func splitTentativeDecisions(decisions, openQuestions []string) ([]string, []string) {
	seen := make(map[string]bool)
	for _, q := range openQuestions {
		seen[normalizeQuestionKey(q)] = true
	}
	var firm, open []string
	open = append(open, openQuestions...)
	for _, d := range decisions {
		tentative := false
		for _, m := range tentativeDecisionMarkers {
			if strings.Contains(d, m) {
				tentative = true
				break
			}
		}
		if tentative {
			key := normalizeQuestionKey(d)
			if key != "" && !seen[key] {
				seen[key] = true
				open = append(open, d)
			}
			continue
		}
		firm = append(firm, d)
	}
	return firm, open
}

func normalizeSynthesisStrings(items []string, max int) []string {
	seen := make(map[string]bool)
	var out []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := normalizeQuestionKey(item)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, truncateRunes(item, 200))
		if len(out) >= max {
			break
		}
	}
	return out
}

func tokenUsageFromModel(phase, modelName string, round int, usage model.Usage) *event.TokenUsage {
	if usage.TotalTokens == 0 && usage.PromptTokens == 0 && usage.CompletionTokens == 0 {
		return nil
	}
	return &event.TokenUsage{
		Model:            modelName,
		Phase:            phase,
		ParticipantID:    "moderator",
		RoundNumber:      round,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
}

func synthesisStreamHandlers(ctx context.Context) (model.StreamHandler, func()) {
	h, ok := stream.HandlersFrom(ctx)
	if !ok {
		return nil, nil
	}
	if h.OnStart != nil {
		h.OnStart(h.Meta)
	}
	return h.OnDelta, h.OnEnd
}
