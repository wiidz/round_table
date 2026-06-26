package engine

import (
	"round_table/apps/server/internal/adapter/knowledge"
	"round_table/apps/server/internal/adapter/participant"
	"round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/adapter/storage"
	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/consensus"
)

// Engine orchestrates Meeting lifecycle (Constitution step 5).
type Engine struct {
	Store       storage.Store
	Strategy    consensus.Strategy
	Participant participant.Port
	Principal   principal.Port
	Workspace   workspace.Port
	Profile     profile.Port
	Knowledge   knowledge.Port
	Progress    ProgressLogger
}

// New returns an Engine with required dependencies.
func New(
	store storage.Store,
	strat consensus.Strategy,
	parts participant.Port,
	prin principal.Port,
	ws workspace.Port,
	prof profile.Port,
	know knowledge.Port,
) *Engine {
	return &Engine{
		Store:       store,
		Strategy:    strat,
		Participant: parts,
		Principal:   prin,
		Workspace:   ws,
		Profile:     prof,
		Knowledge:   know,
	}
}
