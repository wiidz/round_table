package brief_test

import (
	"testing"

	"round_table/apps/server/internal/adapter/brief"
)

func TestParseDocumentRequiresTitle(t *testing.T) {
	_, err := brief.ParseDocument([]byte("brief:\n  goal: x\n"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseMeetingDocSections(t *testing.T) {
	draft, err := brief.ParseMeetingDoc(`## 会议主题

Topic

## 会议目标

Goal text

## 完成标准

Done when ready
`)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Topic != "Topic" || draft.Brief.Goal != "Goal text" || draft.Brief.DoneCriteria != "Done when ready" {
		t.Fatalf("%+v", draft)
	}
}
