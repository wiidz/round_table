package engine

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

type roundSummaryModel struct {
	summary string
}

func (m roundSummaryModel) Complete(_ context.Context, req model.Request) (model.Response, error) {
	for _, msg := range req.Messages {
		if strings.Contains(msg.Content, PhaseModeratorRoundSummary) ||
			strings.Contains(msg.Content, PhaseModeratorExecutiveRecap) {
			return model.Response{Content: m.summary}, nil
		}
	}
	return model.Response{Content: ""}, nil
}

func TestModeratorRoundSummary_llmEnabled(t *testing.T) {
	e := &Engine{
		Model:                    roundSummaryModel{summary: "## Round 1 研讨摘要\n\n### 本轮进展\n专家就骑乘机制达成初步共识。\n\n### 已趋同\n- 骑乘为骑士专属\n\n### 仍存分歧\n- 无\n\n### 待决与下轮焦点\n- 公共 CD 待测"},
		ModelName:                "test",
		LLMModeratorRoundSummary: true,
	}
	s := meeting.State{
		ID:           "mtg-1",
		Topic:        "骑士设计",
		MeetingMode:  meeting.MeetingModeDeliberation,
		CurrentRound: 1,
		MaxRoundsPerSegment: 3,
		RoundOrder:   []string{"player", "dev"},
		Participants: map[string]meeting.ParticipantState{
			"player": {Role: "玩家"},
			"dev":    {Role: "开发"},
		},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {
				"player": {Content: "1. 保留经典技能\n2. 流派多样"},
				"dev":    {Content: "1. 服务端校验 CD"},
			},
		},
	}
	got := e.moderatorRoundSummary(context.Background(), s)
	if !strings.Contains(got, "本轮进展") || strings.Contains(got, "各角色贡献") {
		t.Fatalf("expected LLM summary, got:\n%s", got)
	}
}

func TestModeratorRoundSummary_disabledUsesRules(t *testing.T) {
	e := &Engine{
		Model:                    roundSummaryModel{summary: "## Round 1 研讨摘要\n\n### 本轮进展\nignored"},
		LLMModeratorRoundSummary: false,
	}
	s := meeting.State{
		MeetingMode:  meeting.MeetingModeDeliberation,
		CurrentRound: 1,
		MaxRoundsPerSegment: 2,
		RoundOrder:   []string{"dev"},
		Participants: map[string]meeting.ParticipantState{"dev": {Role: "开发"}},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {"dev": {Content: "1. 服务端校验"}},
		},
	}
	got := e.moderatorRoundSummary(context.Background(), s)
	if !strings.Contains(got, "各角色贡献") {
		t.Fatalf("expected rule fallback, got:\n%s", got)
	}
}

func TestModeratorRoundSummary_llmFailureFallback(t *testing.T) {
	e := &Engine{
		Model:                    roundSummaryModel{summary: "{}"},
		LLMModeratorRoundSummary: true,
	}
	s := meeting.State{
		MeetingMode:  meeting.MeetingModeDeliberation,
		CurrentRound: 1,
		MaxRoundsPerSegment: 2,
		RoundOrder:   []string{"dev"},
		Participants: map[string]meeting.ParticipantState{"dev": {Role: "开发"}},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {"dev": {Content: "1. 测试点"}},
		},
	}
	got := e.moderatorRoundSummary(context.Background(), s)
	if !strings.Contains(got, "各角色贡献") {
		t.Fatalf("expected rule fallback on bad LLM output, got:\n%s", got)
	}
}

func TestBuildModeratorRoundSummaryPrompt_includesStances(t *testing.T) {
	e := &Engine{}
	s := meeting.State{
		Topic:        "API",
		MeetingMode:  meeting.MeetingModeDecision,
		CurrentRound: 1,
		RoundOrder:   []string{"a"},
		Participants: map[string]meeting.ParticipantState{"a": {Role: "Arch"}},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {"a": {Content: "Need tests", Stance: event.StanceObject, ObjectReason: "coverage"}},
		},
	}
	prompt := e.buildModeratorRoundSummaryPrompt(s)
	if !strings.Contains(prompt, "Stance: object") || !strings.Contains(prompt, "coverage") {
		t.Fatalf("prompt missing stance:\n%s", prompt)
	}
}

func TestCleanRoundSummaryOutput_unwrapsJSONContent(t *testing.T) {
	raw := `{"content":"## 会议回顾\n\n### 目标与议程覆盖\n已覆盖全部议程项，讨论围绕核心模块展开并形成可操作方案雏形。","stance":"none","object_reason":""}`
	got := cleanRoundSummaryOutput(raw)
	if !strings.Contains(got, "目标与议程覆盖") {
		t.Fatalf("got=%q", got)
	}
	if strings.HasPrefix(got, "{") {
		t.Fatalf("should unwrap JSON: %q", got)
	}
}
