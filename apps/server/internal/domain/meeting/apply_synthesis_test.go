package meeting

import (
	"encoding/json"
	"testing"
	"time"

	"round_table/apps/server/internal/domain/event"
)

func TestApply_SynthesisCompleted(t *testing.T) {
	s := State{
		ID:          "mtg-1",
		Status:      StatusRunning,
		MeetingMode: MeetingModeDeliberation,
	}
	payload, _ := json.Marshal(event.SynthesisCompletedPayload{
		Summary:       "# Draft\n\ncontent",
		OpenQuestions: []string{"平衡性待验证"},
		ResolvedBy:    "max_rounds",
	})
	env := event.Envelope{
		Type:       event.TypeSynthesisCompleted,
		Payload:    payload,
		OccurredAt: time.Now(),
		Actor:      event.ActorModerator,
	}
	next, err := Apply(s, env)
	if err != nil {
		t.Fatal(err)
	}
	if next.Status != StatusConsensus {
		t.Fatalf("status = %s", next.Status)
	}
	if next.SynthesisSummary == "" {
		t.Fatal("expected synthesis summary")
	}
	if next.Consensus == nil || next.Consensus.Strategy != MeetingModeDeliberation {
		t.Fatalf("consensus = %+v", next.Consensus)
	}
}

func TestApply_SynthesisCompleted_agendaSections(t *testing.T) {
	s := State{
		ID:          "mtg-1",
		Status:      StatusRunning,
		MeetingMode: MeetingModeDeliberation,
	}
	payload, _ := json.Marshal(event.SynthesisCompletedPayload{
		Summary: "# Draft",
		Sections: []event.SynthesisAgendaSectionPayload{
			{AgendaID: "skills", Summary: []string{"连击"}},
		},
		CrossCutting: &event.SynthesisCrossCuttingPayload{
			OpenQuestions: []string{"平衡？"},
		},
	})
	env := event.Envelope{Type: event.TypeSynthesisCompleted, Payload: payload}
	next, err := Apply(s, env)
	if err != nil {
		t.Fatal(err)
	}
	if len(next.SynthesisSections) != 1 || next.SynthesisSections[0].AgendaID != "skills" {
		t.Fatalf("sections = %+v", next.SynthesisSections)
	}
	if next.SynthesisCrossCutting == nil || len(next.SynthesisCrossCutting.OpenQuestions) != 1 {
		t.Fatalf("cross = %+v", next.SynthesisCrossCutting)
	}
}

func TestApply_SynthesisCompleted_requiresDeliberationMode(t *testing.T) {
	s := State{ID: "mtg-1", Status: StatusRunning, MeetingMode: MeetingModeDecision}
	payload, _ := json.Marshal(event.SynthesisCompletedPayload{Summary: "x"})
	env := event.Envelope{Type: event.TypeSynthesisCompleted, Payload: payload}
	_, err := Apply(s, env)
	if err == nil {
		t.Fatal("expected error for decision mode")
	}
}

func TestDefaultMeetingGoal_deliberation(t *testing.T) {
	got := defaultMeetingGoal("职业设计", MeetingModeDeliberation)
	if got == "" || got == defaultMeetingGoal("职业设计", MeetingModeDecision) {
		t.Fatalf("goal = %q", got)
	}
}
