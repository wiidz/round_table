package config

import (
	"os"
	"path/filepath"
	"strings"
)

// configRoot returns apps/server whether make is run from repo root or go run from apps/server.
func configRoot() string {
	if v := os.Getenv("ROUND_TABLE_ROOT"); v != "" {
		return v
	}
	if _, err := os.Stat("apps/server/configs/server.yaml"); err == nil {
		return "apps/server"
	}
	return "."
}

// repoRoot returns the monorepo root (contains deploy/.env.example).
func repoRoot() string {
	if v := os.Getenv("ROUND_TABLE_REPO_ROOT"); v != "" {
		return v
	}
	candidates := []string{".", "..", "../.."}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "deploy", ".env.example")); err == nil {
			return c
		}
	}
	return "."
}

// RepoRoot returns the monorepo root (contains deploy/.env.example).
func RepoRoot() string {
	return repoRoot()
}

// ServerRoot returns apps/server (config + relative data paths).
func ServerRoot() string {
	return configRoot()
}

// AbsPath resolves rel against base; abs paths are unchanged.
func AbsPath(base, rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return base
	}
	if filepath.IsAbs(rel) {
		return filepath.Clean(rel)
	}
	baseAbs, err := filepath.Abs(base)
	if err != nil {
		baseAbs = base
	}
	return filepath.Clean(filepath.Join(baseAbs, rel))
}

// DiscordTransportPIDPath is the pid file written by the discord transport child process.
func DiscordTransportPIDPath(cfg Config) string {
	serverRoot, _ := filepath.Abs(ServerRoot())
	sqlite := AbsPath(serverRoot, cfg.Storage.SQLitePath)
	return filepath.Join(filepath.Dir(sqlite), "logs", "discord-transport.pid")
}

// DiscordInboundDedupDir stores cross-process inbound message claim files.
func DiscordInboundDedupDir(cfg Config) string {
	return filepath.Join(filepath.Dir(DiscordTransportPIDPath(cfg)), "discord-inbound-dedup")
}

// deployEnvPath is the single local secrets file (deploy/.env).
func deployEnvPath() string {
	return filepath.Join(repoRoot(), "deploy", ".env")
}
