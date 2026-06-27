package config

import "context"

// SettingsStore persists non-secret runtime settings (P0.5 app_settings).
type SettingsStore interface {
	GetAllSettings(ctx context.Context) (map[string]string, error)
	SetSettings(ctx context.Context, updates map[string]string) error
}
