package bootstrap

import (
	"fmt"
	"time"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/model/openai_compat"
	"round_table/apps/server/internal/platform/config"
)

// NewModelPort builds an OpenAI-compatible model client from config.
func NewModelPort(cfg config.Config) (model.Port, string, error) {
	key, err := resolveAPIKey(cfg)
	if err != nil {
		return nil, "", err
	}
	timeout := time.Duration(cfg.Model.TimeoutSec) * time.Second
	name := cfg.Model.DefaultModel
	if name == "" {
		name = "deepseek-chat"
	}
	return openai_compat.NewClient(cfg.Model.BaseURL, key, timeout), name, nil
}

// NewModelPortOptional returns nil when API key is not configured (no error).
func NewModelPortOptional(cfg config.Config) (model.Port, string) {
	port, name, err := NewModelPort(cfg)
	if err != nil {
		return nil, ""
	}
	return port, name
}

// MustModelPort is for CLI paths that require a model.
func MustModelPort(cfg config.Config) (model.Port, string, error) {
	port, name, err := NewModelPort(cfg)
	if err != nil {
		return nil, "", fmt.Errorf("bootstrap: model: %w", err)
	}
	return port, name, nil
}
