package discord

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

func TestParseNaturalMeetStart_withTopicAndParticipants(t *testing.T) {
	got, ok := parseNaturalMeetStart("开个会，策划、玩家一起，聊聊骑士职业怎么设计")
	if !ok {
		t.Fatal("expected match")
	}
	if got.Topic != "骑士职业怎么设计" {
		t.Fatalf("topic=%q", got.Topic)
	}
	if !got.HasParticipantHint || got.ParticipantQuery != "策划、玩家" {
		t.Fatalf("participants=%q hint=%v", got.ParticipantQuery, got.HasParticipantHint)
	}
}

func TestParseNaturalMeetStart_multiWordDisplayName(t *testing.T) {
	got, ok := parseNaturalMeetStart("开个会，RO 老玩家代表、开发一起，聊聊骑士职业怎么设计")
	if !ok {
		t.Fatal("expected match")
	}
	if got.Topic != "骑士职业怎么设计" {
		t.Fatalf("topic=%q", got.Topic)
	}
	if got.ParticipantQuery != "RO 老玩家代表、开发" {
		t.Fatalf("participants=%q", got.ParticipantQuery)
	}
}

func TestTryBeginNaturalMeet_resolvesParticipants(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = reg.Bind(principalbind.ScopeKey("discord", "g1", "u1"), "discord", "u1", "Alice")

	meet := &MeetRunner{
		Registry: reg,
		Discord: config.DiscordTransport{
			Locale:           "zh",
			MeetParticipants: "designer:游戏策划:gameplay,player:RO 老玩家代表:experience,dev:开发:engineering",
		},
		Cfg: config.Config{
			Meeting: config.Meeting{MeetPresets: config.DefaultMeetPresets(config.Config{})},
		},
	}
	msg := transport.Inbound{
		Platform: "discord", GuildID: "g1", ChannelID: "ch1",
		AuthorID: "u1", AuthorName: "Alice",
		Content: "开个会，策划、RO 老玩家代表一起，聊聊骑士职业怎么设计",
	}
	reply, err := meet.TryBeginNaturalMeet(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "骑士职业怎么设计") {
		t.Fatalf("reply=%q", reply)
	}
	if !strings.Contains(reply, "会议简报") {
		t.Fatalf("expected brief prompt, reply=%q", reply)
	}
	sess, ok := meet.setups.get("ch1")
	if !ok || sess.step != setupStepBriefGoal {
		t.Fatalf("step=%v ok=%v", sess.step, ok)
	}
	if len(sess.config.ParticipantIDs) != 2 {
		t.Fatalf("participants=%v", sess.config.ParticipantIDs)
	}
}

func TestTryBeginNaturalMeet_resolvesParticipantsWithCasts(t *testing.T) {
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = reg.Bind(principalbind.ScopeKey("discord", "g1", "u1"), "discord", "u1", "Alice")

	meet := &MeetRunner{
		Registry: reg,
		Discord: config.DiscordTransport{
			Locale:           "zh",
			MeetParticipants: "designer:游戏策划:gameplay,player:RO 老玩家代表:experience,dev:开发:backend,ops:运维:infra",
		},
		Cfg: config.Config{
			Meeting: config.Meeting{
				MeetPresets: config.DefaultMeetPresets(config.Config{}),
				MeetCasts: []config.MeetCastConfig{
					{
						ID:             "1",
						NameZH:         "游戏策划+RO 老玩家代表+开发+运维",
						ParticipantIDs: []string{"designer", "player", "dev", "ops"},
					},
				},
			},
		},
	}
	msg := transport.Inbound{
		Platform: "discord", GuildID: "g1", ChannelID: "ch1",
		AuthorID: "u1", AuthorName: "Alice",
		Content: "开个会，策划、玩家一起，聊聊骑士职业怎么设计",
	}
	reply, err := meet.TryBeginNaturalMeet(msg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(reply, "dev·开发") || strings.Contains(reply, "ops·运维") {
		t.Fatalf("should not include all roster, reply=%q", reply)
	}
	sess, ok := meet.setups.get("ch1")
	if !ok || len(sess.config.ParticipantIDs) != 2 {
		t.Fatalf("participants=%v ok=%v", sess.config.ParticipantIDs, ok)
	}
}

func TestResolveParticipantPick_displayNamePartial(t *testing.T) {
	raw := "designer:游戏策划:gameplay,player:RO 老玩家代表:experience,dev:开发:engineering"
	ids, err := resolveParticipantPick("策划,RO 老玩家代表", raw, nil)
	if err != nil || len(ids) != 2 || ids[0] != "designer" || ids[1] != "player" {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
}
