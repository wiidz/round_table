package fs

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteMeetingArchive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	if err := s.EnsureMeeting("mtg-zip", "topic"); err != nil {
		t.Fatal(err)
	}
	usageDir := filepath.Join(dir, "mtg-zip", "usage")
	if err := os.MkdirAll(usageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usageDir, "tokens.jsonl"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := s.WriteMeetingArchive("mtg-zip", &buf); err != nil {
		t.Fatal(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	names := make(map[string]struct{}, len(zr.File))
	for _, f := range zr.File {
		names[f.Name] = struct{}{}
	}
	if _, ok := names["MEETING.md"]; !ok {
		t.Fatalf("missing MEETING.md in %v", names)
	}
	if _, ok := names["usage/tokens.jsonl"]; !ok {
		t.Fatalf("missing usage/tokens.jsonl in %v", names)
	}
}

func TestDeleteMeeting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s := NewStore(dir)
	if err := s.EnsureMeeting("mtg-del", "topic"); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteMeeting("mtg-del"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "mtg-del")); !os.IsNotExist(err) {
		t.Fatalf("dir still exists: %v", err)
	}
	if err := s.DeleteMeeting("mtg-del"); err == nil {
		t.Fatal("want not found")
	}
}
