package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := `
# comment
ROUND_TABLE_ADDR=:7777
EXISTING=from-file
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("EXISTING", "already-set")
	if err := loadEnvFile(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("ROUND_TABLE_ADDR"); got != ":7777" {
		t.Fatalf("ROUND_TABLE_ADDR: got %q", got)
	}
	if got := os.Getenv("EXISTING"); got != "already-set" {
		t.Fatalf("must not override existing env: got %q", got)
	}
}
