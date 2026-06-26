package discord

import (
	"context"
	"strings"
	"testing"
	"time"

	"round_table/apps/server/internal/domain/meeting"
)

func TestParseMeetArgs(t *testing.T) {
	got, err := parseMeetArgs([]string{"设计", "新职业"}, "decision")
	if err != nil || got.Topic != "设计 新职业" || got.Mode != "decision" {
		t.Fatalf("got=%+v err=%v", got, err)
	}

	got, err = parseMeetArgs([]string{"-mode", "deliberation", "影舞者"}, "decision")
	if err != nil || got.Mode != "deliberation" || got.Topic != "影舞者" {
		t.Fatalf("mode override = %+v err=%v", got, err)
	}

	if _, err := parseMeetArgs(nil, "decision"); err == nil {
		t.Fatal("expected topic required")
	}
}

func TestShouldPostProgress(t *testing.T) {
	if !shouldPostProgress("▶ debate round 2 started") {
		t.Fatal("expected milestone")
	}
	if shouldPostProgress("✓ LLM deliberation participant=designer stance=none elapsed=1s") {
		t.Fatal("expected skip LLM done — usage goes on speaker message")
	}
	if shouldPostProgress("◆ synthesis readiness round=2 ready=false (x)") {
		t.Fatal("expected skip duplicate readiness line")
	}
	if shouldPostProgress("… waiting for principal decision") {
		t.Fatal("expected skip waiting")
	}
	if shouldPostProgress("… LLM deliberation participant=designer turn=(1/2) round=1") {
		t.Fatal("expected skip LLM waiting")
	}
}

func TestFormatStreamForDiscord_participant(t *testing.T) {
	raw := `{"content":"双入口触发","stance":"none","object_reason":""}`
	got := formatStreamForDiscord(raw, LocaleZH)
	if !strings.Contains(got, "双入口触发") {
		t.Fatalf("got=%q", got)
	}
}

func TestFormatStreamForDiscord_synthesis(t *testing.T) {
	raw := `{"core_scheme":["A"],"decisions":["B"],"open_questions":["C?"]}`
	got := formatStreamForDiscord(raw, LocaleZH)
	for _, want := range []string{"方案要点", "A", "已决事项", "B", "开放问题", "C?"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestFormatMeetDone_allOpenQuestions(t *testing.T) {
	qs := []string{
		"失败条件待定",
		"奖励曲线待定",
		"资源平衡待定",
		"兑换边界待定",
		"热修资源待定",
		"账号打通待定",
		"阶段解锁待定",
		"测试资格待定",
	}
	s := meeting.State{
		Status:                 meeting.StatusCompleted,
		Outcome:                "completed",
		MeetingMode:            meeting.MeetingModeDeliberation,
		SynthesisOpenQuestions: qs,
		Consensus:              &meeting.ConsensusState{ResolvedBy: "synthesis"},
	}
	got := formatMeetDone(s, "./data/workspaces", "mtg-test", LocaleZH)
	if strings.Contains(got, "另有") {
		t.Fatalf("should list all questions, got:\n%s", got)
	}
	for i, q := range qs {
		if !strings.Contains(got, q) {
			t.Fatalf("missing question %d: %q", i+1, q)
		}
	}
}

func TestFormatTurnFooter(t *testing.T) {
	got := formatTurnFooter(934, 6680*time.Millisecond, LocaleZH)
	if !strings.Contains(got, "934") || !strings.Contains(got, "Token") {
		t.Fatalf("got=%q", got)
	}
}

func TestChannelStream_CompleteTurn(t *testing.T) {
	pool := &BotPool{
		Default: stubSender{id: "main"},
		byID:    map[string]ChannelSender{"designer": stubSender{id: "designer"}},
	}
	cs := &channelStream{pool: pool, channelID: "ch1"}
	cs.speaker = "designer"
	cs.buf.WriteString(`{"content":"hello","stance":"none"}`)
	cs.End()
	if cs.pending.speaker != "designer" {
		t.Fatal("expected deferred send")
	}
	// simulate send by replacing stub
	designer := &captureSender{}
	pool.byID["designer"] = designer
	cs.CompleteTurn("designer", 100, 2*time.Second)
	if len(designer.messages) != 1 {
		t.Fatalf("messages=%v", designer.messages)
	}
	if !strings.Contains(designer.messages[0], "hello") || !strings.Contains(designer.messages[0], "100 tokens") {
		t.Fatalf("got=%q", designer.messages[0])
	}
}

type captureSender struct {
	messages []string
}

func (c *captureSender) Send(_ context.Context, _, content string) error {
	c.messages = append(c.messages, content)
	return nil
}

func TestParseModeratorSummaryLine(t *testing.T) {
	line := "◆ moderator summary round=2\n## Round 2 研讨摘要\n\n内容"
	round, body, ok := parseModeratorSummaryLine(line)
	if !ok || round != 2 || !strings.Contains(body, "研讨摘要") {
		t.Fatalf("round=%d body=%q ok=%v", round, body, ok)
	}
	if _, _, ok := parseModeratorSummaryLine("◆ moderator summary round=1"); ok {
		t.Fatal("expected no body")
	}
}

func TestMeetSessions(t *testing.T) {
	var s meetSessions
	if err := s.tryStart("ch1", "mtg-1"); err != nil {
		t.Fatal(err)
	}
	if err := s.tryStart("ch1", "mtg-2"); err == nil {
		t.Fatal("expected busy")
	}
	s.clear("ch1")
	if err := s.tryStart("ch1", "mtg-3"); err != nil {
		t.Fatal(err)
	}
}
