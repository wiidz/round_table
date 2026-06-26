package fs

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/workspace"
)

func TestStore_EnsureMeetingAndReadWrite(t *testing.T) {
	s := NewStore(t.TempDir())

	if err := s.EnsureMeeting("m-1", "Design API"); err != nil {
		t.Fatal(err)
	}
	data, err := s.Read("m-1", workspace.FileMeeting)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Design API") {
		t.Fatalf("MEETING.md: %s", data)
	}

	if err := s.Write("m-1", "artifacts/proposal.md", []byte("# Proposal\n")); err != nil {
		t.Fatal(err)
	}
	got, err := s.Read("m-1", "artifacts/proposal.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "# Proposal\n" {
		t.Fatalf("artifact: %q", got)
	}
}

func TestStore_PathJail(t *testing.T) {
	s := NewStore(t.TempDir())
	if _, err := s.Resolve("m-1", "../etc/passwd"); err != workspace.ErrOutsideRoot {
		t.Fatalf("expected ErrOutsideRoot, got %v", err)
	}
}

func TestStore_List(t *testing.T) {
	s := NewStore(t.TempDir())
	_ = s.EnsureMeeting("m-1", "t")
	_ = s.Write("m-1", "artifacts/a.md", []byte("a"))
	_ = s.Write("m-1", "artifacts/b.md", []byte("b"))

	names, err := s.List("m-1", workspace.DirArtifacts)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("list: %v", names)
	}
}
