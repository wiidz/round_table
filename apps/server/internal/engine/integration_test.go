package engine_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knowport "round_table/apps/server/internal/adapter/knowledge"
	"round_table/apps/server/internal/adapter/knowledge/fs"
	"round_table/apps/server/internal/adapter/participant/stub"
	prinstub "round_table/apps/server/internal/adapter/principal/stub"
	profilefs "round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/storage/memory"
	"round_table/apps/server/internal/adapter/workspace"
	wsfs "round_table/apps/server/internal/adapter/workspace/fs"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/engine"
)

func TestEngine_Integration_skipConfirmation(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{}, nil)

	spec := engine.CreateMeetingInput{
		MeetingID:        "mtg-int-1",
		Topic:            "API 设计评审",
		ConfirmationMode: meeting.ConfirmationModeSkip,
		Participants: []engine.ParticipantInput{
			{ID: "architect", Role: "Architect", Expertise: "system design"},
			{ID: "developer", Role: "Developer", Expertise: "backend"},
		},
	}

	s, err := eng.CreateMeeting(ctx, spec)
	if err != nil {
		t.Fatalf("CreateMeeting: %v", err)
	}
	if s.Status != meeting.StatusPreparing {
		t.Fatalf("after create status = %s", s.Status)
	}

	s, err = eng.Run(ctx, spec.MeetingID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if s.Status != meeting.StatusCompleted {
		t.Fatalf("final status = %s, want Completed", s.Status)
	}
	if s.Consensus == nil {
		t.Fatal("expected consensus")
	}
	if len(s.Minutes.Rounds) != 1 {
		t.Fatalf("rounds in minutes = %d", len(s.Minutes.Rounds))
	}

	wsRoot := filepath.Join(dataRoot, "workspaces", spec.MeetingID)
	assertFileContains(t, filepath.Join(wsRoot, workspace.FileMeeting), "API 设计评审")
	assertFileContains(t, filepath.Join(wsRoot, workspace.FileMinutes), "Round 1")
	assertFileContains(t, filepath.Join(wsRoot, "rounds", "round-001.md"), "Round 1")
	assertFileContains(t, filepath.Join(wsRoot, "artifacts", "minutes.md"), "Consensus")

	assertFileExists(t, filepath.Join(dataRoot, "profiles", "participants", "architect", "SOUL.md"))
	assertFileExists(t, filepath.Join(dataRoot, "knowledge", "participants", "architect", knowport.FileMemory))

	events, _ := eng.Store.List(ctx, spec.MeetingID)
	if len(events) < 8 {
		t.Fatalf("expected >=8 events, got %d", len(events))
	}
}

func TestEngine_Integration_maxRoundsModeratorDecision(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{Stance: "object", Content: "需要修改", ObjectReason: "方案不完整"}, nil)

	spec := engine.CreateMeetingInput{
		MeetingID:           "mtg-int-2",
		Topic:               "僵局测试",
		ConfirmationMode:    meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment: 1,
		Participants: []engine.ParticipantInput{
			{ID: "p1", Role: "Expert"},
		},
	}

	if _, err := eng.CreateMeeting(ctx, spec); err != nil {
		t.Fatal(err)
	}
	final, err := eng.Run(ctx, spec.MeetingID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if final.Status != meeting.StatusCompleted {
		t.Fatalf("status = %s", final.Status)
	}
	if final.Consensus == nil || final.Consensus.ResolvedBy != "moderator" {
		t.Fatalf("expected moderator decision, got %+v", final.Consensus)
	}
}

func TestEngine_Integration_requiredConfirmation(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{}, &prinstub.Principal{})

	spec := engine.CreateMeetingInput{
		MeetingID:        "mtg-int-3",
		Topic:            "Confirmation 测试",
		ConfirmationMode: meeting.ConfirmationModeRequired,
		Participants: []engine.ParticipantInput{
			{ID: "p1", Role: "Architect"},
			{ID: "p2", Role: "Developer"},
		},
	}

	if _, err := eng.CreateMeeting(ctx, spec); err != nil {
		t.Fatal(err)
	}
	final, err := eng.Run(ctx, spec.MeetingID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if final.Status != meeting.StatusCompleted {
		t.Fatalf("status = %s", final.Status)
	}
	if final.Confirmation == nil || !final.Confirmation.Approved {
		t.Fatal("expected approved confirmation")
	}

	wsRoot := filepath.Join(dataRoot, "workspaces", spec.MeetingID)
	assertFileContains(t, filepath.Join(wsRoot, "confirmation", "brief.md"), "Confirmation Brief")
}

func TestEngine_Integration_confirmationRejectThenApprove(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{}, &prinstub.Principal{
		RejectUntilCycle: 2,
		Feedback:           "需要更多细节",
	})

	spec := engine.CreateMeetingInput{
		MeetingID:        "mtg-int-4",
		Topic:            "驳回后再共识",
		ConfirmationMode: meeting.ConfirmationModeRequired,
		Participants: []engine.ParticipantInput{
			{ID: "p1", Role: "Expert"},
		},
	}

	if _, err := eng.CreateMeeting(ctx, spec); err != nil {
		t.Fatal(err)
	}
	final, err := eng.Run(ctx, spec.MeetingID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if final.Status != meeting.StatusCompleted {
		t.Fatalf("status = %s", final.Status)
	}
	if final.ConfirmationCycle != 1 {
		t.Fatalf("confirmation cycle = %d, want 1 after one reject", final.ConfirmationCycle)
	}
	if len(final.Minutes.Rounds) < 2 {
		t.Fatalf("expected >=2 rounds after reject, got %d", len(final.Minutes.Rounds))
	}
}

func newTestEngine(t *testing.T, dataRoot string, parts *stub.Participant, prin *prinstub.Principal) *engine.Engine {
	t.Helper()
	templates := filepath.Join(repoRoot(t), "data", "_templates")
	if prin == nil {
		prin = &prinstub.Principal{}
	}
	return engine.New(
		memory.New(),
		consensus.NoObjection{},
		parts,
		prin,
		wsfs.NewStore(filepath.Join(dataRoot, "workspaces")),
		profilefs.NewStore(filepath.Join(dataRoot, "profiles"), filepath.Join(templates, "profiles")),
		fs.NewStore(filepath.Join(dataRoot, "knowledge"), filepath.Join(templates, "knowledge")),
	)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.work not found")
		}
		dir = parent
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("missing file %s: %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, sub string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), sub) {
		t.Fatalf("%s: want substring %q in:\n%s", path, sub, data)
	}
}
