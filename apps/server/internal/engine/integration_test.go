package engine_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knowport "round_table/apps/server/internal/adapter/knowledge"
	"round_table/apps/server/internal/adapter/knowledge/fs"
	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/participant/stub"
	prinstub "round_table/apps/server/internal/adapter/principal/stub"
	profilefs "round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/storage/memory"
	"round_table/apps/server/internal/adapter/workspace"
	wsfs "round_table/apps/server/internal/adapter/workspace/fs"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/engine"
)

func TestEngine_Integration_skipConfirmation(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{}, nil, nil)

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
	if !s.FreeDialogueCompleted {
		t.Fatal("expected free dialogue after round 1")
	}
	if len(s.FreeDialogueExchanges) != 2 {
		t.Fatalf("free dialogue exchanges = %d, want 2 (2 participants × 1 question each)", len(s.FreeDialogueExchanges))
	}
	if len(s.Minutes.Rounds) != 2 {
		t.Fatalf("rounds in minutes = %d", len(s.Minutes.Rounds))
	}

	wsRoot := filepath.Join(dataRoot, "workspaces", spec.MeetingID)
	assertFileContains(t, filepath.Join(wsRoot, workspace.FileMeeting), "会议主题")
	assertFileContains(t, filepath.Join(wsRoot, workspace.FileMeeting), "参会人员")
	assertFileContains(t, filepath.Join(wsRoot, workspace.FileMeeting), "architect")
	assertFileContains(t, filepath.Join(wsRoot, workspace.FileMinutes), "Pre-meeting")
	assertFileContains(t, filepath.Join(wsRoot, workspace.FileMinutes), "Round 1")
	assertFileContains(t, filepath.Join(wsRoot, workspace.FileMinutes), "Free dialogue")
	assertFileContains(t, filepath.Join(wsRoot, "free-dialogue", "after-round-001.md"), "Free Dialogue")
	assertFileContains(t, filepath.Join(wsRoot, "pre-meeting", "perspectives.md"), "Pre-meeting")
	assertFileContains(t, filepath.Join(wsRoot, "rounds", "round-001.md"), "Round 1")
	assertFileContains(t, filepath.Join(wsRoot, "artifacts", "minutes.md"), "Consensus")
	assertFileExists(t, filepath.Join(wsRoot, "usage", "summary.md"))

	assertFileExists(t, filepath.Join(dataRoot, "profiles", "participants", "architect", "SOUL.md"))
	assertFileExists(t, filepath.Join(dataRoot, "knowledge", "participants", "architect", knowport.FileMemory))

	events, _ := eng.Store.List(ctx, spec.MeetingID)
	if len(events) < 8 {
		t.Fatalf("expected >=8 events, got %d", len(events))
	}
}

func TestEngine_Integration_freeDialogueDisabled(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{}, nil, nil)

	spec := engine.CreateMeetingInput{
		MeetingID:                "mtg-int-fd-off",
		Topic:                    "关闭自由对话",
		ConfirmationMode:         meeting.ConfirmationModeSkip,
		FreeDialogueMaxQuestions: intPtr(0),
		Participants: []engine.ParticipantInput{
			{ID: "architect", Role: "Architect"},
			{ID: "developer", Role: "Developer"},
		},
	}

	if _, err := eng.CreateMeeting(ctx, spec); err != nil {
		t.Fatal(err)
	}
	final, err := eng.Run(ctx, spec.MeetingID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if final.FreeDialogueMaxQuestions != 0 {
		t.Fatalf("free_dialogue_max_questions = %d, want 0", final.FreeDialogueMaxQuestions)
	}
	if final.FreeDialogueCompleted {
		t.Fatal("free dialogue should not run when disabled")
	}
	if len(final.FreeDialogueExchanges) != 0 {
		t.Fatalf("exchanges = %d, want 0", len(final.FreeDialogueExchanges))
	}

	wsRoot := filepath.Join(dataRoot, "workspaces", spec.MeetingID)
	if _, err := os.Stat(filepath.Join(wsRoot, "free-dialogue", "after-round-001.md")); err == nil {
		t.Fatal("free-dialogue file should not exist when disabled")
	}
}

func TestEngine_Integration_maxRoundsModeratorDecision(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{Stance: "object", Content: "需要修改", ObjectReason: "方案不完整"}, nil, nil)

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

