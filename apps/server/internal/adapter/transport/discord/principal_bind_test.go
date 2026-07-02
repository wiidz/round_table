package discord

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	profFS "round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

func TestBindPrincipalInbound_createsProfile(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	if err := os.MkdirAll(filepath.Join(templates, "principals"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templates, "principals", "USER.md"), []byte("# user"), 0o644); err != nil {
		t.Fatal(err)
	}

	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	profiles := profFS.NewStore(dir, templates)
	msg := transport.Inbound{
		Platform: "discord", GuildID: "g1", AuthorID: "u1", AuthorName: "老皮",
	}

	reply, err := bindPrincipalInbound(reg, profiles, msg, LocaleZH)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "已绑定 Principal") {
		t.Fatalf("reply=%q", reply)
	}
	if _, err := os.Stat(filepath.Join(dir, "principals", "discord:u1", "USER.md")); err != nil {
		t.Fatalf("profile not created: %v", err)
	}
}

func TestCommandHandler_principalBind_createsProfile(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	if err := os.MkdirAll(filepath.Join(templates, "principals"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templates, "principals", "USER.md"), []byte("# user"), 0o644); err != nil {
		t.Fatal(err)
	}

	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	h := NewCommandHandler("!rt", reg, &MeetRunner{Discord: config.DiscordTransport{Locale: "zh"}})
	h.Profiles = profFS.NewStore(dir, templates)

	reply, err := h.Handle(context.Background(), transport.Inbound{
		Platform: "discord", GuildID: "g1", AuthorID: "u1", AuthorName: "老皮",
		Content: "!rt principal bind",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "已绑定 Principal") {
		t.Fatalf("reply=%q", reply)
	}
	if _, err := os.Stat(filepath.Join(dir, "principals", "discord:u1", "USER.md")); err != nil {
		t.Fatalf("profile not created: %v", err)
	}
}
