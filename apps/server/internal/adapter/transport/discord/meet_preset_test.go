package discord

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/meeting"
)

func TestNormalizePresetChoice(t *testing.T) {
	if got := normalizePresetChoice("开始"); got != "1" {
		t.Fatalf("got=%q", got)
	}
	if got := normalizePresetChoice("j1"); got != "J1" {
		t.Fatalf("got=%q", got)
	}
	if got := normalizePresetChoice("J 1"); got != "J1" {
		t.Fatalf("got=%q", got)
	}
	if got := normalizePresetChoice("Ｊ１"); got != "J1" {
		t.Fatalf("got=%q", got)
	}
}

func TestHandlePresetMenu_decisionJ1(t *testing.T) {
	defaultCfg := meetLaunchConfig{
		Topic: "测试", Mode: meeting.MeetingModeDeliberation, MaxRounds: 2,
		MinRoundsBeforeSynthesis: 2, Confirmation: meeting.ConfirmationModeSkip,
	}
	sess := meetSetupSession{config: defaultCfg, step: setupStepPresetMenu}
	got, err := handlePresetMenu(sess, "J1", LocaleZH, "!rt ", defaultCfg)
	if err != nil || !got.launch {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.config.Mode != meeting.MeetingModeDecision || got.config.MaxRounds != 1 {
		t.Fatalf("cfg=%+v", got.config)
	}
}

func TestHandlePresetMenu_decisionJ5(t *testing.T) {
	defaultCfg := meetLaunchConfig{Topic: "测试", Mode: meeting.MeetingModeDecision, MaxRounds: 2}
	sess := meetSetupSession{config: defaultCfg, step: setupStepPresetMenu}
	got, err := handlePresetMenu(sess, "J5", LocaleZH, "!rt ", defaultCfg)
	if err != nil || !got.launch || got.config.MaxRounds != 5 || got.config.Confirmation != meeting.ConfirmationModeRequired {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.config.FreeDialogueQuestions != 1 {
		t.Fatalf("deep decision should include free dialogue, got=%d", got.config.FreeDialogueQuestions)
	}
}

func TestHandlePresetMenu_deepDeliberationFreeDialogue(t *testing.T) {
	defaultCfg := meetLaunchConfig{Topic: "测试", Mode: meeting.MeetingModeDeliberation, MaxRounds: 2}
	sess := meetSetupSession{config: defaultCfg, step: setupStepPresetMenu}
	got, err := handlePresetMenu(sess, "6", LocaleZH, "!rt ", defaultCfg)
	if err != nil || !got.launch || got.config.FreeDialogueQuestions != 1 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestFormatModeratorSetupPrompt_sections(t *testing.T) {
	got := formatModeratorSetupPrompt(LocaleZH, "!rt ", meetLaunchConfig{
		Topic: "测试会议", Mode: meeting.MeetingModeDeliberation, MaxRounds: 2,
		MinRoundsBeforeSynthesis: 2, Confirmation: meeting.ConfirmationModeSkip,
	})
	for _, want := range []string{"研讨", "裁决", "直接开始（默认）", "无自由对话", "**J1**", "**J5**", "闪电裁决", "━━", "**0**", "取消会议"} {
		if strings.Contains(got, "研讨型") || strings.Contains(got, "裁决型") {
			t.Fatalf("should not use 型 suffix in:\n%s", got)
		}
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}
