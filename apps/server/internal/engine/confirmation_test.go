package engine

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

func TestPrepareDeliberationConfirmationBrief_agendaSections(t *testing.T) {
	s := meeting.State{
		Topic:       "影舞者职业设计",
		MeetingMode: meeting.MeetingModeDeliberation,
		Agenda: []event.AgendaItem{
			{ID: "skills", Title: "核心技能与资源机制"},
			{ID: "engineering", Title: "工程实现与平衡约束"},
		},
		SynthesisSections: []event.SynthesisAgendaSectionPayload{
			{
				AgendaID:  "skills",
				Summary:   []string{"连击点机制"},
				Decisions: []string{"冷却 8 秒"},
			},
			{
				AgendaID:      "engineering",
				Summary:       []string{"客户端预测"},
				OpenQuestions: []string{"同步频率？"},
			},
		},
		SynthesisCrossCutting: &event.SynthesisCrossCuttingPayload{
			OpenQuestions: []string{"全局平衡？"},
		},
	}

	brief := prepareDeliberationConfirmationBrief(s)
	if len(brief.Items) != 3 {
		t.Fatalf("items = %d, want 3 (2 agenda + cross_cutting)", len(brief.Items))
	}
	if brief.Items[0].Title != "核心技能与资源机制" {
		t.Fatalf("item[0] title = %q", brief.Items[0].Title)
	}
	if !strings.Contains(brief.Items[0].Description, "连击点机制") {
		t.Fatalf("item[0] desc = %q", brief.Items[0].Description)
	}
	if brief.Items[1].Title != "工程实现与平衡约束" {
		t.Fatalf("item[1] title = %q", brief.Items[1].Title)
	}
	if brief.Items[2].Title != "跨议程事项" {
		t.Fatalf("item[2] title = %q", brief.Items[2].Title)
	}
	if !strings.Contains(brief.ExecutiveSummary, "按议程逐项") {
		t.Fatalf("executive = %q", brief.ExecutiveSummary)
	}
}

func TestPrepareDeliberationConfirmationBrief_flatFallback(t *testing.T) {
	s := meeting.State{
		Topic:            "无议程议题",
		MeetingMode:      meeting.MeetingModeDeliberation,
		SynthesisSummary: "# 方案草案\n\n内容",
		SynthesisOpenQuestions: []string{"待确认 A"},
	}
	brief := prepareDeliberationConfirmationBrief(s)
	if len(brief.Items) != 2 {
		t.Fatalf("items = %d, want 2 flat items", len(brief.Items))
	}
	if brief.Items[0].Title != "方案草案" {
		t.Fatalf("item[0] title = %q", brief.Items[0].Title)
	}
}
