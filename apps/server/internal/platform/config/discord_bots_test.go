package config

import "testing"

func TestApplyDiscordBots(t *testing.T) {
	cfg := loadBase()
	origMeet := cfg.Transport.Discord.MeetParticipants
	entries := []DiscordBotEntry{
		{ApplicationID: "1519615970128171068"},
		{ApplicationID: "1519615198170976356"},
	}
	if err := applyDiscordBots(&cfg, entries); err != nil {
		t.Fatal(err)
	}
	if cfg.Transport.Discord.ParticipantBots != "1519615970128171068,1519615198170976356" {
		t.Fatalf("participant_bots = %q", cfg.Transport.Discord.ParticipantBots)
	}
	if cfg.Transport.Discord.MeetParticipants != origMeet {
		t.Fatalf("meet_participants should be unchanged, got %q", cfg.Transport.Discord.MeetParticipants)
	}
}

func TestNormalizeDiscordBotInputs_requiresApplicationIDOrToken(t *testing.T) {
	_, _, _, _, err := normalizeDiscordBotInputs([]DiscordBotInput{
		{BoundParticipantID: "designer"},
	}, nil, nil)
	if err == nil {
		t.Fatal("expected error without token or application_id")
	}
}

func TestNormalizeDiscordBotInputs_acceptsApplicationID(t *testing.T) {
	entries, tokens, bound, _, err := normalizeDiscordBotInputs([]DiscordBotInput{
		{ApplicationID: "1519615970128171068", Token: "tok", BoundParticipantID: "designer"},
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].ApplicationID != "1519615970128171068" {
		t.Fatalf("entries = %+v", entries)
	}
	if tokens["1519615970128171068"] != "tok" {
		t.Fatalf("tokens = %+v", tokens)
	}
	if bound["1519615970128171068"] != "designer" {
		t.Fatalf("bound = %+v", bound)
	}
}

func TestBuildDiscordBotStates_includesModerator(t *testing.T) {
	cfg := loadBase()
	states := buildDiscordBotStates(cfg, nil)
	if len(states) == 0 {
		t.Fatal("expected bots")
	}
	if states[0].ID != ModeratorBotID || !states[0].Primary || states[0].Deletable {
		t.Fatalf("moderator state = %+v", states[0])
	}
}

func TestFilterDiscordParticipantBotIDs_excludesPrimary(t *testing.T) {
	cfg := loadBase()
	cfg.Transport.Discord.ParticipantBots = "111111111111111111,222222222222222222"
	overrides := map[string]string{
		DiscordModeratorRoleSetting: "222222222222222222",
	}
	got := FilterDiscordParticipantBotIDs(cfg, overrides)
	if len(got) != 1 || got[0] != "111111111111111111" {
		t.Fatalf("filtered = %v", got)
	}
}

func TestBuildDiscordBotStates_usesTokenStore(t *testing.T) {
	cfg := loadBase()
	overrides := map[string]string{
		DiscordBotTokensSetting: `{"moderator":"mod","participants":{"1519615970128171068":"a"}}`,
		DiscordBotsSetting:      `[{"application_id":"1519615970128171068"}]`,
	}
	states := buildDiscordBotStates(cfg, overrides)
	if len(states) < 2 {
		t.Fatalf("states = %+v", states)
	}
	if !states[0].Configured || states[0].TokenMasked == "" {
		t.Fatalf("moderator token_masked = %+v", states[0])
	}
}
