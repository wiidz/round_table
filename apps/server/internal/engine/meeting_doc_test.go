package engine

import (
	"strings"
	"testing"
	"time"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

func TestRenderMeetingDoc(t *testing.T) {
	s := meeting.State{
		ID:                  "mtg-1",
		Status:              meeting.StatusRunning,
		Topic:               "Auth Service 拆分评审",
		Goal:                "就是否拆分达成共识",
		StartedAt:           time.Date(2026, 6, 26, 14, 50, 0, 0, time.UTC),
		ConsensusStrategy:   "no_objection",
		ConfirmationMode:    meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment: 3,
		ParticipantOrder:    []string{"skeptic", "pragmatist"},
		Participants: map[string]meeting.ParticipantState{
			"skeptic":    {ID: "skeptic", Role: "Security Architect", Expertise: "security"},
			"pragmatist": {ID: "pragmatist", Role: "Tech Lead", Expertise: "delivery"},
		},
	}

	doc := renderMeetingDoc(s)
	for _, sub := range []string{
		"会议编号", "mtg-1", "会议时间", "会议主题", "Auth Service",
		"会议目标", "参会人员", "skeptic", "Security Architect", "会议流程", "Round 0",
	} {
		if !strings.Contains(doc, sub) {
			t.Fatalf("missing %q in:\n%s", sub, doc)
		}
	}
}

func TestParseBriefGoalFields(t *testing.T) {
	got := parseBriefGoalFields("输出优化调整方案和氪金点方案\n\n讨论范围：仅限于守爱游戏基座\n\n不在范围：实施排期\n\n完成标准：每议题至少 1 条结论")
	if got.Goal != "输出优化调整方案和氪金点方案" {
		t.Fatalf("goal=%q", got.Goal)
	}
	if got.InScope != "仅限于守爱游戏基座" {
		t.Fatalf("inScope=%q", got.InScope)
	}
	if got.OutOfScope != "实施排期" {
		t.Fatalf("outScope=%q", got.OutOfScope)
	}
	if got.DoneCriteria != "每议题至少 1 条结论" {
		t.Fatalf("done=%q", got.DoneCriteria)
	}
}

func TestRenderMeetingDoc_structuredBrief(t *testing.T) {
	s := meeting.State{
		ID:                  "mtg-brief",
		Status:              meeting.StatusCompleted,
		Topic:               "RO 二开讨论",
		MeetingMode:         meeting.MeetingModeDeliberation,
		Goal:                "输出方案草案\n\n讨论范围：概念层取舍\n\n完成标准：每议题 1 条结论",
		ConfirmationMode:    meeting.ConfirmationModeRequired,
		MaxRoundsPerSegment: 5,
		Agenda: []event.AgendaItem{
			{ID: "gacha", Title: "如何设计氪金方案（扭蛋、头饰、卡片）"},
			{ID: "balance", Title: "如何避免数值膨胀"},
		},
		ParticipantOrder: []string{"designer"},
		Participants: map[string]meeting.ParticipantState{
			"designer": {ID: "designer", Role: "游戏策划", Expertise: "gameplay"},
		},
	}
	doc := renderMeetingDoc(s)
	for _, want := range []string{
		"## 会议目标",
		"输出方案草案",
		"## 讨论议题",
		"1. 如何设计氪金方案（扭蛋、头饰、卡片）",
		"2. 如何避免数值膨胀",
		"## 讨论范围",
		"概念层取舍",
		"## 完成标准",
		"每议题 1 条结论",
		"## 会议流程",
	} {
		if !strings.Contains(doc, want) {
			t.Fatalf("missing %q in:\n%s", want, doc)
		}
	}
	for _, bad := range []string{"附加议程项", "## 议程\n", "(`gacha`)"} {
		if strings.Contains(doc, bad) {
			t.Fatalf("unexpected %q in:\n%s", bad, doc)
		}
	}
}
