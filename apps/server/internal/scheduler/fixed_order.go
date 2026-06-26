package scheduler

// FixedOrder returns the next speaker in registration order (ADR-0007).
func FixedOrder(order []string, spoken map[string]bool) (string, bool) {
	for _, id := range order {
		if !spoken[id] {
			return id, true
		}
	}
	return "", false
}

// RoundComplete reports whether every participant in order has spoken.
func RoundComplete(order []string, spoken map[string]bool) bool {
	_, ok := FixedOrder(order, spoken)
	return !ok
}
