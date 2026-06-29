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
	"round_table/apps/server/internal/llmjson"
	"round_table/apps/server/internal/stream"
)

const readinessSchema = `Respond ONLY with a JSON object (no markdown fences).
Do NOT use ASCII double quotes (") inside string values — use 「」 for emphasis if needed.
{
  "ready": false,
  "rationale": "one sentence in Chinese explaining the judgment",
  "gaps": ["blocking missing element or unresolved conflict, max 5; use [] when ready"]
}`

type readinessLLMOutput struct {
	Ready     bool     `json:"ready"`
	Rationale string   `json:"rationale"`
	Gaps      []string `json:"gaps"`
}

type deliberationReadinessResult struct {
	Ready     bool
	Rationale string
	Gaps      []string
	Usage     *event.TokenUsage
}

func synthesisResolvedBy(round, maxRounds int, ready bool) string {
	if round >= maxRounds {
		if ready {
			return "synthesis"
		}
		return "max_rounds"
	}
	if ready {
		return "readiness"
	}
	return ""
}

func (e *Engine) assessDeliberationReadiness(ctx context.Context, s meeting.State) (deliberationReadinessResult, error) {
	if e.Model == nil {
		return deliberationReadinessResult{
			Ready:     false,
			Rationale: "rule fallback: no model",
		}, nil
	}

	e.logf("◆ LLM synthesis readiness check (round %d)", s.CurrentRound)
	modelName := e.ModelName
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	prompt := buildDeliberationReadinessPrompt(s)
	system, err := e.buildModeratorReadinessSystem(s)
	if err != nil {
		return deliberationReadinessResult{}, err
	}

	phaseLabel := strings.TrimPrefix(PhaseDeliberationReadiness, "Phase: ")
	ctx = e.withStreamCtx(ctx, stream.Meta{
		ParticipantID: "moderator",
		Phase:         phaseLabel,
		Detail:        fmt.Sprintf("round %d readiness", s.CurrentRound),
	})

	start := time.Now()
	onDelta, onEnd := synthesisStreamHandlers(ctx)
	raw, err := e.Model.Complete(ctx, model.Request{
		Model: modelName,
		Messages: []model.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt + "\n\n" + readinessSchema},
		},
		Temperature: 0.2,
		OnDelta:     onDelta,
	})
	if err != nil {
		e.logf("◆ readiness check failed (%v) — treating as not ready", err)
		return deliberationReadinessResult{
			Ready:     false,
			Rationale: "LLM unavailable",
		}, nil
	}
	if onEnd != nil {
		onEnd()
	}
	e.logf("◆ readiness check done in %s", time.Since(start).Round(time.Millisecond))

	out, parseErr := ParseReadinessOutput(raw.Content)
	if parseErr != nil {
		e.logf("◆ readiness parse failed (%v) — treating as not ready", parseErr)
		return deliberationReadinessResult{
			Ready:     false,
			Rationale: "invalid readiness JSON",
		}, nil
	}

	gaps := normalizeSynthesisStrings(out.Gaps, 5)
	rationale := strings.TrimSpace(out.Rationale)
	if rationale == "" && out.Ready {
		rationale = "方案要素已足够合成草案"
	}
	if rationale == "" && !out.Ready {
		rationale = "仍需更多研讨"
	}

	usage := tokenUsageFromModel(phaseLabel, modelName, s.CurrentRound, raw.Usage)
	return deliberationReadinessResult{
		Ready:     out.Ready,
		Rationale: truncateRunes(rationale, 300),
		Gaps:      gaps,
		Usage:     usage,
	}, nil
}

func (e *Engine) buildModeratorReadinessSystem(s meeting.State) (string, error) {
	var b strings.Builder
	b.WriteString("You are the RoundTable Moderator judging whether a deliberation meeting has enough material to synthesize a design draft now.\n")
	b.WriteString("Output JSON only. Write rationale and gaps in Chinese (简体中文).\n\n")
	b.WriteString("Mark ready=true when:\n")
	b.WriteString("- Topic/Goal core elements are covered across rounds\n")
	if len(s.Agenda) > 0 {
		b.WriteString("- Each agenda item has enough substance to fill its synthesis section (gaps may remain as open_questions)\n")
	}
	b.WriteString("- Major conflicts are resolved OR can be captured as open_questions in the draft\n")
	b.WriteString("- Another debate round would likely add little new substance\n\n")
	b.WriteString("Mark ready=false when:\n")
	b.WriteString("- Blocking disagreements remain that would make a draft misleading\n")
	b.WriteString("- Essential scheme elements for the Topic are still missing\n")
	if len(s.Agenda) > 0 {
		b.WriteString("- A whole agenda item is still unaddressed with no tentative direction\n")
	}
	b.WriteString("- This round added no substantive new information and gaps remain\n")
	b.WriteString("- A participant's latest turn is primarily a direct question to another participant that still expects an in-meeting answer (not merely an open_question for the draft)\n")
	b.WriteString("- Recent free dialogue raised a new substantive disagreement or design fork not yet debated in a full round\n\n")
	b.WriteString("open_questions may remain in the final draft — do NOT require all issues to be closed.\n")
	b.WriteString(agendaReadinessSchemaHint(s))
	if e.Profile != nil {
		if data, err := e.Profile.ReadModerator(profile.FileAgents); err == nil {
			b.WriteString("\n--- Moderator AGENTS.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
}

func buildDeliberationReadinessPrompt(s meeting.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", PhaseDeliberationReadiness)
	b.WriteString(strings.TrimPrefix(buildDeliberationSynthesisPrompt(s, ""), PhaseDeliberationSynthesis+"\n"))
	b.WriteString("\nJudge whether synthesis should start now or another debate round is needed.\n")
	return b.String()
}

// ParseReadinessOutput parses deliberation readiness JSON with tolerant repair.
func ParseReadinessOutput(raw string) (readinessLLMOutput, error) {
	raw = llmjson.RepairObject(raw)

	var out readinessLLMOutput
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return out, nil
	}
	return readinessLLMOutput{}, fmt.Errorf("invalid readiness JSON")
}
