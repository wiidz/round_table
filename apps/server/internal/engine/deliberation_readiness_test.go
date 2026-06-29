package engine

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/domain/meeting"
)

type phaseRoutingModel struct {
	readiness string
	synthesis string
}

func (m phaseRoutingModel) Complete(_ context.Context, req model.Request) (model.Response, error) {
	content := m.synthesis
	for _, msg := range req.Messages {
		if strings.Contains(msg.Content, PhaseDeliberationReadiness) {
			content = m.readiness
			break
		}
	}
	return model.Response{
		Content: content,
		Usage:   model.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	}, nil
}

func TestSynthesisResolvedBy(t *testing.T) {
	tests := []struct {
		round, max int
		ready      bool
		want       string
	}{
		{2, 5, true, "readiness"},
		{5, 5, true, "synthesis"},
		{5, 5, false, "max_rounds"},
		{3, 5, false, ""},
	}
	for _, tc := range tests {
		got := synthesisResolvedBy(tc.round, tc.max, tc.ready)
		if got != tc.want {
			t.Fatalf("round=%d max=%d ready=%v: got %q want %q", tc.round, tc.max, tc.ready, got, tc.want)
		}
	}
}

func TestParseReadinessOutput(t *testing.T) {
	raw := `{"ready": true, "rationale": "要素已齐", "gaps": []}`
	out, err := ParseReadinessOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !out.Ready || out.Rationale == "" {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestAssessDeliberationReadiness_noModel(t *testing.T) {
	e := &Engine{}
	got, err := e.assessDeliberationReadiness(context.Background(), meeting.State{CurrentRound: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got.Ready {
		t.Fatal("expected not ready without model")
	}
}

func TestAssessDeliberationReadiness_llmReady(t *testing.T) {
	e := &Engine{
		Model: phaseRoutingModel{
			readiness: `{"ready": true, "rationale": "方案要素已覆盖", "gaps": []}`,
		},
		ModelName: "test-model",
	}
	got, err := e.assessDeliberationReadiness(context.Background(), meeting.State{
		Topic:        "测试",
		CurrentRound: 2,
		MaxRoundsPerSegment: 5,
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {"a": {Content: "收束：方案 A。"}},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 2, Summary: "round 2"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ready {
		t.Fatalf("expected ready, got %+v", got)
	}
	if got.Usage == nil {
		t.Fatal("expected token usage")
	}
}

func TestAssessDeliberationReadiness_parseFailure(t *testing.T) {
	e := &Engine{
		Model:     phaseRoutingModel{readiness: "not json"},
		ModelName: "test-model",
	}
	got, err := e.assessDeliberationReadiness(context.Background(), meeting.State{Topic: "x", CurrentRound: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got.Ready {
		t.Fatal("expected not ready on parse failure")
	}
}
