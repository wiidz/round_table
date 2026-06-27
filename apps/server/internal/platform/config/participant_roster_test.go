package config

import (
	"strings"
	"testing"
)

func TestValidateParticipantRoster_duplicateID(t *testing.T) {
	err := validateParticipantRoster([]ParticipantRosterItem{
		{ID: "a", DisplayName: "Alpha", Expertise: "x"},
		{ID: "a", DisplayName: "Beta", Expertise: "y"},
	})
	if err == nil {
		t.Fatal("expected duplicate id error")
	}
}

func TestValidateParticipantRoster_duplicateName(t *testing.T) {
	err := validateParticipantRoster([]ParticipantRosterItem{
		{ID: "a", DisplayName: "策划", Expertise: "x"},
		{ID: "b", DisplayName: "策划", Expertise: "y"},
	})
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
}

func TestValidateParticipantID(t *testing.T) {
	if err := ValidateParticipantID("Moderator"); err == nil {
		t.Fatal("uppercase should fail")
	}
	if err := ValidateParticipantID("designer"); err != nil {
		t.Fatal(err)
	}
}

func TestApplyParticipantRoster(t *testing.T) {
	cfg := defaults()
	origBots := cfg.Transport.Discord.ParticipantBots
	roster := []ParticipantRosterItem{
		{ID: "alpha", DisplayName: "Alpha", Expertise: "gameplay"},
		{ID: "beta", DisplayName: "Beta", Expertise: "ops"},
	}
	if err := applyParticipantRoster(&cfg, roster); err != nil {
		t.Fatal(err)
	}
	if cfg.Transport.Discord.ParticipantBots != origBots {
		t.Fatalf("participant_bots should be unchanged, got %q", cfg.Transport.Discord.ParticipantBots)
	}
	if !strings.Contains(cfg.Transport.Discord.MeetParticipants, "alpha:Alpha:gameplay") {
		t.Fatalf("meet = %q", cfg.Transport.Discord.MeetParticipants)
	}
}
