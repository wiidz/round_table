package scheduler

import "testing"

func TestFixedOrder(t *testing.T) {
	t.Parallel()

	order := []string{"a", "b", "c"}
	spoken := map[string]bool{}

	id, ok := FixedOrder(order, spoken)
	if !ok || id != "a" {
		t.Fatalf("first = %q, ok = %v", id, ok)
	}
	spoken["a"] = true

	id, ok = FixedOrder(order, spoken)
	if !ok || id != "b" {
		t.Fatalf("second = %q", id)
	}
	spoken["b"] = true
	spoken["c"] = true

	if _, ok := FixedOrder(order, spoken); ok {
		t.Fatal("expected round complete")
	}
	if !RoundComplete(order, spoken) {
		t.Fatal("RoundComplete should be true")
	}
}
