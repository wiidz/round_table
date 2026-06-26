package fs

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/knowledge"
)

func TestStore_IsolatedParticipantKnowledge(t *testing.T) {
	root := t.TempDir()
	templates := t.TempDir()
	s := NewStore(root, templates)

	if err := s.Ensure(knowledge.ScopeParticipant, "p1"); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteMemory(knowledge.ScopeParticipant, "p1", []byte("# p1 memory")); err != nil {
		t.Fatal(err)
	}
	if err := s.Ensure(knowledge.ScopeParticipant, "p2"); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteMemory(knowledge.ScopeParticipant, "p2", []byte("# p2 memory")); err != nil {
		t.Fatal(err)
	}

	m1, _ := s.ReadMemory(knowledge.ScopeParticipant, "p1")
	m2, _ := s.ReadMemory(knowledge.ScopeParticipant, "p2")
	if string(m1) == string(m2) {
		t.Fatal("participant knowledge must be isolated")
	}
}

func TestStore_SharedPool(t *testing.T) {
	s := NewStore(t.TempDir(), t.TempDir())
	if err := s.WriteMemory(knowledge.ScopeShared, "", []byte("# shared")); err != nil {
		t.Fatal(err)
	}
	data, err := s.ReadMemory(knowledge.ScopeShared, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "shared") {
		t.Fatalf("shared memory: %s", data)
	}
}

func TestStore_AppendDailyLog(t *testing.T) {
	s := NewStore(t.TempDir(), t.TempDir())
	if err := s.AppendDailyLog(knowledge.ScopeParticipant, "p1", "2026-06-26", []byte("line1\n")); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendDailyLog(knowledge.ScopeParticipant, "p1", "2026-06-26", []byte("line2\n")); err != nil {
		t.Fatal(err)
	}
	data, err := s.ReadDailyLog(knowledge.ScopeParticipant, "p1", "2026-06-26")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "line1") || !strings.Contains(string(data), "line2") {
		t.Fatalf("daily log: %s", data)
	}
}
