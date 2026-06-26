package discord

import (
	"testing"

	prin "round_table/apps/server/internal/adapter/principal"
)

func TestParseIntervention(t *testing.T) {
	cases := []struct {
		in   string
		kind prin.RunningInterventionKind
	}{
		{"暂停会议", prin.RunningInterventionPause},
		{"恢复会议", prin.RunningInterventionResume},
		{"终止会议", prin.RunningInterventionAbort},
		{"中止会议 方案不成熟", prin.RunningInterventionAbort},
		{"立即合成", prin.RunningInterventionForceSynthesis},
		{"强制共识", prin.RunningInterventionForceConsensus},
		{"pause", prin.RunningInterventionPause},
	}
	for _, tc := range cases {
		got, ok := parseIntervention(tc.in)
		if !ok || got.Kind != tc.kind {
			t.Fatalf("in=%q got=%+v ok=%v want=%s", tc.in, got, ok, tc.kind)
		}
	}
	if _, ok := parseIntervention("hello"); ok {
		t.Fatal("expected no match")
	}
}

func TestExcerptArtifact(t *testing.T) {
	long := string(make([]rune, 2000))
	got := excerptArtifact(long, 100, LocaleZH)
	if len([]rune(got)) > 200 {
		t.Fatalf("expected excerpt, len=%d", len([]rune(got)))
	}
	if !contains(got, "节选") {
		t.Fatalf("got=%q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexSubstring(s, sub) >= 0)
}

func indexSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
