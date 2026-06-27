package config

import (
	"strings"
	"testing"
)

func TestMergeDiscordBotTokenUpdates(t *testing.T) {
	current := DiscordBotTokens{
		Moderator: "mod-old",
		Participants: map[string]string{
			"a": "tok-a",
			"b": "tok-b",
		},
	}
	got := mergeDiscordBotTokenUpdates(current, ModeratorBotID, "mod-new", "", map[string]string{
		"c": "tok-c",
	}, []string{"a", "c"})
	if got.Moderator != "mod-new" {
		t.Fatalf("moderator = %q", got.Moderator)
	}
	if got.Participants["a"] != "tok-a" {
		t.Fatalf("a = %+v", got.Participants)
	}
	if _, ok := got.Participants["b"]; ok {
		t.Fatal("b should be removed")
	}
	if got.Participants["c"] != "tok-c" {
		t.Fatalf("c = %+v", got.Participants)
	}
}

func TestMaskSecretToken(t *testing.T) {
	if got := MaskSecretToken(""); got != "" {
		t.Fatalf("empty = %q", got)
	}
	if got := MaskSecretToken("abc"); len(got) < 12 || !strings.HasPrefix(got, "*") {
		t.Fatalf("short token mask = %q", got)
	}
}

func TestDiscordBotTokensTokenFor(t *testing.T) {
	tokens := DiscordBotTokens{
		Moderator:    "mod",
		Participants: map[string]string{"designer": "d"},
	}
	if tokens.TokenForBot("moderator", ModeratorBotID) != "mod" {
		t.Fatal("moderator token when primary")
	}
	if tokens.TokenForBot("designer", ModeratorBotID) != "d" {
		t.Fatal("designer token when not primary")
	}
	if tokens.TokenForBot("designer", "designer") != "mod" {
		t.Fatal("designer token when primary")
	}
}

func TestSwapPrimaryBotTokens(t *testing.T) {
	tokens := DiscordBotTokens{
		Moderator: "mod-tok",
		Participants: map[string]string{
			"designer": "des-tok",
		},
	}
	got := swapPrimaryBotTokens(tokens, ModeratorBotID, "designer")
	if got.Moderator != "des-tok" {
		t.Fatalf("moderator = %q", got.Moderator)
	}
	if got.Participants["moderator"] != "mod-tok" {
		t.Fatalf("participants[moderator] = %+v", got.Participants)
	}
	if _, ok := got.Participants["designer"]; ok {
		t.Fatal("designer should be removed from participants")
	}
}

func TestBuildDiscordBotStates_primaryRoleSwitch(t *testing.T) {
	cfg := loadBase()
	const designerApp = "1519615970128171068"
	overrides := map[string]string{
		DiscordBotTokensSetting:     `{"moderator":"des-tok","participants":{"moderator":"mod-tok"}}`,
		DiscordBotsSetting:          `[{"application_id":"` + designerApp + `"}]`,
		DiscordModeratorRoleSetting: designerApp,
	}
	states := buildDiscordBotStates(cfg, overrides)
	if len(states) < 2 {
		t.Fatalf("states = %+v", states)
	}
	if states[0].Primary {
		t.Fatal("moderator role should not be primary")
	}
	if !states[1].Primary || states[1].ID != designerApp {
		t.Fatalf("designer state = %+v", states[1])
	}
	if states[1].Token != "des-tok" {
		t.Fatalf("designer token = %q", states[1].Token)
	}
	if states[0].Token != "mod-tok" {
		t.Fatalf("moderator role token = %q", states[0].Token)
	}
}
