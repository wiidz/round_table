package engine

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

func TestModeratorSummarizeRound_synthesized(t *testing.T) {
	s := meeting.State{
		CurrentRound: 1,
		RoundOrder:   []string{"a", "b"},
		Participants: map[string]meeting.ParticipantState{
			"a": {Role: "Architect"},
			"b": {Role: "PM"},
		},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {
				"a": {
					Content:      "1. JWT 泄露面未定义\n2. 租户边界缺失",
					Stance:       event.StanceObject,
					ObjectReason: "安全基线不足",
				},
				"b": {
					Content: "1. 使用 RS256\n2. 租户级 Redis key",
					Stance:  event.StanceAgree,
				},
			},
		},
	}
	out := moderatorSummarizeRound(s)
	if strings.Contains(out, "Moderator summary — Round") {
		t.Fatal("should not use old copy-paste header")
	}
	if !strings.Contains(out, "未解决的分歧") || !strings.Contains(out, "安全基线不足") {
		t.Fatalf("missing objection synthesis:\n%s", out)
	}
	if !strings.Contains(out, "RS256") {
		t.Fatalf("missing mitigation bullets:\n%s", out)
	}
	if strings.Count(out, "1. JWT") > 0 {
		t.Fatalf("should not paste raw numbered list:\n%s", out)
	}
}

func TestModeratorSummarizeRound(t *testing.T) {
	s := meeting.State{
		CurrentRound: 1,
		RoundOrder:   []string{"a", "b"},
		Participants: map[string]meeting.ParticipantState{
			"a": {Role: "Architect"},
			"b": {Role: "PM"},
		},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {
				"a": {Content: "Looks good", Stance: event.StanceAgree},
				"b": {Content: "Need tests", Stance: event.StanceObject, ObjectReason: "missing coverage"},
			},
		},
	}
	out := moderatorSummarizeRound(s)
	if !strings.Contains(out, "未解决的分歧") || !strings.Contains(out, "missing coverage") {
		t.Fatalf("unexpected summary:\n%s", out)
	}
}

func TestSummarizePreMeeting(t *testing.T) {
	s := meeting.State{
		RoundOrder: []string{"a"},
		Participants: map[string]meeting.ParticipantState{
			"a": {Role: "Architect"},
		},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			0: {"a": {Content: "Security angle"}},
		},
	}
	out := summarizePreMeeting(s)
	if !strings.Contains(out, "Security angle") {
		t.Fatalf("unexpected: %s", out)
	}
}

func TestExtractKeyPoints(t *testing.T) {
 pts := extractKeyPoints("1. First point here\n2. Second point here")
	if len(pts) != 2 {
		t.Fatalf("got %v", pts)
	}
}
