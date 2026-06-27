package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupDeployEnv(t *testing.T, root, envContent string) {
	t.Helper()
	deployDir := filepath.Join(root, "deploy")
	if err := os.MkdirAll(deployDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deployDir, ".env.example"), []byte("# example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if envContent != "" {
		if err := os.WriteFile(filepath.Join(deployDir, ".env"), []byte(envContent), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("ROUND_TABLE_REPO_ROOT", root)
}
