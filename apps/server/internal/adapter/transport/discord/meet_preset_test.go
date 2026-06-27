package discord

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/platform/config"
)

func TestNormalizePresetChoice(t *testing.T) {
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

func TestLookupPreset_byCommand(t *testing.T) {
	all := testMeetPresets(config.Config{})
	if _, ok := lookupPreset("1", all); !ok {
		t.Fatal("1 should match preset 1")
	}
	if _, ok := lookupPreset("j1", all); !ok {
		t.Fatal("j1 should match preset J1")
	}
}

func TestHandlePresetMenu_decisionJ1(t *testing.T) {
	all := testMeetPresets(config.Config{})
	defaultCfg := meetLaunchConfig{Topic: "测试"}
	sess := meetSetupSession{config: defaultCfg, step: setupStepPresetMenu}
	got, err := handlePresetMenu(sess, "J1", LocaleZH, "!rt ", all)
	if err != nil || !got.launch {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.config.Mode != meeting.MeetingModeDecision || got.config.MaxRounds != 1 {
		t.Fatalf("cfg=%+v", got.config)
	}
}

func TestHandlePresetMenu_decisionJ5(t *testing.T) {
	all := testMeetPresets(config.Config{})
	defaultCfg := meetLaunchConfig{Topic: "测试"}
	sess := meetSetupSession{config: defaultCfg, step: setupStepPresetMenu}
	got, err := handlePresetMenu(sess, "J5", LocaleZH, "!rt ", all)
	if err != nil || !got.launch || got.config.MaxRounds != 5 || got.config.Confirmation != meeting.ConfirmationModeRequired {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.config.FreeDialogueQuestions != 1 {
		t.Fatalf("deep decision should include free dialogue, got=%d", got.config.FreeDialogueQuestions)
	}
}

func TestHandlePresetMenu_deepDeliberationFreeDialogue(t *testing.T) {
	all := testMeetPresets(config.Config{})
	defaultCfg := meetLaunchConfig{Topic: "测试"}
	sess := meetSetupSession{config: defaultCfg, step: setupStepPresetMenu}
	got, err := handlePresetMenu(sess, "6", LocaleZH, "!rt ", all)
	if err != nil || !got.launch || got.config.FreeDialogueQuestions != 1 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestFormatModeratorSetupPrompt_sections(t *testing.T) {
	all := testMeetPresets(config.Config{})
	got := formatModeratorSetupPrompt(LocaleZH, "!rt ", all)
	for _, want := range []string{"研讨", "裁决", "直接开始（默认）", "无自由对话", "**J1**", "**J5**", "闪电裁决", "━━", "**0**", "取消会议"} {
		if strings.Contains(got, "研讨型") || strings.Contains(got, "裁决型") {
			t.Fatalf("should not use 型 suffix in:\n%s", got)
		}
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}
