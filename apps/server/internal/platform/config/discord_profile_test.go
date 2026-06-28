package config

import (
	"strings"
	"testing"
)

func TestDiscordAvatarURL_custom(t *testing.T) {
	got := discordAvatarURL("123456789012345678", "abc123")
	want := "https://cdn.discordapp.com/avatars/123456789012345678/abc123.png?size=128"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestDiscordAvatarURL_default(t *testing.T) {
	got := discordAvatarURL("123456789012345678", "")
	if got == "" || !strings.HasPrefix(got, "https://cdn.discordapp.com/embed/avatars/") {
		t.Fatalf("got %q", got)
	}
}

func TestApplyCachedDiscordBotProfiles(t *testing.T) {
	states := []DiscordBotState{
		{ID: "moderator", Label: "主持人"},
		{ID: "analyst", Label: "分析师"},
	}
	cache := map[string]DiscordBotProfileCache{
		"moderator": {
			DiscordApplicationID: "987654321098765432",
			DiscordUsername:      "mod_bot",
			AvatarURL:            "https://cdn.example/mod.png",
			FetchedAt:            "2026-01-01T00:00:00Z",
		},
	}
	got := applyCachedDiscordBotProfiles(states, cache)
	if got[0].DiscordApplicationID != "987654321098765432" {
		t.Fatalf("moderator application id = %q", got[0].DiscordApplicationID)
	}
	if got[0].DiscordUsername != "mod_bot" || got[0].AvatarURL == "" {
		t.Fatalf("moderator cache not applied: %+v", got[0])
	}
	if got[1].AvatarURL != "" {
		t.Fatalf("analyst should have no cache: %+v", got[1])
	}
}

func TestPruneDiscordBotProfilesCache(t *testing.T) {
	cache := map[string]DiscordBotProfileCache{
		"moderator": {},
		"old-bot":   {},
	}
	pruneDiscordBotProfilesCache(cache, map[string]bool{"moderator": true, "new-bot": true})
	if _, ok := cache["old-bot"]; ok {
		t.Fatal("old-bot should be removed")
	}
	if _, ok := cache["moderator"]; !ok {
		t.Fatal("moderator should remain")
	}
}

func TestBotShouldHandleCommandsForToken(t *testing.T) {
	cfg := loadBase()
	overrides := map[string]string{
		DiscordBotTokensSetting:     `{"moderator":"host-token","participants":{"111111111111111111":"other-token"}}`,
		DiscordModeratorRoleSetting: "111111111111111111",
	}
	if !BotShouldHandleCommandsForToken("host-token", cfg, overrides) {
		t.Fatal("host token should handle commands")
	}
	if BotShouldHandleCommandsForToken("other-token", cfg, overrides) {
		t.Fatal("non-host token should ignore commands")
	}
}

func TestBotShouldHandleCommands(t *testing.T) {
	overrides := map[string]string{
		DiscordModeratorRoleSetting: "222222222222222222",
		DiscordBotProfilesSetting:   `{"moderator":{"discord_application_id":"222222222222222222"}}`,
	}
	if !BotShouldHandleCommands("222222222222222222", overrides) {
		t.Fatal("host bot should handle commands")
	}
	if BotShouldHandleCommands("111111111111111111", overrides) {
		t.Fatal("non-host bot should ignore commands")
	}
	if BotShouldHandleCommands("", overrides) {
		t.Fatal("empty app id should not handle commands")
	}
}
