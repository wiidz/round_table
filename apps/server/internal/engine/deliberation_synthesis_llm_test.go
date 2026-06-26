package engine

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/domain/meeting"
)

type synthesisFakeModel struct {
	content string
	err     error
}

func (f synthesisFakeModel) Complete(_ context.Context, _ model.Request) (model.Response, error) {
	if f.err != nil {
		return model.Response{}, f.err
	}
	return model.Response{
		Content: f.content,
		Usage: model.Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}, nil
}

func TestParseSynthesisOutput(t *testing.T) {
	raw := `{"core_scheme":["方案 A"],"decisions":["采用三连击"],"open_questions":["冷却时间？"]}`
	out, err := parseSynthesisOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.CoreScheme) != 1 || len(out.Decisions) != 1 || len(out.OpenQuestions) != 1 {
		t.Fatalf("unexpected output: %+v", out)
	}

	wrapped := "```json\n" + raw + "\n```"
	out, err = parseSynthesisOutput(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Decisions) != 1 {
		t.Fatalf("wrapped parse failed: %+v", out)
	}
}

func TestSynthesizeDeliberationFinal_noModelUsesRules(t *testing.T) {
	e := &Engine{}
	s := meeting.State{
		Topic:        "测试主题",
		CurrentRound: 1,
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {"a": {Content: "最终倾向：采用方案 A。"}},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 1, Summary: "round 1"}},
		},
	}
	summary, open, usage, err := e.synthesizeDeliberationFinal(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if usage != nil {
		t.Fatal("expected no token usage without model")
	}
	if !strings.Contains(summary, "Executive Summary") {
		t.Fatalf("missing executive summary: %s", summary)
	}
	_ = open
}

func TestSynthesizeDeliberationFinal_llmPath(t *testing.T) {
	e := &Engine{
		Model: synthesisFakeModel{content: `{
			"core_scheme": ["核心：三连击 + 位移"],
			"decisions": ["统一冷却 8 秒"],
			"open_questions": ["PVP 平衡如何验证？"]
		}`},
		ModelName: "test-model",
	}
	s := meeting.State{
		Topic:        "职业设计",
		Goal:         "形成草案",
		CurrentRound: 2,
		ParticipantOrder: []string{"designer"},
		RoundOrder:       []string{"designer"},
		Participants: map[string]meeting.ParticipantState{
			"designer": {ID: "designer", Role: "策划"},
		},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {"designer": {Content: "收束：三连击 + 位移，冷却 8 秒。"}},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 2, Summary: "Round 2 summary"}},
		},
	}
	summary, open, usage, err := e.synthesizeDeliberationFinal(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if usage == nil || usage.TotalTokens != 150 {
		t.Fatalf("usage = %+v", usage)
	}
	if !strings.Contains(summary, "三连击") {
		t.Fatalf("missing core scheme: %s", summary)
	}
	if !strings.Contains(summary, "统一冷却") {
		t.Fatalf("missing decision: %s", summary)
	}
	if len(open) != 1 || !strings.Contains(open[0], "PVP") {
		t.Fatalf("open = %v", open)
	}
}

func TestSynthesizeDeliberationFinal_llmErrorFallsBack(t *testing.T) {
	e := &Engine{
		Model: synthesisFakeModel{err: context.DeadlineExceeded},
	}
	s := meeting.State{
		Topic:        "fallback",
		CurrentRound: 1,
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {"a": {Content: "最终倾向：方案 B。"}},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 1, Summary: "r1"}},
		},
	}
	summary, _, usage, err := e.synthesizeDeliberationFinal(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if usage != nil {
		t.Fatal("expected no usage on error fallback")
	}
	if !strings.Contains(summary, "Executive Summary") {
		t.Fatalf("expected rule fallback summary: %s", summary)
	}
}
