package discord

import (
	"strings"
	"testing"
	"time"

	"round_table/apps/server/internal/platform/config"
)

func TestStartupReadyLogLines_containsReadyMarker(t *testing.T) {
	h := NewCommandHandler("!rt", nil, &MeetRunner{Discord: config.DiscordTransport{Locale: "zh"}})
	startedAt := time.Date(2026, 6, 27, 20, 33, 4, 0, time.FixedZone("CST", 8*3600))
	lines := StartupReadyLogLines(h, StartupInfo{
		StartedAt:            startedAt,
		Prefix:               "!rt ",
		BindingsFile:         "./data/bindings.json",
		Locale:               LocaleZH,
		ModeratorUsername:    "mod_bot",
		ParticipantConnected: 2,
		ParticipantTotal:     4,
	})

	text := strings.Join(lines, "\n")
	if !strings.Contains(text, ReadyLogMarker) {
		t.Fatalf("ready marker missing:\n%s", text)
	}
	if !strings.Contains(text, "!rt help") {
		t.Fatalf("help command missing:\n%s", text)
	}
	if !strings.Contains(text, "新会议") {
		t.Fatalf("natural meet trigger missing:\n%s", text)
	}
}
