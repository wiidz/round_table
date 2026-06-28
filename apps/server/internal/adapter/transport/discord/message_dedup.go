package discord

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const messageDedupRetention = 24 * time.Hour

// ClaimInboundMessage marks a Discord message as handled across transport processes.
func ClaimInboundMessage(dir, messageID string) bool {
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return true
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	path := filepath.Join(dir, messageID)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// PruneInboundMessageClaims removes stale claim files older than messageDedupRetention.
func PruneInboundMessageClaims(dir string) {
	pruneInboundMessageClaims(dir)
}

func pruneInboundMessageClaims(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-messageDedupRetention)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}
