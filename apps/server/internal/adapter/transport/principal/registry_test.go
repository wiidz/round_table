package principal

import (
	"testing"
	"time"
)

func TestRegistry_bindUnbind(t *testing.T) {
	path := t.TempDir() + "/bindings.json"
	r, err := NewRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	scope := ScopeKey("discord", "guild-1", "")

	b, err := r.Bind(scope, "discord", "user-a", "Alice")
	if err != nil {
		t.Fatal(err)
	}
	if b.PrincipalID != "discord:user-a" {
		t.Fatalf("principal_id = %q", b.PrincipalID)
	}

	if _, err := r.Bind(scope, "discord", "user-b", "Bob"); err == nil {
		t.Fatal("expected conflict")
	}

	if _, err := r.Bind(scope, "discord", "user-a", "Alice"); err != nil {
		t.Fatal("re-bind same user should succeed")
	}

	if err := r.Unbind(scope, "user-b"); err == nil {
		t.Fatal("wrong user unbind")
	}
	if err := r.Unbind(scope, "user-a"); err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Get(scope); ok {
		t.Fatal("expected empty after unbind")
	}
}

func TestRegistry_persistence(t *testing.T) {
	path := t.TempDir() + "/bindings.json"
	scope := ScopeKey("discord", "", "user-dm")

	r1, err := NewRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r1.Bind(scope, "discord", "u1", "DM User"); err != nil {
		t.Fatal(err)
	}

	r2, err := NewRegistry(path)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := r2.Get(scope)
	if !ok || b.ExternalID != "u1" {
		t.Fatalf("reload = %+v ok=%v", b, ok)
	}
	if b.BoundAt.IsZero() {
		t.Fatal("bound_at")
	}
	_ = time.Now()
}

func TestScopeKey(t *testing.T) {
	if got := ScopeKey("discord", "g1", "u1"); got != "discord:guild:g1" {
		t.Fatalf("guild scope = %q", got)
	}
	if got := ScopeKey("discord", "", "u1"); got != "discord:dm:u1" {
		t.Fatalf("dm scope = %q", got)
	}
}
