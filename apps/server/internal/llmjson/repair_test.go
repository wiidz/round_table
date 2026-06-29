package llmjson_test

import (
	"testing"

	"round_table/apps/server/internal/llmjson"
)

func TestRepairObject_missingBrace(t *testing.T) {
	raw := `{"ready": true, "rationale": "ok", "gaps": []`
	got := llmjson.RepairObject(raw)
	if got[len(got)-1] != '}' {
		t.Fatalf("got=%q", got)
	}
}

func TestClean_stripsFence(t *testing.T) {
	raw := "```json\n{\"a\":1}\n```"
	if got := llmjson.Clean(raw); got != `{"a":1}` {
		t.Fatalf("got=%q", got)
	}
}
