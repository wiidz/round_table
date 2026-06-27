package config

import (
	"path/filepath"
	"strings"
)

// legacySettingsMigrations maps deprecated app_settings keys to current keys.
var legacySettingsMigrations = map[string]string{
	"ROUND_TABLE_DISCORD_LOCALE":        "ROUND_TABLE_LOCALE",
	"ROUND_TABLE_DISCORD_MEET_MODE":     "ROUND_TABLE_DEFAULT_MEET_MODE",
	"ROUND_TABLE_DISCORD_MEET_MAX_ROUNDS": "ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT",
}

func migrateLegacySettings(overrides map[string]string) (toSet map[string]string, toDelete []string) {
	toSet = make(map[string]string)
	for oldKey, newKey := range legacySettingsMigrations {
		oldVal, ok := overrides[oldKey]
		if !ok || oldVal == "" {
			continue
		}
		if cur, exists := overrides[newKey]; exists && cur != "" {
			toDelete = append(toDelete, oldKey)
			continue
		}
		toSet[newKey] = oldVal
		toDelete = append(toDelete, oldKey)
	}
	return toSet, toDelete
}

func normalizeLoadedConfig(cfg *Config) {
	if cfg.Server.Locale == "" && cfg.Transport.Discord.Locale != "" {
		cfg.Server.Locale = cfg.Transport.Discord.Locale
	}
	if cfg.Meeting.DefaultMode == "" && cfg.Transport.Discord.MeetMode != "" {
		cfg.Meeting.DefaultMode = cfg.Transport.Discord.MeetMode
	}
	resolveDataPaths(cfg)
}

// resolveDataPaths maps ./data/* paths to the monorepo data/ root (ADR-0010).
func resolveDataPaths(cfg *Config) {
	repo := RepoRoot()
	server := ServerRoot()
	cfg.Storage.SQLitePath = resolveRepoDataPath(repo, server, cfg.Storage.SQLitePath)
	cfg.Workspace.Root = resolveRepoDataPath(repo, server, cfg.Workspace.Root)
	cfg.Profile.Root = resolveRepoDataPath(repo, server, cfg.Profile.Root)
	cfg.Profile.Templates = resolveRepoDataPath(repo, server, cfg.Profile.Templates)
	cfg.Knowledge.Root = resolveRepoDataPath(repo, server, cfg.Knowledge.Root)
	cfg.Knowledge.Templates = resolveRepoDataPath(repo, server, cfg.Knowledge.Templates)
	cfg.Transport.Discord.BindingsFile = resolveRepoDataPath(repo, server, cfg.Transport.Discord.BindingsFile)
}

func resolveRepoDataPath(repoRoot, serverRoot, rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" || filepath.IsAbs(rel) {
		return rel
	}
	if strings.HasPrefix(rel, "./data/") {
		sub := strings.TrimPrefix(rel, "./")
		return filepath.Clean(filepath.Join(repoRoot, sub))
	}
	return AbsPath(serverRoot, rel)
}
