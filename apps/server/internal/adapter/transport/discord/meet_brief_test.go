package discord

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/platform/config"
)

func TestParseAgendaLines(t *testing.T) {
	got := parseAgendaLines("1）职业定位\n2. 核心循环\n技能差异")
	want := []string{"职业定位", "核心循环", "技能差异"}
	if len(got) != len(want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] got=%q want=%q", i, got[i], want[i])
		}
	}
}

func TestFormatBriefForEngineGoal(t *testing.T) {
	b := meetBrief{
		Goal:         "输出技能框架草案",
		InScope:      "定位、循环",
		OutOfScope:   "数值表",
		DoneCriteria: "每议程至少 1 条结论",
	}
	got := formatBriefForEngineGoal(b)
	for _, want := range []string{"输出技能框架草案", "完成标准：", "讨论范围：", "不在范围："} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestParseAgendaLines_preservesCommasInTopic(t *testing.T) {
	got := parseAgendaLines("1）氪金方案（扭蛋、头饰、卡片）\n2）数值膨胀（+1% 双攻，冒险手册）")
	want := []string{
		"氪金方案（扭蛋、头饰、卡片）",
		"数值膨胀（+1% 双攻，冒险手册）",
	}
	if len(got) != len(want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] got=%q want=%q", i, got[i], want[i])
		}
	}
}

func TestAgendaTitlesToItems(t *testing.T) {
	items := agendaTitlesToItems([]string{"核心技能", "职业定位"})
	if len(items) != 2 || items[0].Title != "核心技能" || items[0].ID == "" {
		t.Fatalf("items=%+v", items)
	}
}

func TestHandlePresetMenuPreservesBrief(t *testing.T) {
	all := testMeetPresets(config.Config{})
	sess := meetSetupSession{
		step: setupStepPresetMenu,
		config: meetLaunchConfig{
			Topic: "骑士职业",
			Brief: meetBrief{
				Goal:         "输出草案",
				AgendaTitles: []string{"定位", "循环"},
			},
			ParticipantIDs: []string{"design"},
		},
	}
	got, err := handlePresetMenu(sess, "1", LocaleZH, "!rt ", all)
	if err != nil || !got.launch {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if got.config.Brief.Goal != "输出草案" || len(got.config.Brief.AgendaTitles) != 2 {
		t.Fatalf("brief not preserved: %+v", got.config.Brief)
	}
	if got.config.engineGoal() == "" || len(got.config.engineAgenda()) != 2 {
		t.Fatalf("engine mapping failed goal=%q agenda=%v", got.config.engineGoal(), got.config.engineAgenda())
	}
}

func TestFormatBriefSummaryBody_listsTopicsFully(t *testing.T) {
	body := formatBriefSummaryBody(LocaleZH, meetBrief{
		Goal: "输出草案",
		AgendaTitles: []string{
			"背景与约束梳理",
			"方案选项对比与取舍",
			"风险、依赖与后续行动",
		},
	})
	if strings.Contains(body, " · ") {
		t.Fatalf("topics should not be joined with middle dots: %q", body)
	}
	for _, want := range []string{"1）背景与约束", "2）方案选项", "3）风险、依赖"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %q", want, body)
		}
	}
}

func TestBriefWizardToPresetMenu(t *testing.T) {
	r := &MeetRunner{
		Cfg: config.Config{Meeting: config.Meeting{MeetPresets: config.DefaultMeetPresets(config.Config{})}},
		Discord: config.DiscordTransport{
			Locale:           "zh",
			MeetParticipants: "designer:策划:gameplay",
		},
	}
	sess := meetSetupSession{
		config: meetLaunchConfig{Topic: "测试", Mode: meeting.MeetingModeDeliberation},
		step:   setupStepBriefGoal,
	}
	sess, reply := r.advanceBriefGoal(sess, "输出方案草案", LocaleZH)
	if sess.step != setupStepBriefAgenda || !strings.Contains(reply, "讨论议题") {
		t.Fatalf("topics step: step=%v reply=%q", sess.step, reply)
	}
	sess, reply = r.advanceBriefAgenda(sess, "定位\n循环", LocaleZH)
	if sess.step != setupStepBriefScope || len(sess.config.Brief.AgendaTitles) != 2 || !strings.Contains(reply, "边界与完成标准") {
		t.Fatalf("scope step: step=%v agenda=%v reply=%q", sess.step, sess.config.Brief.AgendaTitles, reply)
	}
	sess, reply = r.advanceBriefScope(sess, "-", LocaleZH)
	if sess.step != setupStepPresetMenu || !strings.Contains(reply, "请选择会议方案") {
		t.Fatalf("preset menu: step=%v reply=%q", sess.step, reply)
	}
	if !strings.Contains(reply, "已记录简报") {
		t.Fatalf("expected brief summary in reply: %q", reply)
	}
}
