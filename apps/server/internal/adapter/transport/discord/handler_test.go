package discord

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
)

func TestCommandHandler_principalBind(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	h := NewCommandHandler("!rt", reg)

	reply, err := h.Handle(context.Background(), transport.Inbound{
		Platform:   "discord",
		GuildID:    "guild-1",
		AuthorID:   "user-1",
		AuthorName: "Alice",
		Content:    "!rt principal bind",
	})
	if err != nil {
		t.Fatal(err)
	}
	if reply == "" || !strings.Contains(reply, "discord:user-1") {
		t.Fatalf("reply = %q", reply)
	}

	reply2, _ := h.Handle(context.Background(), transport.Inbound{
		Platform: "discord", GuildID: "guild-1", AuthorID: "user-2", Content: "!rt principal bind",
	})
	if !strings.Contains(reply2, "绑定失败") {
		t.Fatalf("expected conflict, got %q", reply2)
	}

	reply3, _ := h.Handle(context.Background(), transport.Inbound{
		Platform: "discord", GuildID: "guild-1", AuthorID: "user-1", Content: "!rt principal whoami",
	})
	if !strings.Contains(reply3, "你是本范围的 Principal") {
		t.Fatalf("whoami = %q", reply3)
	}
}

func TestCommandHandler_nonCommandSilent(t *testing.T) {
	h := NewCommandHandler("!rt", mustReg(t))
	reply, err := h.Handle(context.Background(), transport.Inbound{Content: "hello"})
	if err != nil || reply != "" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func mustReg(t *testing.T) *principalbind.Registry {
	t.Helper()
	r, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	return r
}
