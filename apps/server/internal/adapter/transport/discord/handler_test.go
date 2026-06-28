package discord

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

func TestCommandHandler_principalBind(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	h := NewCommandHandler("!rt", reg, &MeetRunner{Discord: config.DiscordTransport{Locale: "zh"}})

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

func TestCommandHandler_helpEN(t *testing.T) {
	h := NewCommandHandler("!rt", mustReg(t), &MeetRunner{Discord: config.DiscordTransport{Locale: "en"}})
	reply, err := h.Handle(context.Background(), transport.Inbound{Content: "!rt help"})
	if err != nil || !strings.Contains(reply, "RoundTable Discord commands") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func TestNewCommandHandler_wiresMeet(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	meet := &MeetRunner{Registry: reg}
	h := NewCommandHandler("!rt", reg, meet)
	if h.Meet != meet {
		t.Fatal("Meet runner not wired")
	}
}

func TestCommandHandler_naturalMeetTrigger(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	meet := &MeetRunner{
		Registry: reg,
		Discord:  config.DiscordTransport{Locale: "zh"},
	}
	h := NewCommandHandler("!rt", reg, meet)

	_, _ = h.Handle(context.Background(), transport.Inbound{
		Platform: "discord", GuildID: "g1", AuthorID: "u1", AuthorName: "Alice",
		ChannelID: "ch1", Content: "!rt principal bind",
	})

	reply, err := h.Handle(context.Background(), transport.Inbound{
		Platform: "discord", GuildID: "g1", AuthorID: "u1",
		ChannelID: "ch1", Content: "新会议",
	})
	if err != nil || !strings.Contains(reply, "请输入会议主题") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}

	reply, err = h.Handle(context.Background(), transport.Inbound{
		Platform: "discord", GuildID: "g1", AuthorID: "u1",
		ChannelID: "ch1", Content: "影舞者设计",
	})
	if err != nil || !strings.Contains(reply, "请选择参会专家") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}

	reply, err = h.Handle(context.Background(), transport.Inbound{
		Platform: "discord", GuildID: "g1", AuthorID: "u1",
		ChannelID: "ch1", Content: "0",
	})
	if err != nil || !strings.Contains(reply, "会议简报") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
	for _, skip := range []string{"-", "-", "-"} {
		reply, err = h.Handle(context.Background(), transport.Inbound{
			Platform: "discord", GuildID: "g1", AuthorID: "u1",
			ChannelID: "ch1", Content: skip,
		})
		if err != nil {
			t.Fatalf("brief skip reply=%q err=%v", reply, err)
		}
	}
	if !strings.Contains(reply, "请选择会议方案") || !strings.Contains(reply, "直接开始（默认）") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func TestCommandHandler_nonCommandSilent(t *testing.T) {
	h := NewCommandHandler("!rt", mustReg(t), nil)
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
