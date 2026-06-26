package engine

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/meeting"
)

func TestModeratorSynthesizeFinal_executiveSummary(t *testing.T) {
	s := meeting.State{
		Topic:              "影舞者技能设计",
		Goal:               "形成草案",
		CurrentRound:       2,
		MaxRoundsPerSegment: 2,
		ParticipantOrder:   []string{"designer", "tech_lead"},
		RoundOrder:         []string{"designer", "tech_lead"},
		Participants: map[string]meeting.ParticipantState{
			"designer":   {ID: "designer", Role: "策划"},
			"tech_lead":  {ID: "tech_lead", Role: "主程"},
		},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {
				"designer": {Content: "最终倾向：影印持续时间3秒，数量上限2个。\n\n**待决问题：**\n- 影印是否可被队友看见/交互？\n- 影闪位移是否无视障碍物（穿墙）？"},
			},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 2, Summary: "round 2"}},
		},
		ModeratorSummaries: map[int]string{
			2: "- 核心机制：影印计数\n- Q技能影步留下影印",
		},
	}
	summary, open := moderatorSynthesizeFinal(s)
	if !strings.Contains(summary, "## Executive Summary") {
		t.Fatal("missing executive summary")
	}
	if !strings.Contains(summary, "### 已决要点") {
		t.Fatal("missing decisions section")
	}
	if !strings.Contains(summary, "详细记录") {
		t.Fatal("missing detailed section")
	}
	if len(open) == 0 {
		t.Fatal("expected open questions")
	}
	for _, q := range open {
		if strings.HasPrefix(q, "待决问题") && !strings.Contains(q, "是否") {
			t.Fatalf("noise question: %q", q)
		}
	}
	foundInteract := false
	for _, q := range open {
		if strings.Contains(q, "队友") || strings.Contains(q, "穿墙") {
			foundInteract = true
		}
	}
	if !foundInteract {
		t.Fatalf("open questions = %v", open)
	}
}

func TestExtractOpenQuestionsFromText(t *testing.T) {
	text := `基于讨论：

**待决问题：**
- 影印是否可被队友看见/交互？
- 3. 待决问题列表：

最终倾向：影印3秒。`
	got := extractOpenQuestionsFromText(text)
	if len(got) != 1 {
		t.Fatalf("got %d questions: %v", len(got), got)
	}
	if !strings.Contains(got[0], "队友") {
		t.Fatalf("got %q", got[0])
	}
}

func TestCollectDeliberationDecisions(t *testing.T) {
	s := meeting.State{
		CurrentRound: 2,
		RoundOrder:   []string{"d"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {"d": {Content: "最终倾向：影印持续时间3秒，数量上限2个。"}},
		},
	}
	got := collectDeliberationDecisions(s)
	if len(got) == 0 {
		t.Fatal("expected decisions")
	}
}

func TestModeratorSynthesizeFinal(t *testing.T) {
	s := meeting.State{
		Topic:        "新职业：影舞者",
		Goal:         "形成技能方案草案",
		CurrentRound: 2,
		MaxRoundsPerSegment: 2,
		ParticipantOrder: []string{"designer", "player"},
		Participants: map[string]meeting.ParticipantState{
			"designer": {ID: "designer", Role: "策划"},
			"player":   {ID: "player", Role: "玩家代表"},
		},
		RoundOrder: []string{"designer", "player"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {
				"designer": {Content: "**待决问题：**\n- 影印是否可被敌方看到/踩灭？"},
			},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 2, Summary: "Round 2 summary"}},
		},
		ModeratorSummaries: map[int]string{2: "提炼 round 2"},
	}
	summary, open := moderatorSynthesizeFinal(s)
	if summary == "" {
		t.Fatal("empty summary")
	}
	if len(open) == 0 {
		t.Fatal("expected open questions")
	}
}
