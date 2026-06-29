package discordsvc

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func devDiscordBinaryPath(serverRoot string) string {
	return filepath.Join(serverRoot, "tmp", "roundtable-discord")
}

// ResolveDiscordBinary returns the discord transport executable path, building a dev binary when needed.
func ResolveDiscordBinary(serverRoot string) (string, error) {
	if bin := discordBinaryPath(); bin != "" {
		return bin, nil
	}
	if bin, err := exec.LookPath("roundtable-discord"); err == nil {
		return bin, nil
	}
	out := devDiscordBinaryPath(serverRoot)
	if devDiscordBinaryStale(serverRoot, out) {
		return buildDevDiscordBinary(serverRoot)
	}
	return out, nil
}

func devDiscordBinaryStale(serverRoot, out string) bool {
	st, err := os.Stat(out)
	if err != nil {
		return true
	}
	binTime := st.ModTime()
	mainGo := filepath.Join(serverRoot, "cmd", "discord", "main.go")
	if fi, err := os.Stat(mainGo); err == nil && fi.ModTime().After(binTime) {
		return true
	}
	discordPkg := filepath.Join(serverRoot, "internal", "adapter", "transport", "discord")
	var stale bool
	_ = filepath.WalkDir(discordPkg, func(path string, d fs.DirEntry, err error) error {
		if err != nil || stale {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fi, err := d.Info()
		if err == nil && fi.ModTime().After(binTime) {
			stale = true
		}
		return nil
	})
	return stale
}

func buildDevDiscordBinary(serverRoot string) (string, error) {
	mainGo := filepath.Join(serverRoot, "cmd", "discord", "main.go")
	if _, err := os.Stat(mainGo); err != nil {
		return "", fmt.Errorf("discord entrypoint not found: %s", mainGo)
	}
	out := devDiscordBinaryPath(serverRoot)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return "", fmt.Errorf("prepare discord binary dir: %w", err)
	}
	cmd := exec.Command("go", "build", "-o", out, mainGo)
	cmd.Dir = serverRoot
	cmd.Env = withDiscordRunEnv(os.Environ())
	if buildOut, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build discord transport: %w: %s", err, strings.TrimSpace(string(buildOut)))
	}
	return out, nil
}
