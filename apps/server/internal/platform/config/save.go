package config

import (
	"fmt"
	"os"
	"strings"
)

type envEntry struct {
	key string
	val string
}

func saveEnvFile(path string, cfg Config) error {
	entries := []envEntry{
		{"ROUND_TABLE_ADDR", cfg.Server.Addr},
		{"ROUND_TABLE_READ_TIMEOUT_SEC", fmt.Sprintf("%d", cfg.Server.ReadTimeoutSec)},
		{"ROUND_TABLE_WRITE_TIMEOUT_SEC", fmt.Sprintf("%d", cfg.Server.WriteTimeoutSec)},
		{"ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT", fmt.Sprintf("%d", cfg.Meeting.MaxRoundsPerSegment)},
		{"ROUND_TABLE_MAX_CONFIRMATION_CYCLES", fmt.Sprintf("%d", cfg.Meeting.MaxConfirmationCycles)},
		{"ROUND_TABLE_CONFIRMATION_MODE", cfg.Meeting.ConfirmationMode},
		{"ROUND_TABLE_STORAGE_DRIVER", cfg.Storage.Driver},
		{"ROUND_TABLE_STORAGE_SQLITE_PATH", cfg.Storage.SQLitePath},
		{"ROUND_TABLE_DATABASE_DSN", cfg.Secrets.DatabaseDSN},
		{"OPENAI_API_KEY", cfg.Secrets.OpenAIAPIKey},
		{"ANTHROPIC_API_KEY", cfg.Secrets.AnthropicAPIKey},
		{"DEEPSEEK_API_KEY", cfg.Secrets.DeepSeekAPIKey},
		{"DEEPSEEK_MODEL_NAME", cfg.Model.DefaultModel},
	}

	existing, _ := os.ReadFile(path)
	lines := strings.Split(string(existing), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}

	known := make(map[string]bool, len(entries))
	for _, e := range entries {
		known[e.key] = true
	}

	var out []string
	seen := make(map[string]bool, len(entries))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			out = append(out, line)
			continue
		}
		key, _, ok := strings.Cut(trimmed, "=")
		key = strings.TrimSpace(key)
		if !ok || !known[key] {
			out = append(out, line)
			continue
		}
		for _, e := range entries {
			if e.key == key {
				out = append(out, fmt.Sprintf("%s=%s", e.key, e.val))
				seen[key] = true
				break
			}
		}
	}

	for _, e := range entries {
		if !seen[e.key] && e.val != "" {
			out = append(out, fmt.Sprintf("%s=%s", e.key, e.val))
		}
	}

	content := strings.Join(out, "\n")
	if content != "" {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0o600)
}
