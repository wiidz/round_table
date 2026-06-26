package meeting

import (
	"encoding/json"
	"testing"
	"time"

	"round_table/apps/server/internal/domain/event"
)

func TestApply_MeetingCreated_minRoundsBeforeSynthesisDefault(t *testing.T) {
	s := State{ID: "mtg-1", Status: StatusCreated}
	payload, _ := json.Marshal(event.MeetingCreatedPayload{
		Topic:               "topic",
		ConfirmationMode:    ConfirmationModeSkip,
		MaxRoundsPerSegment: 3,
	})
	env := event.Envelope{Type: event.TypeMeetingCreated, Payload: payload, OccurredAt: time.Now()}
	next, err := Apply(s, env)
	if err != nil {
		t.Fatal(err)
	}
	if next.MinRoundsBeforeSynthesis != defaultMinRoundsBeforeSynthesis {
		t.Fatalf("min_rounds = %d, want %d", next.MinRoundsBeforeSynthesis, defaultMinRoundsBeforeSynthesis)
	}
}

func TestApply_MeetingCreated_minRoundsBeforeSynthesisExplicit(t *testing.T) {
	min := 1
	s := State{ID: "mtg-1", Status: StatusCreated}
	payload, _ := json.Marshal(event.MeetingCreatedPayload{
		Topic:                    "topic",
		ConfirmationMode:         ConfirmationModeSkip,
		MaxRoundsPerSegment:      3,
		MinRoundsBeforeSynthesis: &min,
	})
	env := event.Envelope{Type: event.TypeMeetingCreated, Payload: payload, OccurredAt: time.Now()}
	next, err := Apply(s, env)
	if err != nil {
		t.Fatal(err)
	}
	if next.MinRoundsBeforeSynthesis != 1 {
		t.Fatalf("min_rounds = %d, want 1", next.MinRoundsBeforeSynthesis)
	}
}

func TestApply_DeliberationReadinessChecked(t *testing.T) {
	s := State{
		ID:          "mtg-1",
		Status:      StatusRunning,
		MeetingMode: MeetingModeDeliberation,
		CurrentRound: 2,
	}
	payload, _ := json.Marshal(event.DeliberationReadinessCheckedPayload{
		RoundNumber: 2,
		Ready:       true,
		Rationale:   "要素已齐",
	})
	env := event.Envelope{
		Type:       event.TypeDeliberationReadinessChecked,
		Payload:    payload,
		OccurredAt: time.Now(),
		Actor:      event.ActorModerator,
	}
	next, err := Apply(s, env)
	if err != nil {
		t.Fatal(err)
	}
	if next.Status != StatusRunning {
		t.Fatalf("status = %s", next.Status)
	}
}

func TestApply_DeliberationReadinessChecked_requiresDeliberationMode(t *testing.T) {
	s := State{ID: "mtg-1", Status: StatusRunning, MeetingMode: MeetingModeDecision, CurrentRound: 2}
	payload, _ := json.Marshal(event.DeliberationReadinessCheckedPayload{RoundNumber: 2, Ready: false})
	env := event.Envelope{Type: event.TypeDeliberationReadinessChecked, Payload: payload}
	_, err := Apply(s, env)
	if err == nil {
		t.Fatal("expected error for decision mode")
	}
}
