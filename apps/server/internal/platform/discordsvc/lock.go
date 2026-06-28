package discordsvc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"round_table/apps/server/internal/platform/config"
)

const lockWaitAfterSignal = 200 * time.Millisecond

// TransportLockPath returns the flock file path for the discord transport singleton.
func TransportLockPath(cfg config.Config) string {
	return config.DiscordTransportPIDPath(cfg) + ".lock"
}

// AcquireTransportLock claims the cross-process singleton lock for discord transport.
func AcquireTransportLock(path string) (*os.File, error) {
	return acquireTransportLock(path, true)
}

func acquireTransportLock(path string, retry bool) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if retry {
			if recovered, recoverErr := tryRecoverStaleTransportLock(path); recovered {
				return acquireTransportLock(path, false)
			} else if recoverErr != nil {
				return nil, recoverErr
			}
		}
		if holder := describeLockHolder(path); holder != "" {
			return nil, fmt.Errorf("another discord transport is already running (%s)", holder)
		}
		return nil, fmt.Errorf("another discord transport is already running")
	}
	_ = f.Truncate(0)
	_, _ = fmt.Fprintf(f, "%d\n", os.Getpid())
	return f, nil
}

// ReleaseTransportLock unlocks and closes the transport lock file.
func ReleaseTransportLock(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}

func terminateDiscordTransportLock(cfg config.Config) {
	lockPath := TransportLockPath(cfg)
	terminatePIDFile(lockPath)
	terminateLockHolderPIDs(lockPath)
	time.Sleep(lockWaitAfterSignal)
	terminateLockHolderPIDs(lockPath)
	if !lockFileInUse(lockPath) {
		_ = os.Remove(lockPath)
	}
}

func tryRecoverStaleTransportLock(path string) (bool, error) {
	if holder := readPIDFromFile(path); holder > 0 {
		if processAlive(holder) {
			return false, fmt.Errorf("another discord transport is already running (pid %d)", holder)
		}
		_ = os.Remove(path)
		return true, nil
	}
	pids := lockHolderPIDs(path)
	if len(pids) == 0 {
		_ = os.Remove(path)
		return true, nil
	}
	for _, pid := range pids {
		if processAlive(pid) {
			return false, fmt.Errorf("another discord transport is already running (pid %d)", pid)
		}
	}
	_ = os.Remove(path)
	return true, nil
}

func describeLockHolder(path string) string {
	if pid := readPIDFromFile(path); pid > 0 {
		return fmt.Sprintf("pid %d", pid)
	}
	pids := lockHolderPIDs(path)
	if len(pids) == 0 {
		return ""
	}
	parts := make([]string, 0, len(pids))
	for _, pid := range pids {
		parts = append(parts, strconv.Itoa(pid))
	}
	return "pid " + strings.Join(parts, ", ")
}

func readPIDFromFile(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0
	}
	return pid
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

func lockHolderPIDs(path string) []int {
	out, err := exec.Command("lsof", "-t", path).Output()
	if err != nil {
		return nil
	}
	seen := make(map[int]struct{})
	var pids []int
	for _, field := range strings.Fields(string(out)) {
		pid, err := strconv.Atoi(field)
		if err != nil || pid <= 0 {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		pids = append(pids, pid)
	}
	return pids
}

func terminateLockHolderPIDs(path string) {
	for _, pid := range lockHolderPIDs(path) {
		signalProcessGroup(pid, syscall.SIGTERM)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		alive := false
		for _, pid := range lockHolderPIDs(path) {
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
	for _, pid := range lockHolderPIDs(path) {
		if processAlive(pid) {
			signalProcessGroup(pid, syscall.SIGKILL)
		}
	}
}

func lockFileInUse(path string) bool {
	return len(lockHolderPIDs(path)) > 0
}
