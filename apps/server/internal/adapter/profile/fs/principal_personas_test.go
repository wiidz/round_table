package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrincipalPersonas_migrateLegacyUserMD(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	if err := os.MkdirAll(filepath.Join(templates, "principals"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templates, "principals", "USER.md"), []byte("# user\n\n## Preferences\n\n- Language: zh-CN\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := NewStore(dir, templates)
	principalID := "discord:alice"
	if err := s.EnsurePrincipal(principalID); err != nil {
		t.Fatal(err)
	}
	legacy := "# USER\n\n## Preferences\n\n- Language: zh-CN\n- Confirmation: review lists\n\n## Context\n\nGame team\n"
	if err := os.WriteFile(filepath.Join(dir, "principals", principalID, "USER.md"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest, err := s.EnsurePrincipalPersonas(principalID)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ActivePersonaID != defaultPersonaID || len(manifest.Personas) != 1 {
		t.Fatalf("manifest=%+v", manifest)
	}
	profile, err := s.ReadPrincipalPersonaUserProfile(principalID, defaultPersonaID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.Confirmation != "review lists" || profile.Context != "Game team" {
		t.Fatalf("profile=%+v", profile)
	}
}

func TestPrincipalPersonas_createAndSwitch(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	if err := os.MkdirAll(filepath.Join(templates, "principals"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templates, "principals", "USER.md"), []byte("# user"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := NewStore(dir, templates)
	principalID := "discord:bob"

	p1, err := s.CreatePrincipalPersona(principalID, "游戏策划")
	if err != nil {
		t.Fatal(err)
	}
	if p1.ID == "" || p1.Title != "游戏策划" {
		t.Fatalf("persona=%+v", p1)
	}
	p2, err := s.CreatePrincipalPersona(principalID, "产品评审")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := s.SetActivePrincipalPersona(principalID, p2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ActivePersonaID != p2.ID {
		t.Fatalf("active=%s", manifest.ActivePersonaID)
	}
	root, err := os.ReadFile(filepath.Join(dir, "principals", principalID, "USER.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(root) == 0 {
		t.Fatal("root USER.md not synced")
	}
}
