package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListMeetings(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	if err := s.EnsureMeeting("mtg-a", "topic a"); err != nil {
		t.Fatal(err)
	}
	if err := s.EnsureMeeting("mtg-b", "topic b"); err != nil {
		t.Fatal(err)
	}

	meetingDoc := filepath.Join(dir, "mtg-a", "MEETING.md")
	body := `# 会议简报 · Meeting Brief

| 项目 | 内容 |
|------|------|
| 会议编号 | mtg-a |
| 会议状态 | 已结束 |

## 会议主题

Auth Service 拆分
`
	if err := os.WriteFile(meetingDoc, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListMeetings()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("len=%d want 2", len(list))
	}
	if list[0].ID != "mtg-a" && list[1].ID != "mtg-a" {
		t.Fatalf("expected mtg-a first by mtime, got %#v", list)
	}

	var found *struct {
		topic, status string
	}
	for i := range list {
		if list[i].ID == "mtg-a" {
			found = &struct{ topic, status string }{list[i].Topic, list[i].Status}
			break
		}
	}
	if found == nil {
		t.Fatal("mtg-a missing")
	}
	if found.topic != "Auth Service 拆分" {
		t.Fatalf("topic=%q", found.topic)
	}
	if found.status != "已结束" {
		t.Fatalf("status=%q", found.status)
	}
}

func TestListMeetingsEmptyRoot(t *testing.T) {
	s := NewStore(t.TempDir())
	list, err := s.ListMeetings()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("len=%d", len(list))
	}
}

func TestListMeetingsPage(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	for i := 1; i <= 12; i++ {
		id := "mtg-" + string(rune('0'+i))
		if err := s.EnsureMeeting(id, ""); err != nil {
			t.Fatal(err)
		}
	}

	page1, err := s.ListMeetingsPage(1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if page1.Total != 12 || len(page1.Meetings) != 10 || page1.Page != 1 {
		t.Fatalf("page1=%+v", page1)
	}

	page2, err := s.ListMeetingsPage(2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Meetings) != 2 {
		t.Fatalf("page2 len=%d", len(page2.Meetings))
	}
}
