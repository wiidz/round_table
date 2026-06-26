package bootstrap

import (
	"fmt"
	"time"

	knowfs "round_table/apps/server/internal/adapter/knowledge/fs"
	"round_table/apps/server/internal/adapter/model/openai_compat"
	participantllm "round_table/apps/server/internal/adapter/participant/llm"
	prinstub "round_table/apps/server/internal/adapter/principal/stub"
	profilefs "round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/storage/memory"
	wsfs "round_table/apps/server/internal/adapter/workspace/fs"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/engine"
	"round_table/apps/server/internal/platform/config"
)

// NewEngine wires adapters from configuration for a real LLM-backed meeting run.
func NewEngine(cfg config.Config) (*engine.Engine, error) {
	key, err := resolveAPIKey(cfg)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(cfg.Model.TimeoutSec) * time.Second
	modelClient := openai_compat.NewClient(cfg.Model.BaseURL, key, timeout)

	ws := wsfs.NewStore(cfg.Workspace.Root)
	prof := profilefs.NewStore(cfg.Profile.Root, cfg.Profile.Templates)
	know := knowfs.NewStore(cfg.Knowledge.Root, cfg.Knowledge.Templates)
	parts := &participantllm.Participant{
		Model:     modelClient,
		Profile:   prof,
		ModelName: cfg.Model.DefaultModel,
	}

	eng := engine.New(
		memory.New(),
		consensus.NoObjection{},
		parts,
		&prinstub.Principal{},
		ws,
		prof,
		know,
	)
	eng.Progress = engine.StdProgressLogger{}
	eng.Stream = engine.StdStreamLogger{}
	return eng, nil
}

func resolveAPIKey(cfg config.Config) (string, error) {
	switch cfg.Model.Provider {
	case "deepseek", "":
		if cfg.Secrets.DeepSeekAPIKey == "" {
			return "", fmt.Errorf("bootstrap: DEEPSEEK_API_KEY required (set in apps/server/.env)")
		}
		return cfg.Secrets.DeepSeekAPIKey, nil
	case "openai":
		if cfg.Secrets.OpenAIAPIKey == "" {
			return "", fmt.Errorf("bootstrap: OPENAI_API_KEY required")
		}
		return cfg.Secrets.OpenAIAPIKey, nil
	default:
		return "", fmt.Errorf("bootstrap: unsupported model provider %q", cfg.Model.Provider)
	}
}
