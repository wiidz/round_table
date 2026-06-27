package config_test

import (
	"testing"

	"round_table/apps/server/internal/platform/config"
)

func TestSettingsView_localeInServiceNotDiscord(t *testing.T) {
	svc, err := config.NewService(nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := svc.SettingsView()

	var localeInService bool
	for _, f := range resp.Fields {
		if f.Key == "ROUND_TABLE_DISCORD_LOCALE" {
			t.Fatalf("deprecated field still exposed: %+v", f)
		}
		if f.Key == "ROUND_TABLE_LOCALE" {
			if f.Group != "服务" {
				t.Fatalf("locale group = %q, want 服务", f.Group)
			}
			if !f.Editable {
				t.Fatal("locale should be editable")
			}
			if f.InputType != "radio" {
				t.Fatalf("locale input_type = %q, want radio", f.InputType)
			}
			if len(f.Options) != 2 {
				t.Fatalf("locale options = %d, want 2", len(f.Options))
			}
			localeInService = true
		}
	}
	if !localeInService {
		t.Fatal("ROUND_TABLE_LOCALE missing from settings view")
	}
}
