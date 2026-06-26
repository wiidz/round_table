package consensus

import (
	"testing"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

func TestNoObjection_allAgree(t *testing.T) {
	s := meetingWithResponses(map[string]string{"p1": "agree", "p2": "agree"})
	res, err := NoObjection{}.Evaluate(Context{Meeting: s})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Reached {
		t.Fatal("expected consensus")
	}
}

func TestNoObjection_objectBlocks(t *testing.T) {
	s := meetingWithResponses(map[string]string{"p1": "agree", "p2": "object"})
	res, err := NoObjection{}.Evaluate(Context{Meeting: s})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reached {
		t.Fatal("expected no consensus")
	}
}

func TestNoObjection_incompleteRound(t *testing.T) {
	s := meetingWithResponses(map[string]string{"p1": "agree"})
	s.RoundOrder = []string{"p1", "p2"}
	res, err := NoObjection{}.Evaluate(Context{Meeting: s})
	if err != nil || res.Reached {
		t.Fatalf("incomplete round: reached=%v err=%v", res.Reached, err)
	}
}

func meetingWithResponses(stances map[string]string) meeting.State {
	s := meeting.NewState("m1")
	s.CurrentRound = 1
	s.RoundOrder = make([]string, 0, len(stances))
	s.RoundResponses[1] = make(map[string]meeting.RoundResponse)
	for id, st := range stances {
		s.Participants[id] = meeting.ParticipantState{ID: id, Role: "Expert"}
		s.RoundOrder = append(s.RoundOrder, id)
		s.RoundResponses[1][id] = meeting.RoundResponse{Content: "x", Stance: event.Stance(st)}
	}
	return s
}
