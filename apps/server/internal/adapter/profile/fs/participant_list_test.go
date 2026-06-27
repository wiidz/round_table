package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListParticipants(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	writeParticipantTemplate(t, templates)

	s := NewStore(dir, templates)
	if err := s.EnsureParticipant("designer"); err != nil {
		t.Fatal(err)
	}

	list, err := s.ListParticipants()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != "designer" {
		t.Fatalf("list=%+v", list)
	}
	if len(list[0].Files) < 3 {
		t.Fatalf("files=%v", list[0].Files)
	}
}

func TestReadParticipantDetail(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	writeParticipantTemplate(t, templates)

	s := NewStore(dir, templates)
	if err := s.EnsureParticipant("dev"); err != nil {
		t.Fatal(err)
	}
	detail, err := s.ReadParticipantDetail("dev")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Files["SOUL.md"] == "" || detail.Files["AGENTS.md"] == "" {
		t.Fatalf("files=%v", detail.Files)
	}
}

func writeParticipantTemplate(t *testing.T, templates string) {
	t.Helper()
	base := filepath.Join(templates, "participants")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"SOUL.md", "AGENTS.md", "TOOLS.md"} {
		if err := os.WriteFile(filepath.Join(base, name), []byte("# "+name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
