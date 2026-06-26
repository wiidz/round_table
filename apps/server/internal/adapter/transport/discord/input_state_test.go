package discord

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

func TestInputPhase_setupTopic(t *testing.T) {
	r := &MeetRunner{Discord: config.DiscordTransport{Locale: "zh"}}
	r.setups.put("ch1", meetSetupSession{channelID: "ch1", step: setupStepAskTopic})
	if got := r.InputPhase("ch1"); got != InputPhaseSetupTopic {
		t.Fatalf("got=%q", got)
	}
}

func TestHandleInputStatus_naturalTrigger(t *testing.T) {
	r := &MeetRunner{Discord: config.DiscordTransport{Locale: "zh"}}
	r.setups.put("ch1", meetSetupSession{channelID: "ch1", step: setupStepAskTopic})
	reply, err := r.HandleInputStatus(transport.Inbound{ChannelID: "ch1", Content: "会议状态"})
	if err != nil || !strings.Contains(reply, "当前输入态") || !strings.Contains(reply, "会议主题") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func TestCommandHandler_rtStatus(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	meet := &MeetRunner{
		Registry: reg,
		Discord:  config.DiscordTransport{Locale: "zh"},
	}
	meet.setups.put("ch1", meetSetupSession{channelID: "ch1", step: setupStepPresetMenu})
	h := NewCommandHandler("!rt", reg, meet)

	reply, err := h.Handle(context.Background(), transport.Inbound{
		ChannelID: "ch1", Content: "!rt status",
	})
	if err != nil || !strings.Contains(reply, "选择预设") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func TestMisplacedInputHint_paused(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	scope := principalbind.ScopeKey("discord", "g1", "u1")
	if _, err := reg.Bind(scope, "discord", "u1", "Alice"); err != nil {
		t.Fatal(err)
	}

	p := NewChannelPrincipal(&BotPool{Default: &captureSender{}}, "zh")
	p.BindMeeting("mtg-1", "ch1", "u1")
	p.sessions["mtg-1"].paused = true

	r := &MeetRunner{
		Registry:  reg,
		Principal: p,
		Discord:   config.DiscordTransport{Locale: "zh"},
	}
	r.sessions.tryStart("ch1", "mtg-1")

	hint, ok := r.MisplacedInputHint(transport.Inbound{
		Platform: "discord", GuildID: "g1", AuthorID: "u1",
		ChannelID: "ch1", Content: "随便说点什么",
	})
	if !ok || !strings.Contains(hint, "恢复会议") {
		t.Fatalf("hint=%q ok=%v", hint, ok)
	}
}
