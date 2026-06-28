package config

import "testing"

func TestNormalizeMeetCasts(t *testing.T) {
	roster := []ParticipantRosterItem{
		{ID: "designer", DisplayName: "游戏策划"},
		{ID: "player", DisplayName: "玩家代表"},
	}
	casts, err := normalizeMeetCasts([]MeetCastConfig{
		{ID: "plan-player", NameZH: "策划+玩家", ParticipantIDs: []string{"designer", "player", "designer"}},
	}, roster)
	if err != nil {
		t.Fatal(err)
	}
	if len(casts) != 1 || len(casts[0].ParticipantIDs) != 2 {
		t.Fatalf("casts=%+v", casts)
	}
}

func TestNormalizeMeetCasts_rejectsUnknownParticipant(t *testing.T) {
	roster := []ParticipantRosterItem{{ID: "designer", DisplayName: "游戏策划"}}
	_, err := normalizeMeetCasts([]MeetCastConfig{
		{ID: "x", NameZH: "x", ParticipantIDs: []string{"missing"}},
	}, roster)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeMeetCasts_autoNameFromParticipants(t *testing.T) {
	roster := []ParticipantRosterItem{
		{ID: "designer", DisplayName: "游戏策划"},
		{ID: "player", DisplayName: "玩家代表"},
	}
	casts, err := normalizeMeetCasts([]MeetCastConfig{
		{ID: "1", ParticipantIDs: []string{"designer", "player"}},
	}, roster)
	if err != nil {
		t.Fatal(err)
	}
	if casts[0].NameZH != "游戏策划+玩家代表" {
		t.Fatalf("name=%q", casts[0].NameZH)
	}
}
