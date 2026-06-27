package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPruneMisplacedBotProfiles(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	writeParticipantTemplate(t, templates)

	s := NewStore(dir, templates)
	for _, id := range []string{"designer", "1519615970128171068", "app-1520303229869756487"} {
		if err := s.EnsureParticipant(id); err != nil {
			t.Fatal(err)
		}
	}

	if err := s.PruneMisplacedBotProfiles(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "participants", "designer")); err != nil {
		t.Fatal("designer profile should remain")
	}
	if _, err := os.Stat(filepath.Join(dir, "participants", "1519615970128171068")); !os.IsNotExist(err) {
		t.Fatal("application id profile should be removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "participants", "app-1520303229869756487")); !os.IsNotExist(err) {
		t.Fatal("app- profile should be removed")
	}
}
