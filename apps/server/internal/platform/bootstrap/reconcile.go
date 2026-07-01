package bootstrap

import (
	"context"

	knowfs "round_table/apps/server/internal/adapter/knowledge/fs"
	profilefs "round_table/apps/server/internal/adapter/profile/fs"
	wsfs "round_table/apps/server/internal/adapter/workspace/fs"
	"round_table/apps/server/internal/adapter/storage"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/engine"
	"round_table/apps/server/internal/platform/config"
)

// NewMaintenanceEngine builds an Engine for abort/reconcile without LLM or Principal ports.
func NewMaintenanceEngine(cfg config.Config, store storage.Store) (*engine.Engine, error) {
	if store == nil {
		var err error
		store, err = OpenStorage(cfg.Storage)
		if err != nil {
			return nil, err
		}
	}
	ws := wsfs.NewStore(cfg.Workspace.Root)
	eng := engine.New(
		store,
		consensus.NoObjection{},
		nil,
		nil,
		ws,
		profilefs.NewStore(cfg.Profile.Root, cfg.Profile.Templates),
		knowfs.NewStore(cfg.Knowledge.Root, cfg.Knowledge.Templates),
	)
	eng.Progress = engine.DiscardProgressLogger{}
	return eng, nil
}

// ReconcileMeetings runs orphan meeting cleanup at startup or via CLI.
func ReconcileMeetings(ctx context.Context, cfg config.Config, store storage.Store, reason string) (engine.ReconcileResult, error) {
	eng, err := NewMaintenanceEngine(cfg, store)
	if err != nil {
		return engine.ReconcileResult{}, err
	}
	return engine.ReconcileMeetings(ctx, eng, reason)
}
