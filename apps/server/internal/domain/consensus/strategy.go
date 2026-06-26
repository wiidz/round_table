package consensus

import "round_table/apps/server/internal/domain/meeting"

// Result is the outcome of a consensus evaluation (ADR-0002).
type Result struct {
	Reached    bool
	Dissent    []DissentingOpinion
	ResolvedBy string // strategy | moderator | principal
}

// DissentingOpinion records a participant objection when consensus still passes.
type DissentingOpinion struct {
	ParticipantID string
	Reason        string
}

// Context supplies data for strategy evaluation.
type Context struct {
	Meeting meeting.State
	// Stances filled by scheduler in later phases.
}

// Strategy evaluates whether consensus is reached (ADR-0002).
type Strategy interface {
	Evaluate(ctx Context) (Result, error)
}
