package discord

import (
	"context"
	"testing"
)

type stubSender struct {
	id string
}

func (s stubSender) Send(_ context.Context, _, _ string) error { return nil }

func TestParseParticipantBotMapping(t *testing.T) {
	got := ParseParticipantBotMapping("designer,player:MY_PLAYER_TOKEN,moderator")
	if got["designer"] != "DISCORD_BOT_TOKEN_DESIGNER" {
		t.Fatalf("designer env = %q", got["designer"])
	}
	if got["player"] != "MY_PLAYER_TOKEN" {
		t.Fatalf("player env = %q", got["player"])
	}
	if got["moderator"] != "DISCORD_BOT_TOKEN_MODERATOR" {
		t.Fatalf("moderator env = %q", got["moderator"])
	}
}

func TestBotPool_SenderFor(t *testing.T) {
	designer := stubSender{id: "designer"}
	pool := &BotPool{
		Default: stubSender{id: "main"},
		byID:    map[string]ChannelSender{"designer": designer},
	}
	if pool.SenderFor("designer") != designer {
		t.Fatal("expected designer bot")
	}
	if pool.SenderFor("unknown") != pool.Default {
		t.Fatal("expected default bot")
	}
}

func TestParticipantBotEnvKey(t *testing.T) {
	if ParticipantBotEnvKey("game-designer") != "DISCORD_BOT_TOKEN_GAME_DESIGNER" {
		t.Fatal("unexpected env key")
	}
}
