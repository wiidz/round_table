package discord

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/brief"
	brieffs "round_table/apps/server/internal/adapter/brief/fs"
	"round_table/apps/server/internal/adapter/transport"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/platform/config"
)

func TestParseMeetArgsWithTemplate(t *testing.T) {
	got, err := parseMeetArgs([]string{"-template", "decision-review", "新职业"}, "decision")
	if err != nil || got.TemplateID != "decision-review" || got.Topic != "新职业" {
		t.Fatalf("got=%+v err=%v", got, err)
	}

	got, err = parseMeetArgs([]string{"--template", "decision-review"}, "decision")
	if err != nil || got.TemplateID != "decision-review" || got.Topic != "" {
		t.Fatalf("template-only = %+v err=%v", got, err)
	}

	if _, err := parseMeetArgs(nil, "decision"); err == nil {
		t.Fatal("expected topic or template required")
	}
}

func TestLaunchDraftToMeetConfig(t *testing.T) {
	base := meetLaunchConfig{Topic: "CLI 主题", Mode: meeting.MeetingModeDecision}
	draft := brief.LaunchDraft{
		Topic: "模板主题",
		Brief: brief.BriefBody{
			Goal:       "形成共识",
			Agenda:     []string{"议题一"},
			OutOfScope: "排期",
		},
		Meeting: brief.MeetingDefaults{
			Mode:                     meeting.MeetingModeDeliberation,
			MaxRounds:                2,
			ConfirmationMode:         meeting.ConfirmationModeSkip,
			FreeDialogueMaxQuestions: 1,
		},
	}

	got := launchDraftToMeetConfig(base, draft)
	if got.Topic != "CLI 主题" {
		t.Fatalf("CLI topic wins: %q", got.Topic)
	}
	if got.Brief.Goal != "形成共识" || len(got.Brief.AgendaTitles) != 1 {
		t.Fatalf("brief not applied: %+v", got.Brief)
	}
	if got.Mode != meeting.MeetingModeDeliberation || got.MaxRounds != 2 {
		t.Fatalf("meeting defaults: mode=%s rounds=%d", got.Mode, got.MaxRounds)
	}
	if got.Confirmation != meeting.ConfirmationModeSkip || got.FreeDialogueQuestions != 1 {
		t.Fatalf("confirmation/free: %+v", got)
	}
}

func TestBeginSetupWithBriefTemplate(t *testing.T) {
	dir := t.TempDir()
	templates := filepath.Join(dir, "_templates", "briefs", "decision-review")
	if err := os.MkdirAll(templates, 0o755); err != nil {
		t.Fatal(err)
	}
	seed := `meta:
  title: 裁决型评审
  description: seed
topic: ""
brief:
  goal: 形成共识
  agenda:
    - 风险
meeting:
  mode: decision
  max_rounds: 3
  confirmation_mode: required
`
	if err := os.WriteFile(filepath.Join(templates, "BRIEF.yaml"), []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	r := &MeetRunner{
		Discord: config.DiscordTransport{Locale: "zh"},
		Briefs:  brieffs.NewStore(filepath.Join(dir, "briefs"), filepath.Join(dir, "_templates", "briefs")),
	}
	msg := transport.Inbound{Platform: "web", ChannelID: "ch1", AuthorID: "u1"}

	reply, err := r.BeginSetup(msg, meetParseResult{Topic: "平衡调整", TemplateID: "decision-review"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "已加载简报模板") || !strings.Contains(reply, "请选择参会专家") {
		t.Fatalf("reply=%q", reply)
	}

	sess, ok := r.setups.get("ch1")
	if !ok || sess.briefTemplateID != "decision-review" {
		t.Fatalf("session: %+v ok=%v", sess, ok)
	}
	if sess.config.Brief.Goal != "形成共识" || sess.step != setupStepPickParticipants {
		t.Fatalf("prefill: goal=%q step=%v", sess.config.Brief.Goal, sess.step)
	}
}

func TestResolveTemplateChoice(t *testing.T) {
	templates := []brief.TemplateIndex{
		{ID: "decision-review", Title: "裁决型评审"},
		{ID: "custom-a", Title: "自定义"},
	}
	id, err := resolveTemplateChoice("1", templates)
	if err != nil || id != "decision-review" {
		t.Fatalf("by number: id=%q err=%v", id, err)
	}
	id, err = resolveTemplateChoice("decision-review", templates)
	if err != nil || id != "decision-review" {
		t.Fatalf("by id: id=%q err=%v", id, err)
	}
	id, err = resolveTemplateChoice("0", templates)
	if err != nil || id != "" {
		t.Fatalf("skip with 0: id=%q err=%v", id, err)
	}
	id, err = resolveTemplateChoice("跳过", templates)
	if err != nil || id != "" {
		t.Fatalf("skip: id=%q err=%v", id, err)
	}
}
