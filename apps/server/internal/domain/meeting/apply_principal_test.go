package meeting

import (
	"encoding/json"
	"testing"
	"time"

	"round_table/apps/server/internal/domain/event"
)

func TestApply_SynthesisForced(t *testing.T) {
	s := State{
		ID:           "mtg-1",
		Status:       StatusRunning,
		MeetingMode:  MeetingModeDeliberation,
		CurrentRound: 2,
	}
	payload, _ := json.Marshal(event.SynthesisForcedPayload{Reason: "够了"})
	env := event.Envelope{
		Type:       event.TypeSynthesisForced,
		Payload:    payload,
		OccurredAt: time.Now(),
		Actor:      event.ActorPrincipal,
	}
	next, err := Apply(s, env)
	if err != nil {
		t.Fatal(err)
	}
	if next.Status != StatusRunning {
		t.Fatalf("status = %s, want Running until SynthesisCompleted", next.Status)
	}
}

func TestApply_SynthesisForced_requiresDeliberation(t *testing.T) {
	s := State{ID: "mtg-1", Status: StatusRunning, MeetingMode: MeetingModeDecision, CurrentRound: 2}
	payload, _ := json.Marshal(event.SynthesisForcedPayload{Reason: "x"})
	env := event.Envelope{Type: event.TypeSynthesisForced, Payload: payload}
	_, err := Apply(s, env)
	if err == nil {
		t.Fatal("expected error for decision mode")
	}
}

func TestApply_ConsensusForced_rejectsDeliberation(t *testing.T) {
	s := State{ID: "mtg-1", Status: StatusRunning, MeetingMode: MeetingModeDeliberation, CurrentRound: 2}
	payload, _ := json.Marshal(event.ConsensusForcedPayload{Reason: "x"})
	env := event.Envelope{Type: event.TypeConsensusForced, Payload: payload}
	_, err := Apply(s, env)
	if err == nil {
		t.Fatal("expected error for deliberation mode")
	}
}

func TestApply_ConsensusForced_decisionMode(t *testing.T) {
	s := State{
		ID:                "mtg-1",
		Status:            StatusRunning,
		MeetingMode:       MeetingModeDecision,
		ConsensusStrategy: "no_objection",
		CurrentRound:      2,
	}
	payload, _ := json.Marshal(event.ConsensusForcedPayload{Reason: "Principal 决定"})
	env := event.Envelope{
		Type:       event.TypeConsensusForced,
		Payload:    payload,
		OccurredAt: time.Now(),
		Actor:      event.ActorPrincipal,
	}
	next, err := Apply(s, env)
	if err != nil {
		t.Fatal(err)
	}
	if next.Status != StatusConsensus {
		t.Fatalf("status = %s", next.Status)
	}
	if next.Consensus == nil || next.Consensus.ResolvedBy != "principal" {
		t.Fatalf("consensus = %+v", next.Consensus)
	}
}
