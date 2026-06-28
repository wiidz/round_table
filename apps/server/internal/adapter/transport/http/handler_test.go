package httptransport

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"round_table/apps/server/internal/platform/config"
	wsfs "round_table/apps/server/internal/adapter/workspace/fs"
)

func TestHandleListMeetings(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 3; i++ {
		id := "mtg-" + string(rune('0'+i))
		if err := os.MkdirAll(filepath.Join(dir, id), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	h, err := NewHandler(config.Config{Workspace: config.Workspace{Root: dir}}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/meetings?page=1&page_size=2", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"total":3`) || !strings.Contains(rec.Body.String(), `"page_size":2`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestHandleGetMeeting(t *testing.T) {
	dir := t.TempDir()
	s := wsfs.NewStore(dir)
	if err := s.EnsureMeeting("mtg-a", "topic"); err != nil {
		t.Fatal(err)
	}
	meetingDoc := filepath.Join(dir, "mtg-a", "MEETING.md")
	body := `# 会议简报

| 项目 | 内容 |
|------|------|
| 会议状态 | 已结束 |

## 会议主题

Test Topic
`
	if err := os.WriteFile(meetingDoc, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	h, err := NewHandler(config.Config{Workspace: config.Workspace{Root: dir}}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/meetings/mtg-a", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Test Topic") {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestHandleListPrincipals(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	if err := os.MkdirAll(filepath.Join(templates, "principals"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templates, "principals", "USER.md"), []byte("# user"), 0o644); err != nil {
		t.Fatal(err)
	}

	h, err := NewHandler(config.Config{
		Profile: config.Profile{Root: dir, Templates: templates},
	}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	store := h.profile
	if err := store.EnsurePrincipal("discord:test"); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	h.Register(mux)
	req := httptest.NewRequest(http.MethodGet, "/api/principals", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "discord:test") {
		t.Fatalf("body=%s", rec.Body.String())
	}
}
