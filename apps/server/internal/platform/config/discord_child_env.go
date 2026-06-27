package config

import (
	"os"
	"path/filepath"
	"strings"
)

// DiscordChildEnv builds environment for the discord transport subprocess so it
// reads the same SQLite and data paths as the running HTTP server.
func DiscordChildEnv(cfg Config) []string {
	serverRoot, _ := filepath.Abs(ServerRoot())
	repoRoot, _ := filepath.Abs(RepoRoot())

	env := append([]string{}, os.Environ()...)
	env = upsertEnv(env, "ROUND_TABLE_REPO_ROOT", repoRoot)
	env = upsertEnv(env, "ROUND_TABLE_ROOT", serverRoot)
	env = upsertEnv(env, "ROUND_TABLE_STORAGE_DRIVER", cfg.Storage.Driver)
	env = upsertEnv(env, "ROUND_TABLE_STORAGE_SQLITE_PATH", AbsPath(serverRoot, cfg.Storage.SQLitePath))
	env = upsertEnv(env, "ROUND_TABLE_WORKSPACE_ROOT", AbsPath(serverRoot, cfg.Workspace.Root))
	env = upsertEnv(env, "ROUND_TABLE_PROFILE_ROOT", AbsPath(serverRoot, cfg.Profile.Root))
	env = upsertEnv(env, "ROUND_TABLE_PROFILE_TEMPLATES", AbsPath(serverRoot, cfg.Profile.Templates))
	env = upsertEnv(env, "ROUND_TABLE_KNOWLEDGE_ROOT", AbsPath(serverRoot, cfg.Knowledge.Root))
	env = upsertEnv(env, "ROUND_TABLE_KNOWLEDGE_TEMPLATES", AbsPath(serverRoot, cfg.Knowledge.Templates))
	env = upsertEnv(env, "ROUND_TABLE_DISCORD_BINDINGS_FILE", AbsPath(serverRoot, cfg.Transport.Discord.BindingsFile))
	return env
}

func upsertEnv(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	found := false
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			if !found {
				out = append(out, prefix+value)
				found = true
			}
			continue
		}
		out = append(out, e)
	}
	if !found {
		out = append(out, prefix+value)
	}
	return out
}
