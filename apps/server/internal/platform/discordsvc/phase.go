package discordsvc

import (
	"os"
	"strings"

	discordtransport "round_table/apps/server/internal/adapter/transport/discord"
)

func detectSessionPhase(logPath string) (phase string, readyAt string) {
	phase = "starting"
	if logPath == "" {
		return phase, ""
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		return phase, ""
	}
	segment := currentSessionSegment(string(data))
	if idx := strings.Index(segment, discordtransport.ReadyLogMarker); idx >= 0 {
		phase = "ready"
		readyAt = parseReadyAt(segment[idx:])
	}
	return phase, readyAt
}

func currentSessionSegment(text string) string {
	if sessionStart := lastSessionStartLine(text); sessionStart >= 0 {
		return text[sessionStart:]
	}
	const spawn = "[supervisor] discord transport process started"
	if idx := strings.LastIndex(text, spawn); idx >= 0 {
		return text[idx:]
	}
	return ""
}

func lastSessionStartLine(text string) int {
	const marker = "======== Discord transport · session"
	idx := -1
	for {
		next := strings.Index(text[idx+1:], marker)
		if next < 0 {
			break
		}
		idx = idx + 1 + next
	}
	if idx >= 0 {
		return idx
	}
	// Legacy sessions before readable log format.
	const legacy = "--- session started "
	idx = -1
	for {
		next := strings.Index(text[idx+1:], legacy)
		if next < 0 {
			break
		}
		idx = idx + 1 + next
	}
	return idx
}

func parseReadyAt(line string) string {
	line = strings.TrimSpace(line)
	prefix := discordtransport.ReadyLogMarker
	if !strings.HasPrefix(line, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(line, prefix))
}
