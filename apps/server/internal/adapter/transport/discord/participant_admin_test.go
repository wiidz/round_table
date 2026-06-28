package discord

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/storage/sqlite"
	"round_table/apps/server/internal/adapter/transport"
	"round_table/apps/server/internal/platform/config"
)

func writeParticipantTemplates(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, "participants")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SOUL.md"), []byte("# SOUL\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# AGENTS\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParticipantAdmin_createWizard(t *testing.T) {
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "rt.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}

	tplRoot := t.TempDir()
	writeParticipantTemplates(t, tplRoot)
	profileRoot := t.TempDir()
	admin := &ParticipantAdmin{
		ConfigSvc: svc,
		Profile:   fs.NewStore(profileRoot, tplRoot),
		Locale:    func() Locale { return LocaleZH },
		Prefix:    "!rt ",
	}

	msg := transport.Inbound{Platform: "discord", ChannelID: "ch1", AuthorID: "u1"}

	reply, err := admin.HandleCommand(msg, []string{"new"})
	if err != nil || !strings.Contains(reply, "1/4") {
		t.Fatalf("begin create: reply=%q err=%v", reply, err)
	}

	steps := []string{"LOL 玩家代表", "player_lol", "moba", "1"}
	for _, body := range steps {
		msg.Content = body
		reply, err = admin.HandleSetupReply(msg)
		if err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(reply, "已创建专家") || !strings.Contains(reply, "player_lol") {
		t.Fatalf("create result=%q", reply)
	}

	roster := config.ParticipantRosterFromConfig(svc.Current())
	found := false
	for _, item := range roster {
		if item.ID == "player_lol" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("roster=%v", roster)
	}
}

func TestParticipantAdmin_listAndDelete(t *testing.T) {
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "rt.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	if err := svc.CreateParticipant(ctx, config.ParticipantRosterItem{
		ID: "alpha", DisplayName: "Alpha", Expertise: "test",
	}); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateParticipant(ctx, config.ParticipantRosterItem{
		ID: "beta", DisplayName: "Beta", Expertise: "test",
	}); err != nil {
		t.Fatal(err)
	}

	admin := &ParticipantAdmin{
		ConfigSvc: svc,
		Locale:    func() Locale { return LocaleZH },
		Prefix:    "!rt ",
	}
	msg := transport.Inbound{Platform: "discord", ChannelID: "ch1", AuthorID: "u1"}

	list, err := admin.HandleCommand(msg, []string{"list"})
	if err != nil || !strings.Contains(list, "alpha") || !strings.Contains(list, "beta") {
		t.Fatalf("list=%q err=%v", list, err)
	}

	reply, err := admin.HandleCommand(msg, []string{"delete", "beta"})
	if err != nil || !strings.Contains(reply, "确认删除") {
		t.Fatalf("delete prompt=%q err=%v", reply, err)
	}
	msg.Content = "1"
	reply, err = admin.HandleSetupReply(msg)
	if err != nil || !strings.Contains(reply, "已删除") {
		t.Fatalf("delete result=%q err=%v", reply, err)
	}
}

func TestCommandHandler_expertCommand(t *testing.T) {
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "rt.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}
	meet := &MeetRunner{
		ConfigSvc: svc,
		Discord:   config.DiscordTransport{Locale: "zh", MeetParticipants: "designer:游戏策划:gameplay"},
	}
	h := NewCommandHandler("!rt", nil, meet)
	h.Participants = &ParticipantAdmin{ConfigSvc: svc, Locale: func() Locale { return LocaleZH }, Prefix: h.Prefix}

	reply, err := h.Handle(t.Context(), transport.Inbound{
		Platform: "discord", ChannelID: "ch1", AuthorID: "u1", Content: "!rt 专家 列表",
	})
	if err != nil || !strings.Contains(reply, "专家名录") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func TestSlugParticipantID(t *testing.T) {
	if got := slugParticipantID("LOL Player"); got != "lol_player" {
		t.Fatalf("got=%q", got)
	}
	if got := slugParticipantID("纯中文名"); got != "" {
		t.Fatalf("got=%q", got)
	}
}
