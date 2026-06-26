package discord

import "testing"

func TestParseFreeDialogueQuestion(t *testing.T) {
	tests := []struct {
		in       string
		wantQ    string
		wantOK   bool
	}{
		{"提问 技能数值合理吗", "技能数值合理吗", true},
		{"提问：上线风险有哪些？", "上线风险有哪些？", true},
		{"ask what about pacing?", "what about pacing?", true},
		{"question: scope?", "scope?", true},
		{"提问", "", true},
		{"暂停会议", "", false},
	}
	for _, tc := range tests {
		got, ok := parseFreeDialogueQuestion(tc.in)
		if ok != tc.wantOK || got != tc.wantQ {
			t.Fatalf("parseFreeDialogueQuestion(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.wantQ, tc.wantOK)
		}
	}
}

func TestShouldPostProgress_freeDialogue(t *testing.T) {
	if shouldPostProgress("◆ free dialogue question 1/2 designer → player\n你好") {
		t.Fatal("free dialogue Q&A progress should not post — stream/ack covers content")
	}
	if shouldPostProgress("◆ free dialogue answer 1/2 player → designer\n回答") {
		t.Fatal("free dialogue answer progress should not post")
	}
	start := "▶ free dialogue after round 1 (2 Q&A pairs, max_questions=1/person)"
	if !shouldPostProgress(start) {
		t.Fatal("expected free dialogue start marker to post")
	}
}
