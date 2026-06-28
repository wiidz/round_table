package discordsvc

import (
	"fmt"
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
	return buildDevDiscordBinary(serverRoot)
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
