package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"round_table/apps/server/internal/adapter/participant"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

func tokenUsageFromResponse(phase, participantID string, round, questionIndex int, resp participant.Response) *event.TokenUsage {
	if resp.Usage.TotalTokens == 0 && resp.Usage.PromptTokens == 0 && resp.Usage.CompletionTokens == 0 {
		return nil
	}
	return &event.TokenUsage{
		Model:            resp.Model,
		Phase:            phase,
		ParticipantID:    participantID,
		RoundNumber:      round,
		QuestionIndex:    questionIndex,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
}

func renderTokenUsageSummary(s meeting.State) string {
	var b strings.Builder
	b.WriteString("# Token Usage\n\n")
	if s.TokenUsageTotals.CallCount == 0 {
		b.WriteString("_No LLM calls recorded (stub participant or provider omitted usage)._\n")
		return b.String()
	}

	fmt.Fprintf(&b, "| 指标 | 数值 |\n|------|------|\n")
	fmt.Fprintf(&b, "| LLM 调用次数 | %d |\n", s.TokenUsageTotals.CallCount)
	fmt.Fprintf(&b, "| Prompt tokens | %d |\n", s.TokenUsageTotals.PromptTokens)
	fmt.Fprintf(&b, "| Completion tokens | %d |\n", s.TokenUsageTotals.CompletionTokens)
	fmt.Fprintf(&b, "| Total tokens | **%d** |\n\n", s.TokenUsageTotals.TotalTokens)

	b.WriteString("## 每次对话\n\n")
	b.WriteString("| # | 环节 | 参会者 | 模型 | Round | Prompt | Completion | Total |\n")
	b.WriteString("|---|------|--------|------|-------|--------|------------|-------|\n")
	for _, r := range s.TokenUsageLog {
		round := "—"
		if r.RoundNumber >= 0 {
			round = fmt.Sprintf("%d", r.RoundNumber)
		}
		model := r.Model
		if model == "" {
			model = "—"
		}
		fmt.Fprintf(&b, "| %d | %s | %s | %s | %s | %d | %d | %d |\n",
			r.Turn, r.Phase, r.ParticipantID, model, round,
			r.PromptTokens, r.CompletionTokens, r.TotalTokens)
	}
	b.WriteByte('\n')
	return b.String()
}

func renderTokenUsageJSONL(s meeting.State) []byte {
	if len(s.TokenUsageLog) == 0 {
		return nil
	}
	var b strings.Builder
	for _, r := range s.TokenUsageLog {
		line, _ := json.Marshal(r)
		b.Write(line)
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

func writeTokenUsageFiles(s meeting.State, write func(name string, body []byte) error) error {
	if err := write("usage/summary.md", []byte(renderTokenUsageSummary(s))); err != nil {
		return err
	}
	if data := renderTokenUsageJSONL(s); len(data) > 0 {
		if err := write("usage/tokens.jsonl", data); err != nil {
			return err
		}
	}
	return nil
}
