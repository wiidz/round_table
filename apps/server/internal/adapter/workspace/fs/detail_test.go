package fs

import (
	"os"
	"path/filepath"
	"testing"

	"round_table/apps/server/internal/adapter/workspace"
)

func TestReadMeetingDetail(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	if err := s.EnsureMeeting("mtg-a", "topic"); err != nil {
		t.Fatal(err)
	}

	meetingDoc := filepath.Join(dir, "mtg-a", "MEETING.md")
	body := `# 会议简报 · Meeting Brief

| 项目 | 内容 |
|------|------|
| 会议编号 | mtg-a |
| 会议时间 | 2026-06-27 19:23 (CST) |
| 会议状态 | 已结束 |
| 会议模式 | 裁决型（decision） |

## 会议主题

Auth Service 拆分
`
	if err := os.WriteFile(meetingDoc, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mtg-a", "MINUTES.md"), []byte("# Minutes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "mtg-a", "artifacts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mtg-a", "artifacts", "minutes.md"), []byte("# Conclusion\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	detail, err := s.ReadMeetingDetail("mtg-a")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Topic != "Auth Service 拆分" {
		t.Fatalf("topic=%q", detail.Topic)
	}
	if detail.Status != "已结束" {
		t.Fatalf("status=%q", detail.Status)
	}
	if detail.Mode == "" {
		t.Fatal("mode empty")
	}
	if detail.StartedAt == "" {
		t.Fatal("started_at empty")
	}
	if len(detail.Files) != 3 {
		t.Fatalf("files=%v", detail.Files)
	}
}

func TestReadMeetingDetailNotFound(t *testing.T) {
	s := NewStore(t.TempDir())
	_, err := s.ReadMeetingDetail("missing")
	if err != workspace.ErrNotFound {
		t.Fatalf("err=%v", err)
	}
}
