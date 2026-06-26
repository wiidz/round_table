package engine

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

func TestParseAgendaSynthesisOutput(t *testing.T) {
	items := []event.AgendaItem{
		{ID: "skills", Title: "核心技能"},
		{ID: "engineering", Title: "工程约束"},
	}
	raw := `{
		"sections": [
			{"agenda_id":"skills","summary":["三连击"],"decisions":["冷却8秒"],"open_questions":["PVP？"]},
			{"agenda_id":"engineering","summary":["客户端预测"],"decisions":[],"open_questions":[]}
		],
		"cross_cutting": {"decisions":["统一日志规范"],"open_questions":["跨服延迟？"]}
	}`
	out, err := parseAgendaSynthesisOutput(raw, items)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Sections) != 2 {
		t.Fatalf("sections = %d", len(out.Sections))
	}
	if out.Sections[0].AgendaID != "skills" || len(out.Sections[0].Summary) != 1 {
		t.Fatalf("skills section = %+v", out.Sections[0])
	}
}

func TestAssembleDesignDraftFromAgenda(t *testing.T) {
	s := meeting.State{
		Topic: "影舞者",
		Goal:  "形成草案",
		Agenda: []event.AgendaItem{
			{ID: "skills", Title: "核心技能与资源机制"},
			{ID: "engineering", Title: "工程实现与平衡约束"},
		},
	}
	out := synthesisAgendaOutput{
		Sections: []synthesisAgendaSection{
			{AgendaID: "skills", Summary: []string{"连击点机制"}, Decisions: []string{"冷却 8 秒"}},
			{AgendaID: "engineering", Summary: []string{"客户端预测"}, OpenQuestions: []string{"同步频率？"}},
		},
		CrossCutting: synthesisAgendaCrossCutting{OpenQuestions: []string{"全局平衡？"}},
	}
	summary, open := assembleDesignDraftFromAgenda(s, out)
	if !strings.Contains(summary, "### 核心技能与资源机制") {
		t.Fatalf("missing agenda section header: %s", summary)
	}
	if !strings.Contains(summary, "连击点机制") {
		t.Fatal("missing skills summary")
	}
	if !strings.Contains(summary, "### 工程实现与平衡约束") {
		t.Fatal("missing engineering section")
	}
	if len(open) < 2 {
		t.Fatalf("open = %v", open)
	}
}

func TestSynthesizeDeliberationFinal_agendaLLMPath(t *testing.T) {
	items := []event.AgendaItem{
		{ID: "skills", Title: "核心技能"},
		{ID: "positioning", Title: "职业定位"},
	}
	e := &Engine{
		Model: synthesisFakeModel{content: `{
			"sections": [
				{"agenda_id":"skills","summary":["三连击 + 位移"],"decisions":[],"open_questions":[]},
				{"agenda_id":"positioning","summary":["高机动刺客"],"decisions":["放弃辅助路线"],"open_questions":[]}
			],
			"cross_cutting": {"decisions":[],"open_questions":["PVP 验证方式？"]}
		}`},
		ModelName: "test-model",
	}
	s := meeting.State{
		Topic:        "职业设计",
		Goal:         "形成草案",
		Agenda:       items,
		CurrentRound: 2,
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {"designer": {Content: "收束：三连击 + 位移。"}},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 2, Summary: "Round 2"}},
		},
	}
	summary, open, usage, _, err := e.synthesizeDeliberationFinal(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if usage == nil {
		t.Fatal("expected token usage")
	}
	if !strings.Contains(summary, "### 核心技能") {
		t.Fatalf("expected agenda section in summary: %s", summary)
	}
	if !strings.Contains(summary, "三连击") {
		t.Fatal("missing skills content")
	}
	if len(open) != 1 || !strings.Contains(open[0], "PVP") {
		t.Fatalf("open = %v", open)
	}
}

func TestRuleBasedAgendaSynthesis(t *testing.T) {
	s := meeting.State{
		Agenda: []event.AgendaItem{
			{ID: "skills", Title: "核心技能与资源机制"},
			{ID: "engineering", Title: "工程实现与平衡约束"},
		},
	}
	out := ruleBasedAgendaSynthesis(s,
		[]string{"核心技能采用连击点机制", "工程实现采用客户端预测降低延迟"},
		[]string{"统一冷却 8 秒"},
		[]string{"跨服延迟如何验证？"},
	)
	if len(out.Sections) != 2 {
		t.Fatalf("sections = %d", len(out.Sections))
	}
	if len(out.Sections[0].Summary) == 0 || !strings.Contains(out.Sections[0].Summary[0], "连击") {
		t.Fatalf("skills summary = %v", out.Sections[0].Summary)
	}
	if len(out.Sections[1].Summary) == 0 || !strings.Contains(out.Sections[1].Summary[0], "客户端") {
		t.Fatalf("engineering summary = %v", out.Sections[1].Summary)
	}
	if len(out.CrossCutting.OpenQuestions) != 1 || !strings.Contains(out.CrossCutting.OpenQuestions[0], "跨服") {
		t.Fatalf("cross open = %v", out.CrossCutting.OpenQuestions)
	}
}

func TestModeratorSynthesizeFinal_agendaRuleFallback(t *testing.T) {
	s := meeting.State{
		Topic: "职业设计",
		Agenda: []event.AgendaItem{
			{ID: "skills", Title: "核心技能与资源机制"},
		},
		CurrentRound: 1,
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {"a": {Content: "1. 核心技能采用三连击 + 位移机制。\n最终倾向：冷却 8 秒。"}},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 1, Summary: "round 1"}},
		},
	}
	summary, _, _ := moderatorSynthesizeFinal(s)
	if !strings.Contains(summary, "### 核心技能与资源机制") {
		t.Fatalf("expected agenda section header: %s", summary)
	}
}

func TestNormalizeAgendaSynthesisOutput_reordersByMeetingAgenda(t *testing.T) {
	items := []event.AgendaItem{
		{ID: "a", Title: "A"},
		{ID: "b", Title: "B"},
	}
	out := synthesisAgendaOutput{
		Sections: []synthesisAgendaSection{
			{AgendaID: "b", Summary: []string{"B only"}},
		},
	}
	got := normalizeAgendaSynthesisOutput(out, items)
	if len(got.Sections) != 2 {
		t.Fatalf("sections = %d", len(got.Sections))
	}
	if got.Sections[0].AgendaID != "a" || got.Sections[1].AgendaID != "b" {
		t.Fatalf("order = %v", got.Sections)
	}
}
