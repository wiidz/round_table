package discordsvc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"round_table/apps/server/internal/platform/config"
)

// Status is the Discord transport child process state (managed by this HTTP server).
type Status struct {
	Running   bool   `json:"running"`
	Phase     string `json:"phase"` // stopped | starting | ready
	PID       int    `json:"pid,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
	ReadyAt   string `json:"ready_at,omitempty"`
	LastExit  string `json:"last_exit,omitempty"`
	LogPath   string `json:"log_path,omitempty"`
}

// Supervisor starts and stops the Discord transport (`go run apps/server/cmd/discord`).
type Supervisor struct {
	mu        sync.Mutex
	cmd       *exec.Cmd
	startedAt time.Time
	logPath   string
	lastExit  string
}

func (s *Supervisor) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.statusLocked()
}

func (s *Supervisor) statusLocked() Status {
	out := Status{Phase: "stopped", LogPath: s.logPath, LastExit: s.lastExit}
	if s.cmd == nil || s.cmd.Process == nil {
		return out
	}
	if err := s.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		s.cmd = nil
		return out
	}
	out.Running = true
	out.PID = s.cmd.Process.Pid
	if !s.startedAt.IsZero() {
		out.StartedAt = s.startedAt.UTC().Format(time.RFC3339)
	}
	path := s.logPath
	if phase, readyAt := detectSessionPhase(path); out.Running {
		out.Phase = phase
		if readyAt != "" {
			out.ReadyAt = readyAt
		}
	}
	return out
}

func (s *Supervisor) Start(_ context.Context, cfg config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.statusLocked().Running {
		return fmt.Errorf("discord transport already running")
	}

	serverRoot, err := filepath.Abs(config.ServerRoot())
	if err != nil {
		return fmt.Errorf("resolve server root: %w", err)
	}

	var cmd *exec.Cmd
	if bin, lookErr := exec.LookPath("roundtable-discord"); lookErr == nil {
		cmd = exec.Command(bin)
		cmd.Dir = serverRoot
	} else {
		mainGo := filepath.Join(serverRoot, "cmd", "discord", "main.go")
		if _, err := os.Stat(mainGo); err != nil {
			return fmt.Errorf("discord entrypoint not found (no roundtable-discord binary and %s missing)", mainGo)
		}
		cmd = exec.Command("go", "run", mainGo)
		cmd.Dir = serverRoot
	}
	cmd.Env = withDiscordRunEnv(config.DiscordChildEnv(cfg))

	logFilePath := logPath(cfg)
	logFile, err := openLogFile(logFilePath)
	if err != nil {
		return fmt.Errorf("open discord log: %w", err)
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		_, _ = fmt.Fprintf(logFile, "[supervisor] start failed: %v\n", err)
		_ = logFile.Close()
		return fmt.Errorf("start discord transport: %w", err)
	}

	_, _ = fmt.Fprintf(logFile, "[supervisor] discord transport process started · pid=%d\n", cmd.Process.Pid)

	s.cmd = cmd
	s.logPath = logFilePath
	s.lastExit = ""
	s.startedAt = time.Now().UTC()
	go s.wait(cmd, logFile)
	return nil
}

func (s *Supervisor) wait(cmd *exec.Cmd, logFile *os.File) {
	err := cmd.Wait()
	if logFile != nil {
		if err != nil {
			_, _ = fmt.Fprintf(logFile, "\n[supervisor] process exited with error: %v\n", err)
		} else {
			_, _ = fmt.Fprintln(logFile, "\n[supervisor] process exited normally")
		}
		_ = logFile.Close()
	}
	s.mu.Lock()
	if s.cmd == cmd {
		s.cmd = nil
		if err != nil {
			s.lastExit = err.Error()
		} else {
			s.lastExit = ""
		}
	}
	s.mu.Unlock()
}

func (s *Supervisor) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("discord transport is not running")
	}
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("stop discord transport: %w", err)
	}
	return nil
}

func (s *Supervisor) Logs(cfg config.Config, maxLines int) (Logs, error) {
	return s.readLogs(cfg, maxLines)
}

func withDiscordRunEnv(base []string) []string {
	env := append([]string{}, base...)
	if !hasEnvKey(env, "GOPROXY") {
		env = append(env, "GOPROXY=https://goproxy.cn,direct")
	}
	return ensureProxyDefaults(env)
}

func ensureProxyDefaults(env []string) []string {
	defaults := map[string]string{
		"https_proxy": "http://127.0.0.1:7897",
		"http_proxy":  "http://127.0.0.1:7897",
		"all_proxy":   "socks5://127.0.0.1:7897",
	}
	for key, val := range defaults {
		if !hasEnvKey(env, key) {
			env = append(env, key+"="+val)
		}
	}
	return env
}

func hasEnvKey(env []string, key string) bool {
	prefix := key + "="
	for _, e := range env {
		if len(e) >= len(prefix) && e[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
