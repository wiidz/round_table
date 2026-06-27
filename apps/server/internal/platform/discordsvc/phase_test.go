package discordsvc

import (
	"os"
	"path/filepath"
	"testing"

	discordtransport "round_table/apps/server/internal/adapter/transport/discord"
)

func TestDetectSessionPhase_staleReadyIgnoredWithoutSpawn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "discord.log")
	content := discordtransport.ReadyLogMarker + " 2026-06-27T12:00:00Z\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	phase, readyAt := detectSessionPhase(path)
	if phase != "starting" {
		t.Fatalf("phase = %q want starting (stale ready ignored)", phase)
	}
	if readyAt != "" {
		t.Fatalf("readyAt = %q want empty", readyAt)
	}
}

func TestDetectSessionPhase_readyAfterSpawn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "discord.log")
	content := discordtransport.ReadyLogMarker + " 2026-06-27T12:00:00Z\n" +
		"[supervisor] discord transport process started · pid=123\n" +
		discordtransport.ReadyLogMarker + " 2026-06-27T12:33:12Z\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	phase, readyAt := detectSessionPhase(path)
	if phase != "ready" {
		t.Fatalf("phase = %q want ready", phase)
	}
	if readyAt != "2026-06-27T12:33:12Z" {
		t.Fatalf("readyAt = %q", readyAt)
	}
}

func TestDetectSessionPhase_ready(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "discord.log")
	content := "noise\n======== Discord transport · session 2026-06-27 20:33:00 CST ========\n" +
		"[discord] 正在启动…\n" +
		discordtransport.ReadyLogMarker + " 2026-06-27T12:33:12Z\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	phase, readyAt := detectSessionPhase(path)
	if phase != "ready" {
		t.Fatalf("phase = %q want ready", phase)
	}
	if readyAt != "2026-06-27T12:33:12Z" {
		t.Fatalf("readyAt = %q", readyAt)
	}
}

func TestDetectSessionPhase_starting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "discord.log")
	content := "======== Discord transport · session 2026-06-27 20:33:00 CST ========\n[discord] 正在启动…\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	phase, readyAt := detectSessionPhase(path)
	if phase != "starting" {
		t.Fatalf("phase = %q want starting", phase)
	}
	if readyAt != "" {
		t.Fatalf("readyAt = %q want empty", readyAt)
	}
}

func TestDetectSessionPhase_legacyReadyLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "discord.log")
	content := "--- session started 2026-06-27T12:00:00Z ---\n" +
		discordtransport.ReadyLogMarker + " 2026-06-27T12:33:12Z\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	phase, _ := detectSessionPhase(path)
	if phase != "ready" {
		t.Fatalf("phase = %q want ready", phase)
	}
}
