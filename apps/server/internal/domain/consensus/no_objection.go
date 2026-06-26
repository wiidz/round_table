package consensus

import (
	"round_table/apps/server/internal/domain/event"
)

// NoObjection passes when every participant in the round responded and none objected (ADR-0002).
type NoObjection struct{}

func (NoObjection) Evaluate(ctx Context) (Result, error) {
	m := ctx.Meeting
	if m.CurrentRound <= 0 || len(m.RoundOrder) == 0 {
		return Result{Reached: false, ResolvedBy: "strategy"}, nil
	}
	responses := m.RoundResponses[m.CurrentRound]
	for _, id := range m.RoundOrder {
		r, ok := responses[id]
		if !ok {
			return Result{Reached: false, ResolvedBy: "strategy"}, nil
		}
		if r.Stance == event.StanceObject {
			return Result{Reached: false, ResolvedBy: "strategy"}, nil
		}
		if r.Stance == event.StanceNone || r.Stance == "" {
			return Result{Reached: false, ResolvedBy: "strategy"}, nil
		}
	}
	return Result{Reached: true, ResolvedBy: "strategy"}, nil
}
