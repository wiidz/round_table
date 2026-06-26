package engine

import (
	"strings"
	"testing"
	"time"

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
		"会议目标", "参会人员", "skeptic", "Security Architect", "议程", "Round 0",
	} {
		if !strings.Contains(doc, sub) {
			t.Fatalf("missing %q in:\n%s", sub, doc)
		}
	}
}