func TestEngine_Integration_deliberationMode(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{Stance: "agree", Content: "技能框架：三连击 + 位移"}, nil, nil)

	spec := engine.CreateMeetingInput{
		MeetingID:                "mtg-delib-1",
		Topic:                    "设计新职业「影舞者」的核心技能",
		MeetingMode:              meeting.MeetingModeDeliberation,
		ConfirmationMode:         meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment:      2,
		FreeDialogueMaxQuestions: intPtr(0),
		Participants: []engine.ParticipantInput{
			{ID: "designer", Role: "策划", Expertise: "gameplay"},
			{ID: "player", Role: "玩家代表", Expertise: "experience"},
		},
	}
	if _, err := eng.CreateMeeting(ctx, spec); err != nil {
		t.Fatal(err)
	}
	final, err := eng.Run(ctx, spec.MeetingID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !final.IsDeliberation() {
		t.Fatal("expected deliberation mode")
	}
	if final.Consensus == nil || final.Consensus.ResolvedBy != "max_rounds" {
		t.Fatalf("consensus = %+v", final.Consensus)
	}
	if final.SynthesisSummary == "" {
		t.Fatal("expected synthesis summary")
	}
	if !strings.Contains(final.SynthesisSummary, "Executive Summary") {
		t.Fatalf("missing executive summary in synthesis")
	}

	wsRoot := filepath.Join(dataRoot, "workspaces", spec.MeetingID)
	assertFileExists(t, filepath.Join(wsRoot, "artifacts", "design-draft.md"))
	assertFileExists(t, filepath.Join(wsRoot, "moderator", "round-002-summary.md"))
}

func TestEngine_Integration_deliberationEarlySynthesis(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	llm := integrationPhaseModel{
		readiness: `{"ready": true, "rationale": "要素已齐", "gaps": []}`,
		synthesis: `{"core_scheme":["核心方案"],"decisions":[],"open_questions":["待验证？"]}`,
	}
	eng := newTestEngine(t, dataRoot, &stub.Participant{Stance: "agree", Content: "技能框架：三连击 + 位移"}, nil, llm)

	spec := engine.CreateMeetingInput{
		MeetingID:                "mtg-delib-early",
		Topic:                    "设计新职业「影舞者」的核心技能",
		MeetingMode:              meeting.MeetingModeDeliberation,
		ConfirmationMode:         meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment:      5,
		MinRoundsBeforeSynthesis: intPtr(2),
		FreeDialogueMaxQuestions: intPtr(0),
		Participants: []engine.ParticipantInput{
			{ID: "designer", Role: "策划"},
			{ID: "player", Role: "玩家代表"},
		},
	}
	if _, err := eng.CreateMeeting(ctx, spec); err != nil {
		t.Fatal(err)
	}
	final, err := eng.Run(ctx, spec.MeetingID)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if final.Consensus == nil || final.Consensus.ResolvedBy != "readiness" {
		t.Fatalf("consensus = %+v, want resolved_by=readiness", final.Consensus)
	}
	if final.DebateRoundCount() != 2 {
		t.Fatalf("debate rounds = %d, want 2 (early stop)", final.DebateRoundCount())
	}

	wsRoot := filepath.Join(dataRoot, "workspaces", spec.MeetingID)
	assertFileExists(t, filepath.Join(wsRoot, "moderator", "round-002-readiness.md"))
}

