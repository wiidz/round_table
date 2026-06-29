package procstats

import (
	"os"
	"runtime"
	"time"
)

var serverStartedAt = time.Now()

// MarkServerStarted records HTTP server boot time (call once from main).
func MarkServerStarted() {
	serverStartedAt = time.Now()
}

// Snapshot is runtime info for one OS process.
type Snapshot struct {
	PID           int    `json:"pid,omitempty"`
	UptimeSeconds int64  `json:"uptime_seconds,omitempty"`
	MemoryBytes   int64  `json:"memory_bytes,omitempty"`
	MemorySource  string `json:"memory_source,omitempty"` // rss | heap
}

// ServerSnapshot returns stats for the current HTTP server process.
func ServerSnapshot() Snapshot {
	return snapshotForPID(os.Getpid(), serverStartedAt)
}

// ProcessSnapshot returns stats for another process when start time is known.
func ProcessSnapshot(pid int, startedAt time.Time) Snapshot {
	if pid <= 0 {
		return Snapshot{}
	}
	return snapshotForPID(pid, startedAt)
}

func snapshotForPID(pid int, startedAt time.Time) Snapshot {
	out := Snapshot{PID: pid}
	if !startedAt.IsZero() {
		sec := int64(time.Since(startedAt).Seconds())
		if sec < 0 {
			sec = 0
		}
		out.UptimeSeconds = sec
	}
	if rss, ok := processRSS(pid); ok && rss > 0 {
		out.MemoryBytes = rss
		out.MemorySource = "rss"
		return out
	}
	if pid == os.Getpid() {
		out.MemoryBytes = int64(heapInUse())
		out.MemorySource = "heap"
	}
	return out
}

func heapInUse() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}
