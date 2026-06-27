package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPriority(t *testing.T) {
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(configsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	yaml := `
server:
  addr: ":7000"
meeting:
  max_rounds_per_segment: 3
storage:
  driver: memory
`
	if err := os.WriteFile(filepath.Join(configsDir, "server.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	setupDeployEnv(t, dir, "ROUND_TABLE_ADDR=:7777\nOPENAI_API_KEY=sk-test\n")

	t.Setenv("ROUND_TABLE_ROOT", dir)
	t.Setenv("ROUND_TABLE_ADDR", ":8888")
	t.Cleanup(func() {
		_ = os.Unsetenv("OPENAI_API_KEY")
	})

	cfg := Load()
	if cfg.Addr() != ":8888" {
		t.Fatalf("env override: got %q want :8888", cfg.Addr())
	}
	if cfg.Meeting.MaxRoundsPerSegment != 3 {
		t.Fatalf("yaml: got %d want 3", cfg.Meeting.MaxRoundsPerSegment)
	}
	if cfg.Secrets.OpenAIAPIKey != "sk-test" {
		t.Fatalf("secret from .env: got %q", cfg.Secrets.OpenAIAPIKey)
	}
}

func TestSecretsNotFromYAML(t *testing.T) {
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(configsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	yaml := `
secrets:
  openai_api_key: sk-from-yaml
`
	if err := os.WriteFile(filepath.Join(configsDir, "server.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ROUND_TABLE_ROOT", dir)
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("ROUND_TABLE_DATABASE_DSN", "")
	cfg := Load()
	if cfg.Secrets.OpenAIAPIKey != "" {
		t.Fatalf("secrets must not load from yaml, got %q", cfg.Secrets.OpenAIAPIKey)
	}
}

func TestLoadModelNameFromEnv(t *testing.T) {
	dir := t.TempDir()
	configsDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(configsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configsDir, "server.yaml"), []byte("model:\n  default_model: deepseek-chat\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	setupDeployEnv(t, dir, "DEEPSEEK_MODEL_NAME=deepseek-v4-flash\n")

	t.Setenv("ROUND_TABLE_ROOT", dir)
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("ROUND_TABLE_DATABASE_DSN", "")
	t.Setenv("DEEPSEEK_MODEL_NAME", "")
	t.Setenv("ROUND_TABLE_MODEL_DEFAULT_MODEL", "")

	cfg := Load()
	if cfg.Model.DefaultModel != "deepseek-v4-flash" {
		t.Fatalf("model from .env: got %q want deepseek-v4-flash", cfg.Model.DefaultModel)
	}
}

func TestSaveEnvPreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := "# local secrets\nROUND_TABLE_ADDR=:7000\nCUSTOM=keep-me\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := defaults()
	cfg.Server.Addr = ":7777"
	if err := saveEnvFile(path, cfg); err != nil {
		t.Fatal(err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "# local secrets") || !strings.Contains(s, "ROUND_TABLE_ADDR=:7777") || !strings.Contains(s, "CUSTOM=keep-me") {
		t.Fatalf("unexpected .env content:\n%s", s)
	}
}
