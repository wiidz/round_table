package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListPrincipals(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	writePrincipalTemplate(t, templates)

	s := NewStore(dir, templates)
	if err := s.EnsurePrincipal("discord:alice"); err != nil {
		t.Fatal(err)
	}
	extra := filepath.Join(dir, "principals", "discord:alice", "SOUL.md")
	if err := os.WriteFile(extra, []byte("# soul"), 0o644); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListPrincipals()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("len=%d", len(list))
	}
	if list[0].ID != "discord:alice" {
		t.Fatalf("id=%q", list[0].ID)
	}
	if len(list[0].Files) < 2 {
		t.Fatalf("files=%v", list[0].Files)
	}
}

func TestReadPrincipalDetail(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	writePrincipalTemplate(t, templates)

	s := NewStore(dir, templates)
	if err := s.EnsurePrincipal("discord:bob"); err != nil {
		t.Fatal(err)
	}
	detail, err := s.ReadPrincipalDetail("discord:bob")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Files["USER.md"] == "" {
		t.Fatal("missing USER.md")
	}
}

func writePrincipalTemplate(t *testing.T, templates string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(templates, "principals"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templates, "principals", "USER.md"), []byte("# user"), 0o644); err != nil {
		t.Fatal(err)
	}
}
