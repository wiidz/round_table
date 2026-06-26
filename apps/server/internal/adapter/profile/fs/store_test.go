package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/profile"
)

func TestStore_EnsureParticipantFromTemplates(t *testing.T) {
	templates := filepath.Join(t.TempDir(), "profiles")
	profiles := t.TempDir()
	partTpl := filepath.Join(templates, "participants")
	if err := os.MkdirAll(partTpl, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(partTpl, profile.FileSoul), []byte("# soul\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewStore(profiles, templates)
	if err := s.EnsureParticipant("architect-1"); err != nil {
		t.Fatal(err)
	}
	data, err := s.ReadParticipant("architect-1", profile.FileSoul)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "soul") {
		t.Fatalf("SOUL: %s", data)
	}
	if err := s.WriteParticipant("architect-1", profile.FileSoul, []byte("custom")); err != nil {
		t.Fatal(err)
	}
	if err := s.EnsureParticipant("architect-1"); err != nil {
		t.Fatal(err)
	}
	data, _ = s.ReadParticipant("architect-1", profile.FileSoul)
	if string(data) != "custom" {
		t.Fatalf("should preserve custom SOUL, got %q", data)
	}
}
