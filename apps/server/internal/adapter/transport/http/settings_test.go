package httptransport

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/storage/sqlite"
	"round_table/apps/server/internal/platform/config"
)

func TestHandleSettings(t *testing.T) {
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}

	h, err := NewHandler(svc.Current(), nil, nil, svc, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"source":"app_settings"`) {
		t.Fatalf("GET status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"meet_presets"`) {
		t.Fatalf("GET missing meet_presets: %s", rec.Body.String())
	}

	put := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(`{"values":{"DEEPSEEK_MODEL_NAME":"deepseek-v4-flash"}}`))
	put.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, put)
	if rec2.Code != http.StatusOK {
		t.Fatalf("PUT status=%d body=%s", rec2.Code, rec2.Body.String())
	}
	if got := svc.Current().Model.DefaultModel; got != "deepseek-v4-flash" {
		t.Fatalf("in-memory model = %q", got)
	}

	ctx := put.Context()
	all, err := st.GetAllSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if all["DEEPSEEK_MODEL_NAME"] != "deepseek-v4-flash" {
		t.Fatalf("db settings = %+v", all)
	}
}

func TestHandleDiscordBots(t *testing.T) {
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "discord-bots.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}

	h, err := NewHandler(svc.Current(), nil, nil, svc, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	put := httptest.NewRequest(http.MethodPut, "/api/settings/discord-bots", strings.NewReader(`{
		"moderator_token": "tok-mod",
		"participants": [
			{"application_id":"1519615970128171068","token":"tok-analyst","bound_participant_id":"designer"}
		]
	}`))
	put.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, put)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT discord-bots status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := svc.Current().Transport.Discord.ParticipantBots; got != "1519615970128171068" {
		t.Fatalf("participant_bots = %q", got)
	}
	if got := svc.Current().Secrets.DiscordBotToken; got != "tok-mod" {
		t.Fatalf("moderator token = %q", got)
	}
	if got := svc.Current().Secrets.DiscordParticipantTokens["1519615970128171068"]; got != "tok-analyst" {
		t.Fatalf("analyst token = %+v", svc.Current().Secrets.DiscordParticipantTokens)
	}
	if !strings.Contains(rec.Body.String(), `"id":"moderator"`) {
		t.Fatalf("response missing moderator: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"token":"tok-mod"`) {
		t.Fatalf("response missing moderator token: %s", rec.Body.String())
	}

	all, err := st.GetAllSettings(put.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(all[config.DiscordBotTokensSetting], "tok-analyst") {
		t.Fatalf("tokens not in sqlite: %+v", all)
	}
}

func TestHandleMeetPresets(t *testing.T) {
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "meet-presets.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}

	h, err := NewHandler(svc.Current(), nil, nil, svc, nil)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)

	put := httptest.NewRequest(http.MethodPut, "/api/settings/meet-presets", strings.NewReader(`{
		"presets": [
			{"id":"1","name_zh":"⚡ 直接开始（默认）","mode":"decision","max_rounds":4,"confirmation":"skip","free_dialogue_questions":0},
			{"id":"2","name_zh":"🌩️ 闪电研讨","mode":"deliberation","max_rounds":1,"confirmation":"skip","free_dialogue_questions":0},
			{"id":"3","name_zh":"📐 标准研讨","mode":"deliberation","max_rounds":3,"confirmation":"skip","free_dialogue_questions":0},
			{"id":"4","name_zh":"💬 研讨 + 自由对话","mode":"deliberation","max_rounds":3,"confirmation":"skip","free_dialogue_questions":1},
			{"id":"5","name_zh":"✅ 研讨 + 需确认","mode":"deliberation","max_rounds":3,"confirmation":"required","free_dialogue_questions":1},
			{"id":"6","name_zh":"🔬 深度研讨","mode":"deliberation","max_rounds":5,"confirmation":"required","free_dialogue_questions":1},
			{"id":"J1","name_zh":"🌩️ 闪电裁决","mode":"decision","max_rounds":1,"confirmation":"skip","free_dialogue_questions":0},
			{"id":"J2","name_zh":"⚡ 快速裁决","mode":"decision","max_rounds":2,"confirmation":"skip","free_dialogue_questions":0},
			{"id":"J3","name_zh":"📋 标准裁决","mode":"decision","max_rounds":3,"confirmation":"skip","free_dialogue_questions":0},
			{"id":"J4","name_zh":"✅ 裁决 + 需确认","mode":"decision","max_rounds":3,"confirmation":"required","free_dialogue_questions":1},
			{"id":"J5","name_zh":"🔬 深度裁决","mode":"decision","max_rounds":5,"confirmation":"required","free_dialogue_questions":1}
		]
	}`))
	put.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, put)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT meet-presets status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := svc.Current().Meeting.MeetPresets[0].Mode; got != "decision" {
		t.Fatalf("preset 1 mode = %q", got)
	}
	if got := svc.Current().Meeting.MeetPresets[0].MaxRounds; got != 4 {
		t.Fatalf("preset 1 rounds = %d", got)
	}

	all, err := st.GetAllSettings(put.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(all[config.MeetPresetsSetting], `"decision"`) {
		t.Fatalf("db presets = %+v", all)
	}
}
