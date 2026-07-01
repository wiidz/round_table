package engine

import (
	"testing"

	"round_table/apps/server/internal/domain/meeting"
)

func TestParseAndPatchMeetingDocStatus(t *testing.T) {
	t.Parallel()
	doc := "# Brief\n\n| 项目 | 内容 |\n|------|------|\n| 会议状态 | 进行中 |\n"
	if got := ParseMeetingDocStatus(doc); got != "进行中" {
		t.Fatalf("parse = %q", got)
	}
	patched := PatchMeetingDocStatus(doc, RenderAbortedMeetingDocStatus())
	if got := ParseMeetingDocStatus(patched); got != "已中断" {
		t.Fatalf("patched = %q", got)
	}
}

func TestIsStaleRunningMeetingDocStatus(t *testing.T) {
	t.Parallel()
	if !IsStaleRunningMeetingDocStatus("进行中") {
		t.Fatal("want stale")
	}
	if IsStaleRunningMeetingDocStatus("已结束") {
		t.Fatal("want not stale")
	}
	if IsStaleRunningMeetingDocStatus("已中断") {
		t.Fatal("want not stale")
	}
}

func TestRenderMeetingStatusLabel_aborted(t *testing.T) {
	t.Parallel()
	label := renderMeetingStatusLabel(meetingStateCompletedAborted())
	if label != "已中断" {
		t.Fatalf("label = %q", label)
	}
}

func meetingStateCompletedAborted() meeting.State {
	s := meeting.State{Status: meeting.StatusCompleted, Outcome: meeting.OutcomeAborted}
	return s
}
