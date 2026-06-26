package consensus

// NoObjection passes when no participant marked object (ADR-0002).
type NoObjection struct{}

func (NoObjection) Evaluate(_ Context) (Result, error) {
	// v0.1 skeleton: real stance counting in Phase 2.
	return Result{Reached: false, ResolvedBy: "strategy"}, nil
}
