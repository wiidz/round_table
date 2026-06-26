package discord

import "testing"

func TestMeetStartTriggers(t *testing.T) {
	for _, s := range []string{"开始会议", "新会议", "会议开始"} {
		if !isMeetStartTrigger(s) {
			t.Fatalf("expected trigger: %q", s)
		}
	}
	for _, s := range []string{"!rt meet", "新会议 测试", "开始", ""} {
		if isMeetStartTrigger(s) {
			t.Fatalf("unexpected trigger: %q", s)
		}
	}
}

func TestMeetCancelTriggers(t *testing.T) {
	if !isMeetCancelTrigger("取消会议") {
		t.Fatal("expected cancel trigger")
	}
	if isMeetCancelTrigger("取消") {
		t.Fatal("partial match should not cancel")
	}
}
