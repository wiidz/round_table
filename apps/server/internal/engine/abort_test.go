package engine

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	knowfs "round_table/apps/server/internal/adapter/knowledge/fs"
	profilefs "round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/storage/memory"
	wsfs "round_table/apps/server/internal/adapter/workspace/fs"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/meeting"
)

func newAbortTestEngine(t *testing.T, dataRoot string) *Engine {
	t.Helper()
	eng := New(
		memory.New(),
		consensus.NoObjection{},
		nil,
		nil,
		wsfs.NewStore(filepath.Join(dataRoot, "workspaces")),
		profilefs.NewStore(filepath.Join(dataRoot, "profiles"), filepath.Join(dataRoot, "templates", "profiles")),
		knowfs.NewStore(filepath.Join(dataRoot, "knowledge"), filepath.Join(dataRoot, "templates", "knowledge")),
	)
	eng.Progress = DiscardProgressLogger{}
	return eng
}

func TestAbortMeeting_runningMeeting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	eng := newAbortTestEngine(t, t.TempDir())

	freeQ := 0
	spec := CreateMeetingInput{
		MeetingID:                "mtg-abort-run-fail",
		Topic:                    "中断测试",
		ConfirmationMode:         meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment:      2,
		FreeDialogueMaxQuestions: &freeQ,
		Participants: []ParticipantInput{
			{ID: "a", Role: "Architect"},
		},
	}
	if _, err := eng.CreateMeeting(ctx, spec); err != nil {
		t.Fatal(err)
	}

	final, err := eng.AbortMeeting(ctx, spec.MeetingID, "进程中断")
	if err != nil {
		t.Fatalf("AbortMeeting: %v", err)
	}
	if final.Status != meeting.StatusCompleted {
		t.Fatalf("status = %s, want Completed", final.Status)
	}
	if final.Outcome != meeting.OutcomeAborted {
		t.Fatalf("outcome = %q, want aborted", final.Outcome)
	}

	_, err = eng.AbortMeeting(ctx, spec.MeetingID, "again")
	if !errors.Is(err, ErrMeetingAlreadyFinished) {
		t.Fatalf("second abort err = %v, want ErrMeetingAlreadyFinished", err)
	}
}

func TestAbortMeeting_notAbortable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	eng := newAbortTestEngine(t, t.TempDir())

	_, err := eng.AbortMeeting(ctx, "mtg-missing", "manual")
	if !errors.Is(err, ErrMeetingNotAbortable) {
		t.Fatalf("err = %v, want ErrMeetingNotAbortable", err)
	}
}
