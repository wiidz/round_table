package discord

import (
	"context"
	"strings"
	"testing"
	"time"

	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/stream"
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

func TestFormatStreamForDiscord_malformedParticipantJSON(t *testing.T) {
	raw := `{"content":"玩家代表提到「流派多样性」，我想追问：你会不会觉得不便？」`
	got := formatStreamForDiscord(raw, LocaleZH)
	if strings.HasPrefix(got, "{") || got == "" {
		t.Fatalf("expected parsed content, got=%q", got)
	}
	if !strings.Contains(got, "流派多样性") {
		t.Fatalf("got=%q", got)
	}
}

func TestFallbackStreamBody_neverLeaksRawJSON(t *testing.T) {
	raw := `{"content":"broken`
	got := fallbackStreamBody(raw, LocaleZH)
	if strings.HasPrefix(got, "{") {
		t.Fatalf("should not leak raw JSON: %q", got)
	}
}

func TestFormatStreamForDiscord_synthesis(t *testing.T) {
	raw := `{"executive_verdict":"建议采用方案 A","key_decisions":["统一冷却"],"core_scheme":["A"],"decisions":["B"],"open_questions":["C?"]}`
	got := formatStreamForDiscord(raw, LocaleZH)
	for _, want := range []string{"方案 A", "Principal 需知", "统一冷却", "方案要点", "A", "已决事项", "B", "开放问题", "C?"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestFormatStreamForDiscord_executiveRecapMarkdown(t *testing.T) {
	raw := "## 会议回顾\n\n### 目标与议程覆盖\n已覆盖核心模块。"
	got := formatStreamForDiscord(raw, LocaleZH)
	if !strings.Contains(got, "会议回顾") || !strings.Contains(got, "目标与议程覆盖") {
		t.Fatalf("got=%q", got)
	}
	if strings.Count(got, "## 会议回顾") > 0 {
		t.Fatalf("duplicate heading in %q", got)
	}
}

func TestFallbackStreamBody_malformedSynthesisJSON(t *testing.T) {
	raw := `{"executive_verdict":"broken`
	got := fallbackStreamBody(raw, LocaleZH)
	if strings.HasPrefix(got, "{") {
		t.Fatalf("should not leak JSON: %q", got)
	}
}

func TestChannelStream_suppressExecutiveRecapStream(t *testing.T) {
	capture := &captureSender{}
	pool := &BotPool{Default: capture, byID: map[string]ChannelSender{"moderator": capture}}
	cs := &channelStream{pool: pool, channelID: "ch1", loc: LocaleZH}
	cs.Start(stream.Meta{ParticipantID: "moderator", Phase: "moderator-executive-recap"})
	cs.buf.WriteString("## 会议回顾\n\n内容")
	cs.End()
	if len(capture.messages) != 0 {
		t.Fatalf("stream should be suppressed, got %v", capture.messages)
	}
}

func TestFormatMeetDone_noInlineOpenQuestions(t *testing.T) {
	s := meeting.State{
		Status:                 meeting.StatusCompleted,
		Outcome:                "completed",
		MeetingMode:            meeting.MeetingModeDeliberation,
		SynthesisOpenQuestions: []string{"失败条件待定", "奖励曲线待定"},
		Consensus:              &meeting.ConsensusState{ResolvedBy: "synthesis"},
	}
	got := formatMeetDone(s, "./data/workspaces", "mtg-test", LocaleZH)
	if strings.Contains(got, "失败条件待定") {
		t.Fatalf("open questions should be in artifact push, not summary:\n%s", got)
	}
	if !strings.Contains(got, "会议结束") {
		t.Fatalf("got=%q", got)
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

func TestParseExecutiveRecapLine(t *testing.T) {
	line := "◆ executive recap\n## 会议回顾\n\n内容"
	body, ok := parseExecutiveRecapLine(line)
	if !ok || !strings.Contains(body, "会议回顾") {
		t.Fatalf("body=%q ok=%v", body, ok)
	}
}

func TestChannelProgress_postsModeratorSummary(t *testing.T) {
	capture := &captureSender{}
	pool := &BotPool{Default: capture}
	cp := &channelProgress{pool: pool, channelID: "ch1", loc: LocaleZH}
	cp.Logf("◆ moderator summary round=%d\n%s", 2, "## Round 2 研讨摘要\n\n### 本轮进展\n共识增加")
	if len(capture.messages) == 0 || !strings.Contains(capture.messages[0], "第 2 轮摘要") {
		t.Fatalf("messages=%v", capture.messages)
	}
}

func TestChannelProgress_postsExecutiveRecap(t *testing.T) {
	capture := &captureSender{}
	pool := &BotPool{Default: capture}
	cp := &channelProgress{pool: pool, channelID: "ch1", loc: LocaleZH}
	cp.Logf("◆ executive recap\n## 会议回顾\n\n过程脉络")
	if len(capture.messages) == 0 || !strings.Contains(capture.messages[0], "会议回顾") {
		t.Fatalf("messages=%v", capture.messages)
	}
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
