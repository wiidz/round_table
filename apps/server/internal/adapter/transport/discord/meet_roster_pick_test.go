package discord

import (
	"testing"

	"round_table/apps/server/internal/platform/config"
)

func TestResolveParticipantPick_cast(t *testing.T) {
	raw := "designer:游戏策划:gameplay,player:玩家代表:experience,dev:主程:engineering"
	casts := []config.MeetCastConfig{
		{ID: "1", NameZH: "策划+玩家", ParticipantIDs: []string{"designer", "player"}},
	}
	ids, err := resolveParticipantPick("C1", raw, casts)
	if err != nil || len(ids) != 2 || ids[0] != "designer" || ids[1] != "player" {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
}

func TestResolveParticipantPick_byIndex(t *testing.T) {
	raw := "designer:游戏策划:gameplay,player:玩家代表:experience"
	ids, err := resolveParticipantPick("1,2", raw, nil)
	if err != nil || len(ids) != 2 {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
}

func TestResolveParticipantPick_byDisplayName(t *testing.T) {
	raw := "designer:游戏策划:gameplay,player:玩家代表:experience"
	ids, err := resolveParticipantPick("策划,玩家", raw, nil)
	if err != nil || len(ids) != 2 {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
}

func TestResolveRosterPick_ignoresCasts(t *testing.T) {
	raw := "designer:游戏策划:gameplay,player:RO 老玩家代表:experience,dev:开发:backend,ops:运维:infra"
	casts := []config.MeetCastConfig{
		{
			ID:             "1",
			NameZH:         "游戏策划+RO 老玩家代表+开发+运维",
			ParticipantIDs: []string{"designer", "player", "dev", "ops"},
		},
	}
	ids, err := resolveRosterPick("策划,玩家", raw)
	if err != nil || len(ids) != 2 || ids[0] != "designer" || ids[1] != "player" {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
	if _, err := resolveRosterPick("C1", raw); err == nil {
		t.Fatal("C1 should not resolve via roster pick")
	}
	// Cast shortcut remains available in the wizard path only.
	ids, err = resolveParticipantPick("C1", raw, casts)
	if err != nil || len(ids) != 4 {
		t.Fatalf("wizard cast pick ids=%v err=%v", ids, err)
	}
}

func TestResolveParticipantPick_byDisplayNameWithCasts(t *testing.T) {
	raw := "designer:游戏策划:gameplay,player:RO 老玩家代表:experience,dev:开发:backend,ops:运维:infra"
	casts := []config.MeetCastConfig{
		{
			ID:             "1",
			NameZH:         "游戏策划+RO 老玩家代表+开发+运维",
			ParticipantIDs: []string{"designer", "player", "dev", "ops"},
		},
		{
			ID:             "2",
			NameZH:         "RO 老玩家代表+开发",
			ParticipantIDs: []string{"player", "dev"},
		},
	}
	ids, err := resolveParticipantPick("策划,玩家", raw, casts)
	if err != nil || len(ids) != 2 || ids[0] != "designer" || ids[1] != "player" {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
}

func TestResolveParticipantPick_all(t *testing.T) {
	raw := "designer:游戏策划:gameplay"
	ids, err := resolveParticipantPick("0", raw, nil)
	if err != nil || ids != nil {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
}

func TestParseParticipantsFiltered(t *testing.T) {
	raw := "designer:游戏策划:gameplay,player:玩家代表:experience,dev:主程:engineering"
	parts, err := parseParticipantsFiltered(raw, []string{"designer", "player"})
	if err != nil || len(parts) != 2 {
		t.Fatalf("parts=%v err=%v", parts, err)
	}
}
