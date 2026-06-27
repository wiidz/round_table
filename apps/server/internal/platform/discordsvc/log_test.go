package discordsvc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTailLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "discord-transport.log")
	if err := os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := tailLog(path, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.Lines != "line2\nline3" {
		t.Fatalf("lines = %q", got.Lines)
	}
}

func TestTailLog_missingFile(t *testing.T) {
	got, err := tailLog(filepath.Join(t.TempDir(), "missing.log"), 50)
	if err != nil {
		t.Fatal(err)
	}
	if got.Lines != "" {
		t.Fatalf("expected empty, got %q", got.Lines)
	}
}

func TestClearLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "discord-transport.log")
	if err := os.WriteFile(path, []byte("old line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := clearLog(path); err != nil {
		t.Fatal(err)
	}
	got, err := tailLog(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Lines, "log cleared") {
		t.Fatalf("expected cleared marker, got %q", got.Lines)
	}
}
