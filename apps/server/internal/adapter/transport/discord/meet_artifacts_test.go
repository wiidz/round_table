package discord

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/platform/config"
)

func TestPostMeetArtifacts(t *testing.T) {
	root := t.TempDir()
	meetingID := "mtg-test-1"
	dir := filepath.Join(root, meetingID, "artifacts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, meetingID, "MINUTES.md"), []byte("# Minutes\n\nround 1 done"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "design-draft.md"), []byte("# Draft\n\nskill tree v2"), 0o644); err != nil {
		t.Fatal(err)
	}

	sender := &captureSender{}
	r := &MeetRunner{
		Cfg:   config.Config{Workspace: config.Workspace{Root: root}},
		Bots:  &BotPool{Default: sender},
		Discord: config.DiscordTransport{Locale: "zh"},
	}
	final := meeting.State{
		ID:          meetingID,
		Status:      meeting.StatusCompleted,
		MeetingMode: meeting.MeetingModeDeliberation,
	}
	r.postMeetArtifacts(context.Background(), "ch1", final, meetingID, LocaleZH)

	if len(sender.messages) < 2 {
		t.Fatalf("messages=%v", sender.messages)
	}
	if !strings.Contains(sender.messages[0], "会议纪要") || !strings.Contains(sender.messages[1], "方案草案") {
		t.Fatalf("messages=%v", sender.messages)
	}
}
