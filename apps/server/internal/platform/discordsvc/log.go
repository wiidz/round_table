package discordsvc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"round_table/apps/server/internal/platform/config"
)

const defaultLogTailLines = 200

// Logs holds a tail of the discord transport log file.
type Logs struct {
	Path  string `json:"path"`
	Lines string `json:"lines"`
}

func logPath(cfg config.Config) string {
	serverRoot, _ := filepath.Abs(config.ServerRoot())
	sqlite := config.AbsPath(serverRoot, cfg.Storage.SQLitePath)
	return filepath.Join(filepath.Dir(sqlite), "logs", "discord-transport.log")
}

func openLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	_, _ = fmt.Fprintf(f, "\n======== Discord transport · session %s ========\n", time.Now().Format("2006-01-02 15:04:05 MST"))
	return f, nil
}

func tailLog(path string, maxLines int) (Logs, error) {
	out := Logs{Path: path}
	if maxLines <= 0 {
		maxLines = defaultLogTailLines
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return out, err
	}
	text := strings.TrimRight(string(data), "\n")
	if text == "" {
		return out, nil
	}
	lines := strings.Split(text, "\n")
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	out.Lines = strings.Join(lines, "\n")
	return out, nil
}

func (s *Supervisor) readLogs(cfg config.Config, maxLines int) (Logs, error) {
	s.mu.Lock()
	path := s.logPath
	if path == "" {
		path = logPath(cfg)
	}
	s.mu.Unlock()
	return tailLog(path, maxLines)
}

func (s *Supervisor) ClearLogs(cfg config.Config) error {
	s.mu.Lock()
	path := s.logPath
	if path == "" {
		path = logPath(cfg)
	}
	s.mu.Unlock()
	return clearLog(path)
}

func clearLog(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	marker := fmt.Sprintf("--- log cleared %s ---\n", time.Now().Format(time.RFC3339))
	return os.WriteFile(path, []byte(marker), 0o644)
}
