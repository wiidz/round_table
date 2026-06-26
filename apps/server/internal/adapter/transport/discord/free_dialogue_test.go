package discord

import "testing"

func TestParseFreeDialogueQuestion(t *testing.T) {
	tests := []struct {
		in         string
		wantQ      string
		wantTarget string
		wantOK     bool
	}{
		{"提问 技能数值合理吗", "技能数值合理吗", "", true},
		{"提问 designer 数值怎么定", "数值怎么定", "designer", true},
		{"ask player: scope?", "scope?", "player", true},
		{"question: scope?", "scope?", "", true},
		{"提问", "", "", true},
		{"暂停会议", "", "", false},
	}
	for _, tc := range tests {
		q, target, ok := parseFreeDialogueQuestion(tc.in)
		if ok != tc.wantOK || q != tc.wantQ || target != tc.wantTarget {
			t.Fatalf("parseFreeDialogueQuestion(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tc.in, q, target, ok, tc.wantQ, tc.wantTarget, tc.wantOK)
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
	turn := "▶ free dialogue turn 1/2 answerer=designer"
	if !shouldPostProgress(turn) {
		t.Fatal("expected free dialogue turn marker to post")
	}
}

func TestParseConfirmationItemNotes(t *testing.T) {
	notes := parseConfirmationItemNotes("2: 技能树需重算")
	if notes[2] != "技能树需重算" {
		t.Fatalf("notes=%v", notes)
	}
}
