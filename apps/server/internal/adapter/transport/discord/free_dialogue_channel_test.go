package discord

import (
	"context"
	"testing"

	"round_table/apps/server/internal/domain/meeting"
)

func TestChannelPrincipalFreeDialogueQuestion(t *testing.T) {
	p := NewChannelPrincipal(nil, "zh")
	p.BindMeeting("mtg-1", "ch1", "user-1")
	p.MarkFreeDialogue("ch1", true)

	reply, err := p.DeliverFreeDialogueQuestion("ch1", "user-1", "提问 数值怎么定？")
	if err != nil {
		t.Fatal(err)
	}
	if reply == "" {
		t.Fatal("expected ack reply")
	}

	q, ok, err := p.FreeDialogueQuestion(context.Background(), "mtg-1", meeting.State{})
	if err != nil || !ok || q != "数值怎么定？" {
		t.Fatalf("FreeDialogueQuestion = (%q, %v, %v)", q, ok, err)
	}

	q2, ok2, _ := p.FreeDialogueQuestion(context.Background(), "mtg-1", meeting.State{})
	if ok2 || q2 != "" {
		t.Fatalf("expected empty queue, got (%q, %v)", q2, ok2)
	}
}

func TestChannelPrincipalFreeDialogueQuestion_wrongPhase(t *testing.T) {
	p := NewChannelPrincipal(nil, "zh")
	p.BindMeeting("mtg-1", "ch1", "user-1")

	reply, err := p.DeliverFreeDialogueQuestion("ch1", "user-1", "提问 test")
	if err != nil {
		t.Fatal(err)
	}
	if reply == "" {
		t.Fatal("expected wrong-phase reply")
	}
}
