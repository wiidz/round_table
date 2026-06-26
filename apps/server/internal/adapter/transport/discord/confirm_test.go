package discord

import (
	"context"
	"strings"
	"testing"
	"time"

	prin "round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/domain/event"
)

func TestParseConfirmationLimitReply(t *testing.T) {
	cases := []struct {
		in       string
		decision prin.Decision
		feedback string
	}{
		{"1", prin.DecisionLimitForceApprove, ""},
		{"2", prin.DecisionLimitContinue, ""},
		{"3", prin.DecisionLimitAbort, ""},
		{"2 技能树需重算", prin.DecisionLimitContinue, "技能树需重算"},
		{"强制批准", prin.DecisionLimitForceApprove, ""},
	}
	for _, tc := range cases {
		got, err := parseConfirmationLimitReply(tc.in)
		if err != nil {
			t.Fatalf("in=%q err=%v", tc.in, err)
		}
		if got.Decision != tc.decision || got.Feedback != tc.feedback {
			t.Fatalf("in=%q got=%+v", tc.in, got)
		}
	}
}

func TestFormatConfirmationLimitFallback_zh(t *testing.T) {
	got := formatConfirmationLimitFallback(LocaleZH, "mtg-1", 3, event.ConfirmationBrief{
		LimitFallback:       true,
		LimitRejectFeedback: "还需改数值",
	})
	for _, want := range []string{"确认关已达上限", "强制批准", "继续研讨", "中止会议", "**1**", "**2**", "**3**"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestParseConfirmationReply(t *testing.T) {
	cases := []struct {
		in       string
		decision prin.Decision
		feedback string
	}{
		{"批准", prin.DecisionApproved, ""},
		{"1", prin.DecisionApproved, ""},
		{"驳回", prin.DecisionRejected, ""},
		{"驳回 技能数值需重算", prin.DecisionRejected, "技能数值需重算"},
		{"2", prin.DecisionRejected, ""},
		{"reject need more detail", prin.DecisionRejected, "need more detail"},
	}
	for _, tc := range cases {
		got, err := parseConfirmationReply(tc.in)
		if err != nil {
			t.Fatalf("in=%q err=%v", tc.in, err)
		}
		if got.Decision != tc.decision || got.Feedback != tc.feedback {
			t.Fatalf("in=%q got=%+v want decision=%s feedback=%q", tc.in, got, tc.decision, tc.feedback)
		}
	}
}

func TestFormatConfirmationBrief_zh(t *testing.T) {
	got := formatConfirmationBrief(LocaleZH, "mtg-1", 1, event.ConfirmationBrief{
		ExecutiveSummary: "主题 — 请审阅",
		Items: []event.ConfirmationItem{{
			Index: 1, Title: "方案草案", Description: "核心技能…",
		}},
	})
	for _, want := range []string{"Principal 确认关", "方案草案", "批准", "驳回", "mtg-1"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestChannelPrincipal_ConfirmAndDeliver(t *testing.T) {
	sender := &captureSender{}
	pool := &BotPool{Default: sender}
	p := NewChannelPrincipal(pool, "zh")
	p.BindMeeting("mtg-1", "ch1", "user-1")
	defer p.UnbindMeeting("mtg-1")

	brief := event.ConfirmationBrief{
		ExecutiveSummary: "测试",
		Items:            []event.ConfirmationItem{{Index: 1, Title: "结论", Description: "ok"}},
	}

	done := make(chan prin.Response, 1)
	errCh := make(chan error, 1)
	go func() {
		resp, err := p.Confirm(context.Background(), "mtg-1", brief, 1)
		if err != nil {
			errCh <- err
			return
		}
		done <- resp
	}()

	for i := 0; i < 50 && len(sender.messages) == 0; i++ {
		time.Sleep(2 * time.Millisecond)
	}
	if len(sender.messages) == 0 {
		t.Fatal("expected confirmation brief posted")
	}
	if !strings.Contains(sender.messages[0], "确认关") {
		t.Fatalf("brief=%q", sender.messages[0])
	}

	reply, err := p.DeliverConfirmationReply("ch1", "user-1", "批准")
	if err != nil || reply == "" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}

	select {
	case err := <-errCh:
		t.Fatalf("Confirm: %v", err)
	case resp := <-done:
		if resp.Decision != prin.DecisionApproved {
			t.Fatalf("decision=%s", resp.Decision)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Confirm timed out")
	}
}

func TestChannelPrincipal_wrongAuthor(t *testing.T) {
	p := NewChannelPrincipal(&BotPool{Default: &captureSender{}}, "zh")
	p.BindMeeting("mtg-1", "ch1", "user-1")
	p.mu.Lock()
	p.sessions["mtg-1"].confirmCycle = 1
	p.mu.Unlock()

	reply, err := p.DeliverConfirmationReply("ch1", "user-2", "批准")
	if err != nil || !strings.Contains(reply, "Principal") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}
