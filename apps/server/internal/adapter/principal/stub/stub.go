package stub

import (
	"context"

	"round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/domain/event"
)

// Principal is a test double for Confirmation decisions.
type Principal struct {
	// RejectUntilCycle rejects when cycle < RejectUntilCycle (1-based cycle from engine).
	RejectUntilCycle int
	Feedback           string
}

var _ principal.Port = (*Principal)(nil)

// Confirm implements principal.Port.
func (p *Principal) Confirm(ctx context.Context, _ string, _ event.ConfirmationBrief, cycle int) (principal.Response, error) {
	if err := ctx.Err(); err != nil {
		return principal.Response{}, err
	}
	if p.RejectUntilCycle > 0 && cycle < p.RejectUntilCycle {
		fb := p.Feedback
		if fb == "" {
			fb = "需要更多细节"
		}
		return principal.Response{Decision: principal.DecisionRejected, Feedback: fb}, nil
	}
	return principal.Response{Decision: principal.DecisionApproved}, nil
}
