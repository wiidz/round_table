package discord

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/storage/sqlite"
	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

type fakeReceptionModel struct {
	content string
}

func (f fakeReceptionModel) Complete(_ context.Context, _ model.Request) (model.Response, error) {
	return model.Response{Content: f.content}, nil
}

func TestParseReceptionDecision(t *testing.T) {
	raw := "```json\n{\"tool\":\"get_artifact\",\"artifact\":\"minutes\",\"message\":\"\"}\n```"
	got, err := parseReceptionDecision(raw)
	if err != nil || got.Tool != receptionToolGetArtifact || got.Artifact != "minutes" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestParseReceptionDecision_mutatingFields(t *testing.T) {
	raw := `{"tool":"create_participant","display_name":"玩家小美","expertise":"moba"}`
	got, err := parseReceptionDecision(raw)
	if err != nil || got.Tool != receptionToolCreateParticipant || got.DisplayName != "玩家小美" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestNormalizeArtifactKind(t *testing.T) {
	if normalizeArtifactKind("纪要") != "minutes" {
		t.Fatal("minutes")
	}
	if normalizeArtifactKind("设计草案") != "draft" {
		t.Fatal("draft")
	}
}

func TestReception_listParticipants(t *testing.T) {
	meet := &MeetRunner{
		Discord: config.DiscordTransport{
			Locale:           "zh",
			MeetParticipants: "designer:游戏策划:gameplay,player:玩家:experience",
		},
	}
	r := &Reception{
		Model:   fakeReceptionModel{content: `{"tool":"list_participants","artifact":"","message":""}`},
		Enabled: true,
		Meet:    meet,
		Participants: &ParticipantAdmin{
			Locale: func() Locale { return LocaleZH },
		},
		Phase: func(string) ChannelInputPhase { return InputPhaseIdle },
	}
	reply, err := r.TryHandle(context.Background(), transport.Inbound{
		ChannelID: "ch1", Content: "现在有哪些专家可以参加？",
	})
	if err != nil {
		t.Fatal(err)
	}
	if reply == "" {
		t.Fatal("expected list reply")
	}
}

func TestReception_meetingStatus(t *testing.T) {
	meet := &MeetRunner{Discord: config.DiscordTransport{Locale: "zh"}}
	r := &Reception{
		Model:   fakeReceptionModel{content: `{"tool":"meeting_status","artifact":"","message":""}`},
		Enabled: true,
		Meet:    meet,
		Phase:   func(string) ChannelInputPhase { return InputPhaseIdle },
		Locale:  func() Locale { return LocaleZH },
	}
	reply, err := r.TryHandle(context.Background(), transport.Inbound{
		ChannelID: "ch1", Content: "现在什么状态",
	})
	if err != nil || !strings.Contains(reply, "空闲") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func TestShouldSkipReception_naturalMeet(t *testing.T) {
	if !shouldSkipReception("开个会，策划、玩家一起，聊聊测试") {
		t.Fatal("should skip for natural meet fast path")
	}
}

func TestReception_createParticipant_confirmFlow(t *testing.T) {
	st, err := sqlite.Open(t.TempDir() + "/rt.db")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = reg.Bind(principalbind.ScopeKey("discord", "g1", "u1"), "discord", "u1", "Alice")

	tplRoot := t.TempDir()
	writeParticipantTemplates(t, tplRoot)
	admin := &ParticipantAdmin{
		ConfigSvc: svc,
		Profile:   fs.NewStore(t.TempDir(), tplRoot),
		Locale:    func() Locale { return LocaleZH },
	}
	meet := &MeetRunner{
		Registry: reg,
		ConfigSvc: svc,
		Discord: config.DiscordTransport{
			Locale:           "zh",
			MeetParticipants: "designer:游戏策划:gameplay",
		},
	}
	r := &Reception{
		Model: fakeReceptionModel{content: `{"tool":"create_participant","display_name":"玩家小美","expertise":"moba"}`},
		Enabled: true, Registry: reg, Meet: meet, Participants: admin,
		Locale: func() Locale { return LocaleZH },
	}
	msg := transport.Inbound{Platform: "discord", GuildID: "g1", ChannelID: "ch1", AuthorID: "u1", Content: "新增专家 玩家小美"}

	reply, err := r.TryHandle(context.Background(), msg)
	if err != nil || !strings.Contains(reply, "确认") || !strings.Contains(reply, "玩家小美") {
		t.Fatalf("preview=%q err=%v", reply, err)
	}
	if r.InputPhase("ch1") != InputPhaseReceptionConfirm {
		t.Fatal("expected reception confirm phase")
	}

	msg.Content = "1"
	reply, err = r.HandleConfirmReply(context.Background(), msg)
	if err != nil || !strings.Contains(reply, "已创建专家") {
		t.Fatalf("create=%q err=%v", reply, err)
	}
}

func TestReception_createParticipant_missingName_clarify(t *testing.T) {
	st, err := sqlite.Open(t.TempDir() + "/rt.db")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, _ := config.NewService(st)
	reg, _ := principalbind.NewRegistry(t.TempDir() + "/b.json")
	_, _ = reg.Bind(principalbind.ScopeKey("discord", "g1", "u1"), "discord", "u1", "Alice")
	admin := &ParticipantAdmin{ConfigSvc: svc, Locale: func() Locale { return LocaleZH }}
	meet := &MeetRunner{Registry: reg, ConfigSvc: svc, Discord: config.DiscordTransport{Locale: "zh"}}
	r := &Reception{
		Model: fakeReceptionModel{content: `{"tool":"create_participant","display_name":"","message":""}`},
		Enabled: true, Registry: reg, Meet: meet, Participants: admin,
		Locale: func() Locale { return LocaleZH },
	}
	reply, err := r.TryHandle(context.Background(), transport.Inbound{
		Platform: "discord", GuildID: "g1", ChannelID: "ch1", AuthorID: "u1", Content: "新增专家",
	})
	if err != nil || !strings.Contains(reply, "显示名") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func TestPrepareCreateItem_suggestID(t *testing.T) {
	admin := &ParticipantAdmin{
		ConfigSvc: nil,
		Locale:    func() Locale { return LocaleZH },
	}
	r := &Reception{Participants: admin}
	item, err := r.prepareCreateItem(receptionDecision{DisplayName: "玩家小美", Expertise: "moba"})
	if err != nil {
		t.Fatal(err)
	}
	if item.ID == "" || item.DisplayName != "玩家小美" {
		t.Fatalf("item=%+v", item)
	}
}

func TestSuggestParticipantID_chineseDisplayName(t *testing.T) {
	id := suggestParticipantID("测试专家", nil)
	if id == "" || id == "expert" {
		t.Fatalf("expected hash-based id, got %q", id)
	}
	if err := config.ValidateParticipantID(id); err != nil {
		t.Fatalf("invalid id %q: %v", id, err)
	}
}

func TestReception_clarifyFollowUp_create(t *testing.T) {
	st, err := sqlite.Open(t.TempDir() + "/rt.db")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}
	reg, err := principalbind.NewRegistry(t.TempDir() + "/b.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = reg.Bind(principalbind.ScopeKey("discord", "g1", "u1"), "discord", "u1", "Alice")
	admin := &ParticipantAdmin{ConfigSvc: svc, Locale: func() Locale { return LocaleZH }}
	meet := &MeetRunner{Registry: reg, ConfigSvc: svc, Discord: config.DiscordTransport{Locale: "zh"}}
	r := &Reception{
		Enabled: true, Registry: reg, Meet: meet, Participants: admin,
		Locale: func() Locale { return LocaleZH },
	}
	msg := transport.Inbound{Platform: "discord", GuildID: "g1", ChannelID: "ch1", AuthorID: "u1", Content: "新增专家"}
	r.storeClarifySession(msg, msg.Content, receptionDecision{
		Tool: receptionToolClarify, PendingTool: receptionToolCreateParticipant,
		Message: "请提供要新增的专家名称（显示名）",
	})
	if r.InputPhase("ch1") != InputPhaseReceptionClarify {
		t.Fatal("expected clarify phase")
	}
	msg.Content = "测试专家"
	reply, err := r.HandleClarifyFollowUp(context.Background(), msg)
	if err != nil || !strings.Contains(reply, "确认") || !strings.Contains(reply, "测试专家") {
		t.Fatalf("preview=%q err=%v", reply, err)
	}
}

func TestExtractCreateDisplayName(t *testing.T) {
	if got := extractCreateDisplayName("新增测试专家"); got != "测试专家" {
		t.Fatalf("got=%q", got)
	}
	if got := extractCreateDisplayName("新增专家 测试专家"); got != "测试专家" {
		t.Fatalf("got=%q", got)
	}
	if got := extractCreateDisplayName("QA Lead"); got != "QA Lead" {
		t.Fatalf("got=%q", got)
	}
}

func TestInferReceptionPendingTool(t *testing.T) {
	if inferReceptionPendingTool("新增专家") != receptionToolCreateParticipant {
		t.Fatal("create")
	}
}
