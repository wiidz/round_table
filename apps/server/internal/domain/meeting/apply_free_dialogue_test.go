package meeting

import (
	"encoding/json"
	"testing"

	"round_table/apps/server/internal/domain/event"
)

func TestApplyFreeDialogueQuestionAsked_principalMediated(t *testing.T) {
	s := State{
		ID:               "mtg-1",
		Status:           StatusRunning,
		InFreeDialogue:   true,
		ParticipantOrder: []string{"designer", "player"},
		Participants: map[string]ParticipantState{
			"designer": {ID: "designer", Role: "design"},
			"player":   {ID: "player", Role: "player"},
		},
		FreeDialogueQuestionIndex: 0,
	}
	payload, _ := json.Marshal(event.FreeDialogueQuestionAskedPayload{
		AskerID:           PrincipalRelayAskerID,
		AnswererID:        "designer",
		QuestionIndex:     0,
		Content:           "Principal 问题",
		PrincipalMediated: true,
	})
	env := event.Envelope{Type: event.TypeFreeDialogueQuestion, Payload: payload}
	ns, err := Apply(s, env)
	if err != nil {
		t.Fatal(err)
	}
	if ns.PendingFreeDialogue == nil || !ns.PendingFreeDialogue.PrincipalMediated {
		t.Fatalf("pending = %+v", ns.PendingFreeDialogue)
	}
}
