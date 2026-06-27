package config_test

import (
	"context"
	"path/filepath"
	"testing"

	"round_table/apps/server/internal/adapter/storage/sqlite"
	"round_table/apps/server/internal/platform/config"
)

type memorySettings struct {
	data map[string]string
}

func (m *memorySettings) GetAllSettings(_ context.Context) (map[string]string, error) {
	out := make(map[string]string, len(m.data))
	for k, v := range m.data {
		out[k] = v
	}
	return out, nil
}

func (m *memorySettings) SetSettings(_ context.Context, updates map[string]string) error {
	for k, v := range updates {
		m.data[k] = v
	}
	return nil
}

func TestService_UpdateDiscordAutoStart(t *testing.T) {
	store := &memorySettings{data: map[string]string{}}
	svc, err := config.NewService(store)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := svc.Update(ctx, map[string]string{
		"ROUND_TABLE_DISCORD_AUTO_START": "true",
	}); err != nil {
		t.Fatal(err)
	}
	if !svc.Current().Transport.Discord.AutoStart {
		t.Fatal("AutoStart should be true")
	}
	if err := svc.Update(ctx, map[string]string{
		"ROUND_TABLE_DISCORD_AUTO_START": "false",
	}); err != nil {
		t.Fatal(err)
	}
	if svc.Current().Transport.Discord.AutoStart {
		t.Fatal("AutoStart should be false")
	}
}

func TestService_UpdateAppliesInMemory(t *testing.T) {
	store := &memorySettings{data: map[string]string{}}
	svc, err := config.NewService(store)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := svc.Update(ctx, map[string]string{
		"DEEPSEEK_MODEL_NAME": "deepseek-v4-flash",
	}); err != nil {
		t.Fatal(err)
	}
	if got := svc.Current().Model.DefaultModel; got != "deepseek-v4-flash" {
		t.Fatalf("model = %q", got)
	}
	if got := store.data["DEEPSEEK_MODEL_NAME"]; got != "deepseek-v4-flash" {
		t.Fatalf("stored = %q", got)
	}
}

func TestService_RejectsInvalidUpdate(t *testing.T) {
	store := &memorySettings{data: map[string]string{}}
	svc, err := config.NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	err = svc.Update(context.Background(), map[string]string{
		"ROUND_TABLE_MODEL_TIMEOUT_SEC": "not-a-number",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestService_IgnoresSecretKeys(t *testing.T) {
	store := &memorySettings{data: map[string]string{}}
	svc, err := config.NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Update(context.Background(), map[string]string{
		"DEEPSEEK_API_KEY": "sk-should-not-store",
	}); err != nil {
		t.Fatal(err)
	}
	if len(store.data) != 0 {
		t.Fatalf("secrets must not be stored: %+v", store.data)
	}
}

func TestService_SettingsViewSecretsReadOnly(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "sk-test")
	svc, err := config.NewService(nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := svc.SettingsView()
	var deepseek config.SettingsFieldState
	for _, f := range resp.Fields {
		if f.Key == "DEEPSEEK_API_KEY" {
			deepseek = f
		}
	}
	if !deepseek.Secret || deepseek.Editable || !deepseek.Configured || deepseek.Value != "" {
		t.Fatalf("deepseek field = %+v", deepseek)
	}
	if resp.Source != "yaml" {
		t.Fatalf("source = %q", resp.Source)
	}
}

func TestService_LoadsFromSQLite(t *testing.T) {
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "cfg.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	ctx := context.Background()
	if err := st.SetSettings(ctx, map[string]string{
		"DEEPSEEK_MODEL_NAME": "from-db",
	}); err != nil {
		t.Fatal(err)
	}

	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}
	if got := svc.Current().Model.DefaultModel; got != "from-db" {
		t.Fatalf("model = %q", got)
	}
	if svc.SettingsView().Source != "app_settings" {
		t.Fatalf("expected app_settings source")
	}
}

func TestService_SettingsView_appliesDiscordBotProfileCache(t *testing.T) {
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "bot-profiles.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	ctx := context.Background()
	cache := `{"moderator":{"discord_username":"mod_bot","avatar_url":"https://cdn.example/mod.png","fetched_at":"2026-01-01T00:00:00Z"}}`
	if err := st.SetSettings(ctx, map[string]string{
		config.DiscordBotProfilesSetting: cache,
	}); err != nil {
		t.Fatal(err)
	}

	svc, err := config.NewService(st)
	if err != nil {
		t.Fatal(err)
	}
	resp := svc.SettingsView()
	var mod config.DiscordBotState
	for _, b := range resp.DiscordBots {
		if b.Primary {
			mod = b
			break
		}
	}
	if mod.AvatarURL != "https://cdn.example/mod.png" {
		t.Fatalf("moderator avatar = %+v", mod)
	}
	if mod.DiscordUsername != "mod_bot" {
		t.Fatalf("moderator username = %+v", mod)
	}
}
