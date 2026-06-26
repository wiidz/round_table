package discord

import (
	"context"
	"strings"
	"testing"
	"time"

	prin "round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/domain/meeting"
)

func TestChannelPrincipal_InterventionQueue(t *testing.T) {
	p := NewChannelPrincipal(&BotPool{Default: &captureSender{}}, "zh")
	p.BindMeeting("mtg-1", "ch1", "user-1")
	defer p.UnbindMeeting("mtg-1")

	reply, err := p.DeliverIntervention("ch1", "user-1", "暂停会议")
	if err != nil || !strings.Contains(reply, "暂停") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}

	action, err := p.RunningAction(context.Background(), "mtg-1", meeting.State{Status: meeting.StatusRunning, CurrentRound: 1})
	if err != nil || action.Kind != prin.RunningInterventionPause {
		t.Fatalf("action=%+v err=%v", action, err)
	}

	action, err = p.RunningAction(context.Background(), "mtg-1", meeting.State{CurrentRound: 1})
	if err != nil || action.Kind != "" {
		t.Fatalf("expected empty second poll, got=%+v", action)
	}
}

func TestChannelPrincipal_PausedResume(t *testing.T) {
	p := NewChannelPrincipal(&BotPool{Default: &captureSender{}}, "zh")
	p.BindMeeting("mtg-1", "ch1", "user-1")
	defer p.UnbindMeeting("mtg-1")

	done := make(chan prin.RunningIntervention, 1)
	go func() {
		action, err := p.PausedAction(context.Background(), "mtg-1", meeting.State{Status: meeting.StatusPaused})
		if err != nil {
			t.Errorf("PausedAction: %v", err)
			return
		}
		done <- action
	}()

	for i := 0; i < 50 && !p.PendingPaused("ch1"); i++ {
		time.Sleep(2 * time.Millisecond)
	}
	if !p.PendingPaused("ch1") {
		t.Fatal("expected paused state")
	}

	reply, err := p.DeliverIntervention("ch1", "user-1", "恢复会议")
	if err != nil || reply == "" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}

	select {
	case action := <-done:
		if action.Kind != prin.RunningInterventionResume {
			t.Fatalf("action=%+v", action)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("PausedAction timed out")
	}
}

func TestChannelPrincipal_InterventionBlocksDuringConfirm(t *testing.T) {
	p := NewChannelPrincipal(&BotPool{Default: &captureSender{}}, "zh")
	p.BindMeeting("mtg-1", "ch1", "user-1")
	defer p.UnbindMeeting("mtg-1")
	p.mu.Lock()
	p.sessions["mtg-1"].confirmCycle = 1
	p.mu.Unlock()

	reply, err := p.DeliverIntervention("ch1", "user-1", "暂停会议")
	if err != nil || !strings.Contains(reply, "确认关") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}
