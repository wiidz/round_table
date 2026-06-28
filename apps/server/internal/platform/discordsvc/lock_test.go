package discordsvc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadPIDFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "discord-transport.pid.lock")
	if err := os.WriteFile(path, []byte("12345\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := readPIDFromFile(path); got != 12345 {
		t.Fatalf("pid = %d", got)
	}
}

func TestProcessAlive(t *testing.T) {
	if !processAlive(os.Getpid()) {
		t.Fatal("current process should be alive")
	}
	if processAlive(999999999) {
		t.Fatal("unexpected pid should not be alive")
	}
}

func TestAcquireTransportLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "discord-transport.pid.lock")
	f, err := AcquireTransportLock(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := readPIDFromFile(path); got != os.Getpid() {
		t.Fatalf("lock file pid = %d", got)
	}
	ReleaseTransportLock(f)

	f2, err := AcquireTransportLock(path)
	if err != nil {
		t.Fatal(err)
	}
	ReleaseTransportLock(f2)
}

func TestAcquireTransportLockContention(t *testing.T) {
	path := filepath.Join(t.TempDir(), "discord-transport.pid.lock")
	first, err := AcquireTransportLock(path)
	if err != nil {
		t.Fatal(err)
	}
	defer ReleaseTransportLock(first)

	if _, err := AcquireTransportLock(path); err == nil {
		t.Fatal("expected contention error")
	}
}
