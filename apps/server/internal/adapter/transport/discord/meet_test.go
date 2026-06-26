package discord

import (
	"strings"
	"testing"
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
	if !shouldPostProgress("✓ LLM deliberation participant=designer stance=none elapsed=1s") {
		t.Fatal("expected LLM done line")
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
	got := formatStreamForDiscord(raw)
	if !strings.Contains(got, "双入口触发") {
		t.Fatalf("got=%q", got)
	}
}

func TestFormatStreamForDiscord_synthesis(t *testing.T) {
	raw := `{"core_scheme":["A"],"decisions":["B"],"open_questions":["C?"]}`
	got := formatStreamForDiscord(raw)
	for _, want := range []string{"方案要点", "A", "已决事项", "B", "开放问题", "C?"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
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
