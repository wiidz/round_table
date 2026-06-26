package engine

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

func TestFormatDiscussionContext_withPreMeetingAndModerator(t *testing.T) {
	s := meeting.State{
		CurrentRound:        2,
		PreMeetingSummary:   "Pre-meeting perspectives\n\n- **skeptic**: JWT risk\n",
		PreMeetingCompleted: true,
		ModeratorSummaries: map[int]string{
			1: "Moderator summary — Round 1\n\nNo objections.",
		},
		RoundOrder: []string{"skeptic", "pragmatist"},
		Participants: map[string]meeting.ParticipantState{
			"skeptic":    {Role: "Security"},
			"pragmatist": {Role: "Tech Lead"},
		},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {
				"skeptic": {Content: "Still missing audit", Stance: event.StanceObject},
			},
		},
	}

	ctx := formatDiscussionContext(s, "pragmatist")
	if !strings.Contains(ctx, "Pre-meeting") || !strings.Contains(ctx, "JWT risk") {
		t.Fatalf("missing pre-meeting:\n%s", ctx)
	}
	if !strings.Contains(ctx, "Moderator summary after Round 1") {
		t.Fatalf("missing moderator summary:\n%s", ctx)
	}
	if !strings.Contains(ctx, "Round 2 (in progress)") || !strings.Contains(ctx, "audit") {
		t.Fatalf("missing in-progress:\n%s", ctx)
	}
}

func TestFormatDiscussionContext_preMeetingRound(t *testing.T) {
	if ctx := formatDiscussionContext(meeting.State{CurrentRound: 0}, "a"); ctx != "" {
		t.Fatalf("round 0 should have empty context, got %q", ctx)
	}
}

func TestFormatDiscussionContext_firstSpeaker(t *testing.T) {
	s := meeting.State{
		CurrentRound:      1,
		PreMeetingSummary: "views",
		PreMeetingCompleted: true,
		RoundOrder:        []string{"a", "b"},
		Participants: map[string]meeting.ParticipantState{
			"a": {Role: "A"},
			"b": {Role: "B"},
		},
	}
	ctx := formatDiscussionContext(s, "a")
	if !strings.Contains(ctx, "views") {
		t.Fatalf("expected pre-meeting only:\n%s", ctx)
	}
	if strings.Contains(ctx, "in progress") {
		t.Fatalf("first speaker should not see in-progress:\n%s", ctx)
	}
}
