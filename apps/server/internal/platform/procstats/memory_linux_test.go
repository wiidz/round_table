//go:build linux

package procstats

import (
	"os"
	"testing"
	"time"
)

func TestLinuxProcRSS_self(t *testing.T) {
	rss, ok := linuxProcRSS(os.Getpid())
	if !ok || rss <= 0 {
		t.Fatalf("rss=%d ok=%v", rss, ok)
	}
}

func TestServerSnapshot(t *testing.T) {
	MarkServerStarted()
	snap := ServerSnapshot()
	if snap.PID != os.Getpid() {
		t.Fatalf("pid=%d", snap.PID)
	}
	if snap.UptimeSeconds < 0 {
		t.Fatalf("uptime=%d", snap.UptimeSeconds)
	}
	if snap.MemoryBytes <= 0 {
		t.Fatalf("memory=%d source=%q", snap.MemoryBytes, snap.MemorySource)
	}
}

func TestProcessSnapshot(t *testing.T) {
	started := time.Now().Add(-2 * time.Minute)
	snap := ProcessSnapshot(os.Getpid(), started)
	if snap.UptimeSeconds < 110 {
		t.Fatalf("uptime=%d", snap.UptimeSeconds)
	}
}
