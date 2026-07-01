package httptransport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/storage"
	"round_table/apps/server/internal/adapter/storage/memory"
	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/engine"
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

	h, err := NewHandler(config.Config{Workspace: config.Workspace{Root: dir}}, nil, nil, nil, nil)
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

type stubMeetingCatalog struct {
	total int
}

func (s stubMeetingCatalog) ListMeetingsPage(_ context.Context, _, _ int) (workspace.PaginatedMeetings, error) {
	return workspace.PaginatedMeetings{
		Meetings: []workspace.MeetingIndex{{ID: "mtg-index-only", Topic: "indexed"}},
		Total:    s.total,
		Page:     1,
		PageSize: 10,
	}, nil
}

func TestHandleListMeetingsPrefersWorkspaceOverCatalog(t *testing.T) {
	dir := t.TempDir()
	for _, id := range []string{"mtg-a", "mtg-b", "mtg-c"} {
		if err := os.MkdirAll(filepath.Join(dir, id), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	h, err := NewHandler(config.Config{Workspace: config.Workspace{Root: dir}}, stubMeetingCatalog{total: 1}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/meetings?page=1&page_size=10", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"total":3`) {
		t.Fatalf("expected workspace scan total=3, body=%s", rec.Body.String())
	}
}

var _ storage.MeetingCatalog = stubMeetingCatalog{}

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
| 辩论轮次上限 | 2（不含 Pre-meeting Round 0） |

## 会议主题

Test Topic

## 参会人员

| 参会者 | 角色 | 专长 | 参会目标 |
|--------|------|------|----------|
| a | a | x | — |
`
	if err := os.WriteFile(meetingDoc, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	h, err := NewHandler(config.Config{Workspace: config.Workspace{Root: dir}}, nil, nil, nil, nil)
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
	if !strings.Contains(rec.Body.String(), `"participant_count"`) {
		t.Fatalf("missing participant_count in body=%s", rec.Body.String())
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
	}, nil, nil, nil, nil)
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

func TestHandlePostMeetingAbort(t *testing.T) {
	dir := t.TempDir()
	store := memory.New()
	ws := wsfs.NewStore(dir)
	ctx := context.Background()

	eng := engine.New(store, consensus.NoObjection{}, nil, nil, ws, nil, nil)
	freeQ := 0
	meetingID := "mtg-abort-api"
	if _, err := eng.CreateMeeting(ctx, engine.CreateMeetingInput{
		MeetingID:                meetingID,
		Topic:                    "abort api",
		ConfirmationMode:         meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment:      1,
		FreeDialogueMaxQuestions: &freeQ,
		Participants: []engine.ParticipantInput{
			{ID: "a", Role: "Architect"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	h, err := NewHandler(config.Config{Workspace: config.Workspace{Root: dir}}, nil, store, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/meetings/"+meetingID+"/abort", strings.NewReader(`{"reason":"手动清理"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("abort status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"outcome":"aborted"`) {
		t.Fatalf("body=%s", rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/meetings/"+meetingID+"/abort", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("second abort status=%d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestHandleGetMeetingArchive(t *testing.T) {
	dir := t.TempDir()
	s := wsfs.NewStore(dir)
	if err := s.EnsureMeeting("mtg-zip-api", "topic"); err != nil {
		t.Fatal(err)
	}

	h, err := NewHandler(config.Config{Workspace: config.Workspace{Root: dir}}, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/meetings/mtg-zip-api/archive", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/zip" {
		t.Fatalf("content-type=%q", ct)
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), "mtg-zip-api.zip") {
		t.Fatalf("disposition=%q", rec.Header().Get("Content-Disposition"))
	}
	if len(rec.Body.Bytes()) < 22 || rec.Body.Bytes()[0] != 'P' || rec.Body.Bytes()[1] != 'K' {
		t.Fatalf("expected zip magic, got %d bytes", len(rec.Body.Bytes()))
	}
}

func TestHandleDeleteMeeting(t *testing.T) {
	dir := t.TempDir()
	store := memory.New()
	ws := wsfs.NewStore(dir)
	ctx := context.Background()

	eng := engine.New(store, consensus.NoObjection{}, nil, nil, ws, nil, nil)
	freeQ := 0
	meetingID := "mtg-delete-api"
	if _, err := eng.CreateMeeting(ctx, engine.CreateMeetingInput{
		MeetingID:                meetingID,
		Topic:                    "delete api",
		ConfirmationMode:         meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment:      1,
		FreeDialogueMaxQuestions: &freeQ,
		Participants: []engine.ParticipantInput{
			{ID: "a", Role: "Architect"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	h, err := NewHandler(config.Config{Workspace: config.Workspace{Root: dir}}, nil, store, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/meetings/"+meetingID, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", rec.Code, rec.Body.String())
	}

	if _, err := ws.ReadMeetingDetail(meetingID); !errors.Is(err, workspace.ErrNotFound) {
		if err == nil {
			t.Fatal("workspace still exists")
		}
		t.Fatalf("read detail: %v", err)
	}
	envs, err := store.List(ctx, meetingID)
	if err != nil {
		t.Fatal(err)
	}
	if len(envs) != 0 {
		t.Fatalf("events remain: %d", len(envs))
	}

	req2 := httptest.NewRequest(http.MethodDelete, "/api/meetings/"+meetingID, nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("second delete status=%d body=%s", rec2.Code, rec2.Body.String())
	}
}
