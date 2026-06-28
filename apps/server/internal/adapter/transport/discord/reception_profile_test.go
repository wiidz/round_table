package discord

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/storage/sqlite"
	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

func TestParseProfileUpdateIntent(t *testing.T) {
	d, ok := parseProfileUpdateIntent("测试专家添加 soul")
	if !ok || d.ParticipantRef != "测试专家" || d.ProfileFile != "SOUL.md" {
		t.Fatalf("got=%+v ok=%v", d, ok)
	}
	if matchesCreateExpertIntent("测试专家添加 soul") {
		t.Fatal("should not match create expert")
	}
}

func TestParseProfileUpdateIntent_withContent(t *testing.T) {
	body := "测试专家 SOUL：\n# 测试人格\n负责纠错"
	d, ok := parseProfileUpdateIntent(body)
	if !ok || d.ParticipantRef != "测试专家" || !strings.Contains(d.ProfileContent, "测试人格") {
		t.Fatalf("got=%+v ok=%v", d, ok)
	}
}

func TestReception_profileUpdate_clarifyAndConfirm(t *testing.T) {
	st, err := sqlite.Open(t.TempDir() + "/rt.db")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_ = svc.CreateParticipant(ctx, config.ParticipantRosterItem{
		ID: "p_test", DisplayName: "测试专家", Expertise: "qa",
	})

	reg, _ := principalbind.NewRegistry(t.TempDir() + "/b.json")
	_, _ = reg.Bind(principalbind.ScopeKey("discord", "g1", "u1"), "discord", "u1", "Alice")

	tplRoot := t.TempDir()
	writeParticipantTemplates(t, tplRoot)
	admin := &ParticipantAdmin{
		ConfigSvc: svc,
		Profile:   fs.NewStore(t.TempDir(), tplRoot),
		Locale:    func() Locale { return LocaleZH },
	}
	meet := &MeetRunner{Registry: reg, ConfigSvc: svc, Discord: config.DiscordTransport{Locale: "zh"}}
	r := &Reception{Enabled: true, Registry: reg, Meet: meet, Participants: admin, Locale: func() Locale { return LocaleZH }}

	msg := transport.Inbound{Platform: "discord", GuildID: "g1", ChannelID: "ch1", AuthorID: "u1", Content: "测试专家添加 soul"}
	reply, err := r.TryHandle(ctx, msg)
	if err != nil || !strings.Contains(reply, "SOUL") || !strings.Contains(reply, "方向") {
		t.Fatalf("ask direction: reply=%q err=%v", reply, err)
	}

	msg.Content = "# 测试 SOUL\n负责游戏测试"
	reply, err = r.HandleClarifyFollowUp(ctx, msg)
	if err != nil || !strings.Contains(reply, "确认") {
		t.Fatalf("confirm preview paste: reply=%q err=%v", reply, err)
	}

	msg.Content = "1"
	reply, err = r.HandleConfirmReply(ctx, msg)
	if err != nil || !strings.Contains(reply, "已更新") {
		t.Fatalf("write: reply=%q err=%v", reply, err)
	}
	data, err := admin.Profile.ReadParticipant("p_test", "SOUL.md")
	if err != nil || !strings.Contains(string(data), "测试 SOUL") {
		t.Fatalf("file=%q err=%v", data, err)
	}
}

func TestReception_profileUpdate_llmDraft(t *testing.T) {
	st, err := sqlite.Open(t.TempDir() + "/rt.db")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, _ := config.NewService(st)
	ctx := context.Background()
	_ = svc.CreateParticipant(ctx, config.ParticipantRosterItem{
		ID: "p_test", DisplayName: "测试专家", Expertise: "qa",
	})
	reg, _ := principalbind.NewRegistry(t.TempDir() + "/b.json")
	_, _ = reg.Bind(principalbind.ScopeKey("discord", "g1", "u1"), "discord", "u1", "Alice")
	tplRoot := t.TempDir()
	writeParticipantTemplates(t, tplRoot)
	admin := &ParticipantAdmin{
		ConfigSvc: svc,
		Profile:   fs.NewStore(t.TempDir(), tplRoot),
		Locale:    func() Locale { return LocaleZH },
	}
	meet := &MeetRunner{Registry: reg, ConfigSvc: svc, Discord: config.DiscordTransport{Locale: "zh"}}
	r := &Reception{
		Enabled: true,
		Model: fakeReceptionModel{content: "# SOUL\n\n## 语气\n- 严谨找 bug\n\n## 边界\n- 只谈测试"},
		Registry: reg, Meet: meet, Participants: admin,
		Locale: func() Locale { return LocaleZH },
	}
	msg := transport.Inbound{Platform: "discord", GuildID: "g1", ChannelID: "ch1", AuthorID: "u1", Content: "测试专家添加 soul"}
	_, _ = r.TryHandle(ctx, msg)
	msg.Content = "游戏测试专家，擅长复现和写 bug 报告"
	reply, err := r.HandleClarifyFollowUp(ctx, msg)
	if err != nil || !strings.Contains(reply, "AI 根据你的描述生成") || !strings.Contains(reply, "严谨找 bug") {
		t.Fatalf("draft preview: reply=%q err=%v", reply, err)
	}
}

func TestLooksLikeProfileMarkdown(t *testing.T) {
	if !looksLikeProfileMarkdown("# SOUL\n\n## 语气") {
		t.Fatal("markdown")
	}
	if looksLikeProfileMarkdown("游戏测试专家，擅长找 bug") {
		t.Fatal("brief should not be markdown")
	}
}

func TestNormalizeProfileFile(t *testing.T) {
	if normalizeProfileFile("soul") != "SOUL.md" {
		t.Fatal("soul")
	}
	if normalizeProfileFile("AGENTS") != "AGENTS.md" {
		t.Fatal("agents")
	}
}
