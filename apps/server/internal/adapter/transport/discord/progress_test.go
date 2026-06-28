package discord

import (
	"strings"
	"testing"
)

func TestMergeMeetingStartProgress(t *testing.T) {
	engine := "▶ engine run started status=Preparing meeting=mtg-dc-1"
	pre := "▶ pre-meeting started order=designer → player"
	got := mergeMeetingStartProgress(engine, pre, LocaleZH)
	if !strings.Contains(got, "会议运行中") || !strings.Contains(got, "会前准备开始") {
		t.Fatalf("merged=%q", got)
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("expected two lines in one message, got %q", got)
	}
}

func TestFormatMeetLaunchAck_includesBrief(t *testing.T) {
	cfg := meetLaunchConfig{
		Topic:                    "狂击技能设计",
		Mode:                     "deliberation",
		MaxRounds:                1,
		MinRoundsBeforeSynthesis: 1,
		Confirmation:             "skip",
		Brief: meetBrief{
			Goal:         "评估经典技能手感",
			AgendaTitles: []string{"技能循环", "PVP 表现"},
			InScope:      "端游原版机制",
		},
	}
	got := formatMeetLaunchAck(LocaleZH, "mtg-dc-1", cfg, "player_i96", 3)
	for _, want := range []string{"会议简报", "评估经典技能手感", "技能循环", "端游原版机制", "MEETING.md"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "已写入 MEETING.md") && !strings.Contains(got, "目标：") {
		t.Fatalf("should show brief body, got:\n%s", got)
	}
}

func TestFormatMeetLaunchAck_includesParticipants(t *testing.T) {
	cfg := meetLaunchConfig{
		Topic:                    "骑士职业怎么设计",
		Mode:                     "deliberation",
		MaxRounds:                1,
		MinRoundsBeforeSynthesis: 1,
		Confirmation:             "skip",
		FreeDialogueQuestions:    0,
		ParticipantsSummary:      "designer" + participantSummarySep + "游戏策划, player" + participantSummarySep + "RO 老玩家代表",
	}
	got := formatMeetLaunchAck(LocaleZH, "mtg-dc-1", cfg, "player_i96", 3)
	if !strings.Contains(got, "👥 参会：游戏策划、RO 老玩家代表") {
		t.Fatalf("missing participants line:\n%s", got)
	}
}

func TestChannelProgress_coalescesMeetingStart(t *testing.T) {
	sender := &captureSender{}
	p := &channelProgress{
		pool:      &BotPool{Default: sender},
		channelID: "ch1",
		loc:       LocaleZH,
	}
	p.Logf("▶ engine run started status=Preparing meeting=mtg-dc-1")
	if len(sender.messages) != 0 {
		t.Fatalf("engine start should be held, sent=%v", sender.messages)
	}
	p.Logf("▶ pre-meeting started order=designer → player")
	if len(sender.messages) != 1 {
		t.Fatalf("expected one merged send, got %d: %v", len(sender.messages), sender.messages)
	}
	if !strings.Contains(sender.messages[0], "会议运行中") || !strings.Contains(sender.messages[0], "会前准备开始") {
		t.Fatalf("merged content=%q", sender.messages[0])
	}
}
