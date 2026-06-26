package discord

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/platform/config"
)

func TestDefaultLaunchConfig(t *testing.T) {
	r := &MeetRunner{
		Cfg: config.Config{
			Meeting: config.Meeting{
				MaxRoundsPerSegment:      5,
				MinRoundsBeforeSynthesis: 2,
			},
		},
		Discord: config.DiscordTransport{
			MeetMode:                  meeting.MeetingModeDeliberation,
			MeetMaxRounds:             2,
			MeetConfirmation:          meeting.ConfirmationModeSkip,
			MeetFreeDialogueQuestions: 0,
			MeetParticipants:          "designer:游戏策划:gameplay,player:玩家代表:experience",
		},
	}
	cfg := r.defaultLaunchConfig("测试会议", "")
	if cfg.Topic != "测试会议" || cfg.Mode != meeting.MeetingModeDeliberation {
		t.Fatalf("cfg=%+v", cfg)
	}
	if cfg.MaxRounds != 2 || cfg.MinRoundsBeforeSynthesis != 2 || cfg.FreeDialogueQuestions != 0 {
		t.Fatalf("rounds/free = %+v", cfg)
	}
	if !strings.Contains(cfg.ParticipantsSummary, "designer·游戏策划") {
		t.Fatalf("participants=%q", cfg.ParticipantsSummary)
	}
}

func TestNormalizeSetupChoice(t *testing.T) {
	if got := normalizeSetupChoice("开始"); got != "1" {
		t.Fatalf("got=%q", got)
	}
	if got := normalizeSetupChoice("２"); got != "2" {
		t.Fatalf("got=%q", got)
	}
}

func TestHandlePresetMenu_startDefault(t *testing.T) {
	defaultCfg := meetLaunchConfig{
		Topic: "测试", Mode: meeting.MeetingModeDeliberation, MaxRounds: 2,
		MinRoundsBeforeSynthesis: 2, Confirmation: meeting.ConfirmationModeSkip,
		FreeDialogueQuestions: 1,
	}
	sess := meetSetupSession{config: defaultCfg, step: setupStepPresetMenu}
	got, err := handlePresetMenu(sess, "1", LocaleZH, "!rt ", defaultCfg)
	if err != nil || !got.launch {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.config.FreeDialogueQuestions != 0 {
		t.Fatalf("default preset should disable free dialogue, got=%d", got.config.FreeDialogueQuestions)
	}
}

func TestHandlePresetMenu_flashDeliberation(t *testing.T) {
	defaultCfg := meetLaunchConfig{
		Topic: "测试", Mode: meeting.MeetingModeDeliberation, MaxRounds: 2,
		MinRoundsBeforeSynthesis: 2, Confirmation: meeting.ConfirmationModeSkip,
	}
	sess := meetSetupSession{config: defaultCfg, step: setupStepPresetMenu}
	got, err := handlePresetMenu(sess, "2", LocaleZH, "!rt ", defaultCfg)
	if err != nil || !got.launch {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.config.MaxRounds != 1 || got.config.MinRoundsBeforeSynthesis != 1 {
		t.Fatalf("flash preset = %+v", got.config)
	}
	if got.config.Mode != meeting.MeetingModeDeliberation {
		t.Fatalf("mode=%q", got.config.Mode)
	}
}

func TestHandlePresetMenu_customWizard(t *testing.T) {
	defaultCfg := meetLaunchConfig{Topic: "测试", Mode: meeting.MeetingModeDeliberation, MaxRounds: 2}
	sess := meetSetupSession{config: defaultCfg, step: setupStepPresetMenu}
	got, err := handlePresetMenu(sess, "0", LocaleZH, "!rt ", defaultCfg)
	if err != nil || got.launch || got.step != setupStepCustomMode {
		t.Fatalf("got=%+v err=%v", got, err)
	}

	sess = meetSetupSession{config: defaultCfg, step: setupStepCustomMode}
	got, err = handleCustomMode(sess, "1", LocaleZH, "!rt ", defaultCfg)
	if err != nil || got.step != setupStepCustomRounds {
		t.Fatalf("mode step: got=%+v err=%v", got, err)
	}

	sess = meetSetupSession{config: got.config, step: setupStepCustomRounds}
	got, err = handleCustomRounds(sess, "3", LocaleZH, "!rt ", defaultCfg)
	if err != nil || got.config.MaxRounds != 3 {
		t.Fatalf("rounds: got=%+v err=%v", got, err)
	}

	sess = meetSetupSession{config: got.config, step: setupStepCustomConfirmation}
	got, err = handleCustomConfirmation(sess, "1", LocaleZH, "!rt ", defaultCfg)
	if err != nil {
		t.Fatal(err)
	}

	sess = meetSetupSession{config: got.config, step: setupStepCustomFree}
	got, err = handleCustomFree(sess, "1", LocaleZH, "!rt ", defaultCfg)
	if err != nil || got.step != setupStepCustomConfirm {
		t.Fatalf("free: got=%+v err=%v", got, err)
	}

	sess = meetSetupSession{config: got.config, step: setupStepCustomConfirm}
	got, err = handleCustomConfirm(sess, "1", LocaleZH, "!rt ", defaultCfg)
	if err != nil || !got.launch {
		t.Fatalf("confirm: got=%+v err=%v", got, err)
	}
}

func TestFormatModeratorSetupPrompt(t *testing.T) {
	got := formatModeratorSetupPrompt(LocaleZH, "!rt ", meetLaunchConfig{
		Topic: "测试会议", Mode: meeting.MeetingModeDeliberation, MaxRounds: 2,
		MinRoundsBeforeSynthesis: 2, Confirmation: meeting.ConfirmationModeSkip,
	})
	for _, want := range []string{"测试会议", "研讨", "J1", "J5", "0"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestMeetSetupSessions(t *testing.T) {
	var s meetSetupSessions
	s.put("ch1", meetSetupSession{channelID: "ch1", authorID: "u1"})
	if !s.pending("ch1") {
		t.Fatal("expected pending")
	}
	s.clear("ch1")
	if s.pending("ch1") {
		t.Fatal("expected cleared")
	}
}
