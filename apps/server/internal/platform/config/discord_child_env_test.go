package config

import (
	"path/filepath"
	"testing"
)

func TestDiscordChildEnv_usesServerSQLitePath(t *testing.T) {
	t.Setenv("ROUND_TABLE_ROOT", "")
	t.Setenv("ROUND_TABLE_REPO_ROOT", "")

	cfg := defaults()
	env := DiscordChildEnv(cfg)

	sqlite := envValue(env, "ROUND_TABLE_STORAGE_SQLITE_PATH")
	if sqlite == "" {
		t.Fatal("missing sqlite path")
	}
	if !filepath.IsAbs(sqlite) {
		t.Fatalf("sqlite path should be absolute, got %q", sqlite)
	}
	root := envValue(env, "ROUND_TABLE_ROOT")
	if root == "" {
		t.Fatal("missing ROUND_TABLE_ROOT")
	}
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if len(e) >= len(prefix) && e[:len(prefix)] == prefix {
			return e[len(prefix):]
		}
	}
	return ""
}