func TestEngine_Integration_principalForceSynthesis(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{Stance: "agree", Content: "技能框架：三连击 + 位移"}, &prinstub.Principal{
		ForceSynthesisWhenRoundGTE: 2,
		ForceSynthesisReason:         "Principal 要求立即出草案",
	}, nil)

	spec := engine.CreateMeetingInput{
		MeetingID:                "mtg-force-synth",
		Topic:                    "设计新职业核心技能",
		MeetingMode:              meeting.MeetingModeDeliberation,
		ConfirmationMode:         meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment:      5,
		MinRoundsBeforeSynthesis: intPtr(2),
		FreeDialogueMaxQuestions: intPtr(0),
		Participants: []engine.ParticipantInput{
			{ID: "designer", Role: "策划"},
			{ID: "player", Role: "玩家代表"},
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
	if final.Consensus == nil || final.Consensus.ResolvedBy != "principal" {
		t.Fatalf("consensus = %+v, want resolved_by=principal", final.Consensus)
	}
	if final.SynthesisSummary == "" {
		t.Fatal("expected synthesis summary")
	}
	if final.DebateRoundCount() >= 5 {
		t.Fatalf("expected early stop via force synthesis, got %d debate rounds", final.DebateRoundCount())
	}
}

func TestEngine_Integration_principalForceConsensus(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{Stance: "agree", Content: "方案可行"}, &prinstub.Principal{
		ForceConsensus: true,
	}, nil)

	spec := engine.CreateMeetingInput{
		MeetingID:           "mtg-force-consensus",
		Topic:               "是否批准上线",
		MeetingMode:         meeting.MeetingModeDecision,
		ConfirmationMode:    meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment: 5,
		Participants: []engine.ParticipantInput{
			{ID: "a", Role: "Architect"},
			{ID: "b", Role: "Developer"},
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
	if final.Consensus == nil || final.Consensus.ResolvedBy != "principal" {
		t.Fatalf("consensus = %+v, want resolved_by=principal", final.Consensus)
	}
	if final.DebateRoundCount() >= 5 {
		t.Fatal("expected early stop via force consensus")
	}
}

func TestEngine_Integration_principalPauseResume(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{Stance: "object", Content: "还需补充细节", ObjectReason: "方案不完整"}, &prinstub.Principal{
		PauseWhenRoundGTE: 2,
		PauseReason:       "Principal 暂停检查",
	}, nil)

	spec := engine.CreateMeetingInput{
		MeetingID:                "mtg-pause-resume",
		Topic:                    "暂停恢复测试",
		ConfirmationMode:         meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment:      5,
		FreeDialogueMaxQuestions: intPtr(0),
		Participants: []engine.ParticipantInput{
			{ID: "a", Role: "Architect"},
			{ID: "b", Role: "Developer"},
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
		t.Fatalf("status = %s, want Completed", final.Status)
	}
	if final.Consensus == nil {
		t.Fatal("expected consensus after pause/resume")
	}
	events, err := eng.Store.List(ctx, spec.MeetingID)
	if err != nil {
		t.Fatal(err)
	}
	if !eventTypesContain(events, event.TypeMeetingPaused) {
		t.Fatal("expected MeetingPaused event")
	}
	if !eventTypesContain(events, event.TypeMeetingResumed) {
		t.Fatal("expected MeetingResumed event")
	}
}

func TestEngine_Integration_principalAbort(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{Stance: "object", Content: "还需补充细节", ObjectReason: "方案不完整"}, &prinstub.Principal{
		AbortWhenRoundGTE: 2,
		AbortReason:       "Principal 终止",
	}, nil)

	spec := engine.CreateMeetingInput{
		MeetingID:                "mtg-abort",
		Topic:                    "终止测试",
		ConfirmationMode:         meeting.ConfirmationModeSkip,
		MaxRoundsPerSegment:      5,
		FreeDialogueMaxQuestions: intPtr(0),
		Participants: []engine.ParticipantInput{
			{ID: "a", Role: "Architect"},
			{ID: "b", Role: "Developer"},
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
		t.Fatalf("status = %s, want Completed", final.Status)
	}
	if final.Outcome != meeting.OutcomeAborted {
		t.Fatalf("outcome = %q, want aborted", final.Outcome)
	}
	if final.Consensus != nil {
		t.Fatal("expected no consensus on abort")
	}
}

func eventTypesContain(events []event.Envelope, typ event.Type) bool {
	for _, ev := range events {
		if ev.Type == typ {
			return true
		}
	}
	return false
}

func TestEngine_Integration_requiredConfirmation(t *testing.T) {
	ctx := context.Background()
	dataRoot := t.TempDir()
	eng := newTestEngine(t, dataRoot, &stub.Participant{}, &prinstub.Principal{}, nil)

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
	}, nil)

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
	if len(final.Minutes.Rounds) < 3 {
		t.Fatalf("expected >=3 rounds after reject (pre-meeting + round 1 + round 2), got %d", len(final.Minutes.Rounds))
	}
}

func intPtr(n int) *int { return &n }

type integrationPhaseModel struct {
	readiness string
	synthesis string
}

func (m integrationPhaseModel) Complete(_ context.Context, req model.Request) (model.Response, error) {
	content := m.synthesis
	for _, msg := range req.Messages {
		if strings.Contains(msg.Content, "Phase: deliberation-readiness") {
			content = m.readiness
			break
		}
	}
	return model.Response{Content: content}, nil
}

func newTestEngine(t *testing.T, dataRoot string, parts *stub.Participant, prin *prinstub.Principal, llm model.Port) *engine.Engine {
	t.Helper()
	templates := filepath.Join(repoRoot(t), "data", "_templates")
	if prin == nil {
		prin = &prinstub.Principal{}
	}
	eng := engine.New(
		memory.New(),
		consensus.NoObjection{},
		parts,
		prin,
		wsfs.NewStore(filepath.Join(dataRoot, "workspaces")),
		profilefs.NewStore(filepath.Join(dataRoot, "profiles"), filepath.Join(templates, "profiles")),
		fs.NewStore(filepath.Join(dataRoot, "knowledge"), filepath.Join(templates, "knowledge")),
	)
	eng.Progress = engine.DiscardProgressLogger{}
	if llm != nil {
		eng.Model = llm
		eng.ModelName = "test-model"
	}
	return eng
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
