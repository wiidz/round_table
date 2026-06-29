package web

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/transport"
	"round_table/apps/server/internal/platform/config"
)

func testHandler(t *testing.T) *Handler {
	t.Helper()
	return NewHandler(config.Load(), nil, NewHub())
}

func TestHandlerWebStatusNoPrincipal(t *testing.T) {
	h := testHandler(t)
	reply, err := h.Handle(t.Context(), transport.Inbound{
		Platform:  "web",
		ChannelID: "sess-1",
		AuthorID:  "sess-1",
		Content:   "会议状态",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reply.AsModerator {
		t.Fatal("status reply should be moderator")
	}
	if strings.Contains(reply.Content, "principal bind") {
		t.Fatalf("web status must not mention principal bind: %q", reply.Content)
	}
}

func TestHandlerWebExpertListNatural(t *testing.T) {
	h := testHandler(t)
	reply, err := h.Handle(t.Context(), transport.Inbound{
		Platform:  "web",
		ChannelID: "sess-1",
		AuthorID:  "sess-1",
		Content:   "有哪些专家",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reply.AsModerator && reply.Content == "" {
		t.Fatal("expected expert list or fallback")
	}
}

func TestHandlerWebNoMatchFallback(t *testing.T) {
	h := testHandler(t)
	reply, err := h.Handle(t.Context(), transport.Inbound{
		Platform:  "web",
		ChannelID: "sess-1",
		AuthorID:  "sess-1",
		Content:   "xyzunknown",
	})
	if err != nil {
		t.Fatal(err)
	}
	if reply.AsModerator {
		t.Fatal("no-match should be system")
	}
	if !strings.Contains(reply.Content, "暂未理解") && !strings.Contains(strings.ToLower(reply.Content), "didn't understand") {
		t.Fatalf("reply=%q", reply.Content)
	}
}

func TestHandlerWebPrincipalCommandRejected(t *testing.T) {
	h := testHandler(t)
	reply, err := h.Handle(t.Context(), transport.Inbound{
		Platform:  "web",
		ChannelID: "sess-1",
		AuthorID:  "sess-1",
		Content:   "!rt principal bind",
	})
	if err != nil {
		t.Fatal(err)
	}
	if reply.AsModerator {
		t.Fatal("principal block should be system")
	}
	if !strings.Contains(reply.Content, "Principal") {
		t.Fatalf("reply=%q", reply.Content)
	}
}

func TestHandlerWebMeetStartEntersSetup(t *testing.T) {
	h := testHandler(t)
	reply, err := h.Handle(t.Context(), transport.Inbound{
		Platform:  "web",
		ChannelID: "sess-1",
		AuthorID:  "sess-1",
		Content:   "开个会",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reply.AsModerator {
		t.Fatal("meet setup prompt should be moderator")
	}
	if strings.Contains(reply.Content, "暂不支持") || strings.Contains(strings.ToLower(reply.Content), "cannot start") {
		t.Fatalf("web should enter meet setup, got: %q", reply.Content)
	}
	if !strings.Contains(reply.Content, "主题") && !strings.Contains(strings.ToLower(reply.Content), "topic") {
		t.Fatalf("expected topic prompt, got: %q", reply.Content)
	}
}
