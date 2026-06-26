package engine

import "testing"

func TestIsPreMeetingPhase(t *testing.T) {
	pre := "Topic: x\nPhase: pre-meeting\nPre-meeting (Round 0)\nYou are a\n"
	debate := "Topic: x\nPhase: debate\nRound: 1\nYou are a\n\n--- Discussion so far ---\n## Pre-meeting (Round 0)\n\nviews\n"
	if !isPreMeetingPhase(pre) {
		t.Fatal("expected pre-meeting")
	}
	if isPreMeetingPhase(debate) {
		t.Fatal("debate prompt must not match pre-meeting phase")
	}
}
