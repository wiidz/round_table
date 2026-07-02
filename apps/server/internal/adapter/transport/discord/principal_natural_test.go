package discord

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

func TestMatchesPrincipalBindIntent(t *testing.T) {
	for _, s := range []string{
		"绑定委托人", "绑定 principal", "bind principal", "principal bind", "我要绑定委托人",
	} {
		if !matchesPrincipalBindIntent(s) {
			t.Fatalf("expected bind intent for %q", s)
		}
	}
	for _, s := range []string{"绑定", "新增专家", "绑定一个新专家"} {
		if matchesPrincipalBindIntent(s) {
			t.Fatalf("unexpected bind intent for %q", s)
		}
	}
}

func TestExtractCreateDisplayName_principalBind(t *testing.T) {
	if got := extractCreateDisplayName("绑定委托人"); got != "" {
		t.Fatalf("got=%q want empty", got)
	}
}

func TestReception_principalBindFastPath(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	meet := &MeetRunner{Registry: reg, Discord: config.DiscordTransport{Locale: "zh"}}
	r := &Reception{
		Enabled:  true,
		Registry: reg,
		Meet:     meet,
		Locale:   func() Locale { return LocaleZH },
	}
	msg := transport.Inbound{
		Platform: "discord", GuildID: "g1", ChannelID: "ch1",
		AuthorID: "u1", AuthorName: "老皮", Content: "绑定委托人",
	}
	reply, err := r.TryHandle(context.Background(), msg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "已绑定 Principal") {
		t.Fatalf("reply=%q", reply)
	}
}

func TestReception_principalBindClarifyFollowUp(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	meet := &MeetRunner{Registry: reg, Discord: config.DiscordTransport{Locale: "zh"}}
	r := &Reception{
		Enabled:  true,
		Registry: reg,
		Meet:     meet,
		Locale:   func() Locale { return LocaleZH },
	}
	msg := transport.Inbound{
		Platform: "discord", GuildID: "g1", ChannelID: "ch1",
		AuthorID: "u1", AuthorName: "老皮", Content: "绑定",
	}
	r.storeClarifySession(msg, msg.Content, receptionDecision{
		Tool:        receptionToolClarify,
		PendingTool: receptionToolCreateParticipant,
		Message:     "请问你想绑定什么？比如绑定一个新专家？",
	})
	msg.Content = "绑定委托人"
	reply, err := r.HandleClarifyFollowUp(context.Background(), msg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(reply, "新建专家") {
		t.Fatalf("should not create expert, reply=%q", reply)
	}
	if !strings.Contains(reply, "已绑定 Principal") {
		t.Fatalf("reply=%q", reply)
	}
}

func TestCommandHandler_principalBindNatural(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	h := NewCommandHandler("!rt", reg, &MeetRunner{
		Registry: reg,
		Discord:  config.DiscordTransport{Locale: "zh"},
	})
	h.Reception = &Reception{
		Enabled:  true,
		Registry: reg,
		Meet:     h.Meet,
		Locale:   func() Locale { return LocaleZH },
	}
	reply, err := h.Handle(context.Background(), transport.Inbound{
		Platform: "discord", GuildID: "g1", ChannelID: "ch1",
		AuthorID: "u1", AuthorName: "老皮", Content: "绑定委托人",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "已绑定 Principal") {
		t.Fatalf("reply=%q", reply)
	}
}
