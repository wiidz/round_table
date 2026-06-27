package config

import "testing"

func TestAssignDiscordPair(t *testing.T) {
	b := make(ParticipantIMBindings)
	assignDiscordPair(b, "1519615970128171068", "designer")
	if got := discordApplicationForParticipant(b, "designer"); got != "1519615970128171068" {
		t.Fatalf("got %q", got)
	}
}

func TestEnforceDiscordBotOwnership(t *testing.T) {
	b := ParticipantIMBindings{
		"a": {{Platform: IMPlatformDiscord, ApplicationID: "1519615970128171068"}},
		"b": {{Platform: IMPlatformDiscord, ApplicationID: "1519615970128171068"}},
	}
	setParticipantIMBindings(b, "b", []ParticipantIMBind{{Platform: IMPlatformDiscord, ApplicationID: "1519615970128171068"}})
	if _, ok := b["a"]; ok {
		t.Fatal("expected a unbound")
	}
}

func TestValidateParticipantIMBindings_duplicateBot(t *testing.T) {
	b := ParticipantIMBindings{
		"a": {{Platform: IMPlatformDiscord, ApplicationID: "1519615970128171068"}},
		"b": {{Platform: IMPlatformDiscord, ApplicationID: "1519615970128171068"}},
	}
	err := validateParticipantIMBindings(b, map[string]struct{}{"a": {}, "b": {}}, map[string]struct{}{"1519615970128171068": {}})
	if err == nil {
		t.Fatal("expected duplicate bot error")
	}
}

func TestParticipantForDiscordBot(t *testing.T) {
	b := ParticipantIMBindings{
		"designer": {{Platform: IMPlatformDiscord, ApplicationID: "1519615970128171068"}},
	}
	if got := participantForDiscordBot(b, "1519615970128171068"); got != "designer" {
		t.Fatalf("got %q", got)
	}
}
