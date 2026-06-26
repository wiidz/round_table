package meeting

import (
	"encoding/json"
	"testing"

	"round_table/apps/server/internal/domain/event"
)

func TestApplyConfirmationRejected_itemNotes(t *testing.T) {
	s := State{
		ID:     "mtg-1",
		Status: StatusConfirmation,
		Confirmation: &ConfirmationState{
			Cycle: 1,
		},
		MaxRoundsPerSegment: 5,
		CurrentRound:        2,
	}
	payload, _ := json.Marshal(event.ConfirmationRejectedPayload{
		Cycle:     1,
		ItemNotes: map[int]string{2: "技能树需重算"},
	})
	env := event.Envelope{Type: event.TypeConfirmationRejected, Payload: payload}
	ns, err := Apply(s, env)
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(ns.PrincipalFeedback, "Item 2", "技能树需重算") {
		t.Fatalf("feedback=%q", ns.PrincipalFeedback)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if len(p) > 0 && !containsSubstring(s, p) {
			return false
		}
	}
	return true
}

func containsSubstring(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexSubstring(s, sub) >= 0)
}

func indexSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
