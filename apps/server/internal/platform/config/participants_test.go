package config

import (
	"path/filepath"
	"testing"
)

func TestResolveRepoDataPath(t *testing.T) {
	repo := "/repo"
	server := "/repo/apps/server"
	got := resolveRepoDataPath(repo, server, "./data/profiles")
	want := filepath.Join(repo, "data/profiles")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if resolveRepoDataPath(repo, server, "/abs/profiles") != "/abs/profiles" {
		t.Fatal("abs unchanged")
	}
}

func TestMeetParticipantIDs(t *testing.T) {
	raw := "designer:游戏策划:gameplay,player:玩家代表,dev"
	got := MeetParticipantIDs(raw)
	want := []string{"designer", "player", "dev"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestParseMeetParticipants(t *testing.T) {
	m := ParseMeetParticipants("designer:游戏策划:gameplay,player:玩家代表:experience")
	if m["designer"].DisplayName != "游戏策划" || m["designer"].Expertise != "gameplay" {
		t.Fatalf("designer = %+v", m["designer"])
	}
	if m["player"].DisplayName != "玩家代表" {
		t.Fatalf("player = %+v", m["player"])
	}
}
