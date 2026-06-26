package engine

import (
	"testing"

	"round_table/apps/server/internal/domain/meeting"
)

func TestRenderTokenUsageSummary(t *testing.T) {
	s := meeting.State{
		TokenUsageLog: []meeting.TokenUsageRecord{
			{Turn: 1, Phase: "pre-meeting", ParticipantID: "a", Model: "deepseek-chat", RoundNumber: 0, PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
			{Turn: 2, Phase: "debate", ParticipantID: "b", Model: "deepseek-chat", RoundNumber: 1, PromptTokens: 200, CompletionTokens: 80, TotalTokens: 280},
		},
		TokenUsageTotals: meeting.TokenUsageTotals{
			CallCount: 2, PromptTokens: 300, CompletionTokens: 130, TotalTokens: 430,
		},
	}
	out := renderTokenUsageSummary(s)
	if !containsAll(out, "430", "pre-meeting", "debate", "a", "b") {
		t.Fatalf("summary missing expected content:\n%s", out)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (s == sub || len(s) > 0 && findSub(s, sub)))
}

func findSub(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
