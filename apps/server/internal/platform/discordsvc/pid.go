package discordsvc

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"round_table/apps/server/internal/platform/config"
)

func terminateDiscordTransportProcesses(cfg config.Config) {
	terminatePIDFile(config.DiscordTransportPIDPath(cfg))
	terminateDiscordTransportLock(cfg)
	terminateProcessesUsingFiles(
		logPath(cfg),
		TransportLockPath(cfg),
		config.DiscordInboundDedupDir(cfg),
	)
	for _, pattern := range []string{
		"apps/server/cmd/discord/main.go",
		"round_table/apps/server/cmd/discord",
		"tmp/roundtable-discord",
		"roundtable-discord",
	} {
		_ = exec.Command("pkill", "-f", pattern).Run()
	}
	if bin := discordBinaryPath(); bin != "" {
		_ = exec.Command("pkill", "-f", bin).Run()
	}
	if serverRoot, err := filepath.Abs(config.ServerRoot()); err == nil {
		_ = exec.Command("pkill", "-f", devDiscordBinaryPath(serverRoot)).Run()
	}
	time.Sleep(300 * time.Millisecond)
	terminateProcessesUsingFiles(
		logPath(cfg),
		TransportLockPath(cfg),
	)
	terminateDiscordTransportLock(cfg)
}

// CountDiscordTransportProcesses reports running roundtable-discord PIDs on the host.
func CountDiscordTransportProcesses() int {
	out, err := exec.Command("pgrep", "-f", "roundtable-discord").Output()
	if err != nil {
		return 0
	}
	return len(strings.Fields(string(out)))
}

func terminateProcessesUsingFiles(paths ...string) {
	seen := make(map[int]struct{})
	var pids []int
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		for _, pid := range lockHolderPIDs(path) {
			if _, ok := seen[pid]; ok {
				continue
			}
			seen[pid] = struct{}{}
			pids = append(pids, pid)
		}
	}
	for _, pid := range pids {
		signalProcessGroup(pid, syscall.SIGTERM)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		alive := false
		for _, pid := range pids {
			if processAlive(pid) {
				alive = true
				break
			}
		}
		if !alive {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	for _, pid := range pids {
		if processAlive(pid) {
			signalProcessGroup(pid, syscall.SIGKILL)
		}
	}
}

func discordBinaryPath() string {
	bin := config.AbsPath(config.RepoRoot(), "bin/roundtable-discord")
	if st, err := os.Stat(bin); err == nil && !st.IsDir() {
		return bin
	}
	if path, err := exec.LookPath("roundtable-discord"); err == nil {
		return path
	}
	return ""
}

func terminatePIDFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		_ = os.Remove(path)
		return
	}
	signalProcessGroup(pid, syscall.SIGTERM)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); err != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := syscall.Kill(pid, 0); err == nil {
		signalProcessGroup(pid, syscall.SIGKILL)
	}
	_ = os.Remove(path)
}
