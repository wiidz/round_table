package bootstrap

import (
	"fmt"
	"time"

	knowfs "round_table/apps/server/internal/adapter/knowledge/fs"
	"round_table/apps/server/internal/adapter/model/openai_compat"
	participantllm "round_table/apps/server/internal/adapter/participant/llm"
	"round_table/apps/server/internal/adapter/principal"
	prinstub "round_table/apps/server/internal/adapter/principal/stub"
	profilefs "round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/storage/memory"
	wsfs "round_table/apps/server/internal/adapter/workspace/fs"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/engine"
	"round_table/apps/server/internal/platform/config"
)

// PrincipalOptions configures the default Principal stub for CLI runs.
type PrincipalOptions struct {
	ForceSynthesisAtRound int
	ForceSynthesisReason  string
	PauseAtRound          int
	AbortAtRound          int
}

// NewEngine wires adapters from configuration for a real LLM-backed meeting run.
func NewEngine(cfg config.Config, principalOpts ...PrincipalOptions) (*engine.Engine, error) {
	var stub PrincipalOptions
	if len(principalOpts) > 0 {
		stub = principalOpts[0]
	}
	return newEngine(cfg, nil, stub)
}

// NewEngineWithPrincipal wires a custom Principal port (e.g. Discord confirmation).
func NewEngineWithPrincipal(cfg config.Config, prin principal.Port) (*engine.Engine, error) {
	if prin == nil {
		return nil, fmt.Errorf("bootstrap: principal port required")
	}
	return newEngine(cfg, prin, PrincipalOptions{})
}

func newEngine(cfg config.Config, prin principal.Port, stubOpts PrincipalOptions) (*engine.Engine, error) {
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

	if prin == nil {
		stub := &prinstub.Principal{
			ForceSynthesisWhenRoundGTE: stubOpts.ForceSynthesisAtRound,
			ForceSynthesisReason:       stubOpts.ForceSynthesisReason,
			PauseWhenRoundGTE:          stubOpts.PauseAtRound,
			AbortWhenRoundGTE:          stubOpts.AbortAtRound,
		}
		prin = stub
	}

	eng := engine.New(
		memory.New(),
		consensus.NoObjection{},
		parts,
		prin,
		ws,
		prof,
		know,
	)
	eng.Model = modelClient
	eng.ModelName = cfg.Model.DefaultModel
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
