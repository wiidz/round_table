package discordsvc

import "testing"

func TestEnsureProxyDefaults(t *testing.T) {
	got := ensureProxyDefaults([]string{"PATH=/usr/bin"})
	if !hasEnvKey(got, "https_proxy") || !hasEnvKey(got, "http_proxy") {
		t.Fatalf("proxy defaults missing: %v", got)
	}

	keep := ensureProxyDefaults([]string{"https_proxy=http://custom:9999"})
	for _, e := range keep {
		if e == "https_proxy=http://127.0.0.1:7897" {
			t.Fatal("should not override existing proxy")
		}
	}
}

func TestSupervisorStatusIdle(t *testing.T) {
	var s Supervisor
	st := s.Status()
	if st.Running || st.PID != 0 {
		t.Fatalf("expected idle, got %+v", st)
	}
}

func TestSupervisorStopNotRunning(t *testing.T) {
	var s Supervisor
	if err := s.Stop(); err == nil {
		t.Fatal("expected error when not running")
	}
}

func TestSupervisorShutdownIdle(t *testing.T) {
	var s Supervisor
	s.Shutdown() // must not panic or error
}
