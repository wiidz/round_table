package meeting

import (
	"encoding/json"
	"testing"
	"time"

	"round_table/apps/server/internal/domain/event"
)

func env(seq int, typ event.Type, payload []byte, actor event.Actor) event.Envelope {
	return event.Envelope{
		ID:         "evt",
		MeetingID:  "mtg-1",
		Sequence:   seq,
		Type:       typ,
		Version:    1,
		Payload:    payload,
		OccurredAt: time.Now(),
		Actor:      actor,
	}
}

func preMeetingEvents(startSeq int, order []string) []event.Envelope {
	must := func(v any) []byte {
		b, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return b
	}
	var out []event.Envelope
	seq := startSeq
	out = append(out, env(seq, event.TypeRoundStarted, must(event.RoundStartedPayload{
		RoundNumber: 0, Order: order,
	}), event.ActorModerator))
	seq++
	for _, id := range order {
		out = append(out, env(seq, event.TypeParticipantResponded, must(event.ParticipantRespondedPayload{
			ParticipantID: id, RoundNumber: 0, Content: "initial view", Stance: event.StanceNone,
		}), event.ActorParticipant))
		seq++
	}
	out = append(out, env(seq, event.TypeRoundCompleted, must(event.RoundCompletedPayload{
		RoundNumber: 0, Summary: "pre-meeting done",
	}), event.ActorModerator))
	return out
}

func mustPayload(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestApply_MeetingCreated(t *testing.T) {
	t.Parallel()

	payload := mustPayload(t, event.MeetingCreatedPayload{
		Topic:                 "架构评审",
		MaxRoundsPerSegment:   5,
		ConfirmationMode:      ConfirmationModeRequired,
		MaxConfirmationCycles: 3,
	})

	s, err := Apply(NewState("mtg-1"), env(1, event.TypeMeetingCreated, payload, event.ActorPrincipal))
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if s.Status != StatusPreparing {
		t.Fatalf("status = %s, want Preparing", s.Status)
	}
	if s.Topic != "架构评审" {
		t.Fatalf("topic = %q", s.Topic)
	}
	if s.MaxRoundsPerSegment != 5 {
		t.Fatalf("max_rounds = %d", s.MaxRoundsPerSegment)
	}
}

func TestApply_MeetingCreated_defaults(t *testing.T) {
	t.Parallel()

	payload := mustPayload(t, event.MeetingCreatedPayload{Topic: "t"})
	s, err := Apply(NewState("mtg-1"), env(1, event.TypeMeetingCreated, payload, event.ActorPrincipal))
	if err != nil {
		t.Fatal(err)
	}
	if s.MaxRoundsPerSegment != defaultMaxRoundsPerSegment {
		t.Fatalf("max_rounds default = %d", s.MaxRoundsPerSegment)
	}
	if s.ConfirmationMode != ConfirmationModeRequired {
		t.Fatalf("confirmation_mode = %q", s.ConfirmationMode)
	}
	if s.FreeDialogueMaxQuestions != defaultFreeDialogueMaxQuestions {
		t.Fatalf("free_dialogue_max_questions default = %d", s.FreeDialogueMaxQuestions)
	}
}

func TestApply_MeetingCreated_freeDialogueDisabled(t *testing.T) {
	t.Parallel()

	zero := 0
	payload := mustPayload(t, event.MeetingCreatedPayload{
		Topic:                    "t",
		FreeDialogueMaxQuestions: &zero,
	})
	s, err := Apply(NewState("mtg-1"), env(1, event.TypeMeetingCreated, payload, event.ActorPrincipal))
	if err != nil {
		t.Fatal(err)
	}
	if s.FreeDialogueMaxQuestions != 0 {
		t.Fatalf("free_dialogue_max_questions = %d, want 0", s.FreeDialogueMaxQuestions)
	}
}

func TestApply_MeetingCreated_twiceRejected(t *testing.T) {
	t.Parallel()

	payload := mustPayload(t, event.MeetingCreatedPayload{Topic: "x"})
	e := env(1, event.TypeMeetingCreated, payload, event.ActorPrincipal)

	s, err := Apply(NewState("mtg-1"), e)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Apply(s, e)
	if err == nil {
		t.Fatal("expected error on second MeetingCreated")
	}
}

func TestFold_happyPathWithConfirmation(t *testing.T) {
	t.Parallel()

	order := []string{"p1", "p2"}
	events := []event.Envelope{
		env(1, event.TypeMeetingCreated, mustPayload(t, event.MeetingCreatedPayload{
			Topic:            "设计评审",
			ConfirmationMode: ConfirmationModeRequired,
		}), event.ActorPrincipal),
		env(2, event.TypeParticipantInvited, mustPayload(t, event.ParticipantInvitedPayload{
			ParticipantID: "p1", Role: "Architect",
		}), event.ActorModerator),
		env(3, event.TypeParticipantInvited, mustPayload(t, event.ParticipantInvitedPayload{
			ParticipantID: "p2", Role: "Programmer",
		}), event.ActorModerator),
	}
	events = append(events, preMeetingEvents(4, order)...)
	events = append(events,
		env(8, event.TypeRoundStarted, mustPayload(t, event.RoundStartedPayload{
			RoundNumber: 1, Order: order,
		}), event.ActorModerator),
		env(9, event.TypeParticipantResponded, mustPayload(t, event.ParticipantRespondedPayload{
			ParticipantID: "p1", RoundNumber: 1, Content: "方案 A", Stance: event.StanceAgree,
		}), event.ActorParticipant),
		env(10, event.TypeParticipantResponded, mustPayload(t, event.ParticipantRespondedPayload{
			ParticipantID: "p2", RoundNumber: 1, Content: "同意", Stance: event.StanceAgree,
		}), event.ActorParticipant),
		env(11, event.TypeRoundCompleted, mustPayload(t, event.RoundCompletedPayload{
			RoundNumber: 1, Summary: "达成一致",
		}), event.ActorModerator),
		env(12, event.TypeConsensusReached, mustPayload(t, event.ConsensusReachedPayload{
			Strategy: "no_objection", ResolvedBy: "strategy",
		}), event.ActorModerator),
		env(13, event.TypeConfirmationPrepared, mustPayload(t, event.ConfirmationPreparedPayload{
			Cycle: 1,
			Brief: event.ConfirmationBrief{
				ExecutiveSummary: "采用方案 A",
				Items:            []event.ConfirmationItem{{Index: 1, Title: "架构", Description: "方案 A"}},
			},
		}), event.ActorModerator),
		env(14, event.TypeConfirmationPresented, mustPayload(t, event.ConfirmationPresentedPayload{Cycle: 1}), event.ActorModerator),
		env(15, event.TypeConfirmationApproved, mustPayload(t, event.ConfirmationApprovedPayload{Cycle: 1}), event.ActorPrincipal),
		env(16, event.TypeMeetingFinished, mustPayload(t, event.MeetingFinishedPayload{}), event.ActorModerator),
	)

	events = events[:16]

	s, err := Fold("mtg-1", events)
	if err != nil {
		t.Fatalf("Fold: %v", err)
	}
	if s.Status != StatusCompleted {
		t.Fatalf("status = %s, want Completed", s.Status)
	}
	if s.Consensus == nil {
		t.Fatal("expected consensus state")
	}
	if s.Confirmation == nil || !s.Confirmation.Approved {
		t.Fatal("expected approved confirmation")
	}
	if len(s.Minutes.Rounds) != 2 {
		t.Fatalf("round summaries = %d", len(s.Minutes.Rounds))
	}
	if len(s.Participants) != 2 {
		t.Fatalf("participants = %d", len(s.Participants))
	}
}

func TestFold_skipConfirmation(t *testing.T) {
	t.Parallel()

	events := []event.Envelope{
		env(1, event.TypeMeetingCreated, mustPayload(t, event.MeetingCreatedPayload{
			Topic: "快速决策", ConfirmationMode: ConfirmationModeSkip,
		}), event.ActorPrincipal),
		env(2, event.TypeParticipantInvited, mustPayload(t, event.ParticipantInvitedPayload{
			ParticipantID: "p1", Role: "Expert",
		}), event.ActorModerator),
	}
	events = append(events, preMeetingEvents(3, []string{"p1"})...)
	events = append(events,
		env(7, event.TypeRoundStarted, mustPayload(t, event.RoundStartedPayload{
			RoundNumber: 1, Order: []string{"p1"},
		}), event.ActorModerator),
		env(8, event.TypeParticipantResponded, mustPayload(t, event.ParticipantRespondedPayload{
			ParticipantID: "p1", RoundNumber: 1, Content: "ok", Stance: event.StanceAgree,
		}), event.ActorParticipant),
		env(9, event.TypeRoundCompleted, mustPayload(t, event.RoundCompletedPayload{
			RoundNumber: 1, Summary: "done",
		}), event.ActorModerator),
		env(10, event.TypeConsensusReached, mustPayload(t, event.ConsensusReachedPayload{
			Strategy: "no_objection", ResolvedBy: "strategy",
		}), event.ActorModerator),
		env(11, event.TypeMeetingFinished, mustPayload(t, event.MeetingFinishedPayload{}), event.ActorModerator),
	)

	s, err := Fold("mtg-1", events)
	if err != nil {
		t.Fatalf("Fold: %v", err)
	}
	if s.Status != StatusCompleted {
		t.Fatalf("status = %s", s.Status)
	}
}

func TestFold_confirmationRejected_addsOneRound(t *testing.T) {
	t.Parallel()

	order := []string{"p1"}
	base := []event.Envelope{
		env(1, event.TypeMeetingCreated, mustPayload(t, event.MeetingCreatedPayload{Topic: "x"}), event.ActorPrincipal),
		env(2, event.TypeParticipantInvited, mustPayload(t, event.ParticipantInvitedPayload{
			ParticipantID: "p1", Role: "Expert",
		}), event.ActorModerator),
	}
	base = append(base, preMeetingEvents(3, order)...)
	base = append(base,
		env(7, event.TypeRoundStarted, mustPayload(t, event.RoundStartedPayload{RoundNumber: 1, Order: order}), event.ActorModerator),
		env(8, event.TypeParticipantResponded, mustPayload(t, event.ParticipantRespondedPayload{
			ParticipantID: "p1", RoundNumber: 1, Content: "v1", Stance: event.StanceAgree,
		}), event.ActorParticipant),
		env(9, event.TypeRoundCompleted, mustPayload(t, event.RoundCompletedPayload{RoundNumber: 1, Summary: "s1"}), event.ActorModerator),
		env(10, event.TypeConsensusReached, mustPayload(t, event.ConsensusReachedPayload{
			Strategy: "no_objection", ResolvedBy: "strategy",
		}), event.ActorModerator),
		env(11, event.TypeConfirmationPrepared, mustPayload(t, event.ConfirmationPreparedPayload{
			Cycle: 1, Brief: event.ConfirmationBrief{ExecutiveSummary: "brief"},
		}), event.ActorModerator),
		env(12, event.TypeConfirmationPresented, mustPayload(t, event.ConfirmationPresentedPayload{Cycle: 1}), event.ActorModerator),
		env(13, event.TypeConfirmationRejected, mustPayload(t, event.ConfirmationRejectedPayload{
			Cycle: 1, Feedback: "需要更多细节",
		}), event.ActorPrincipal),
	)

	s, err := Fold("mtg-1", base)
	if err != nil {
		t.Fatalf("Fold after reject: %v", err)
	}
	if s.Status != StatusRunning {
		t.Fatalf("status = %s, want Running", s.Status)
	}
	if s.CurrentRound != 1 {
		t.Fatalf("current round = %d, want 1 (last completed)", s.CurrentRound)
	}
	if s.ConfirmationCycle != 1 {
		t.Fatalf("confirmation cycle = %d", s.ConfirmationCycle)
	}
	if s.PrincipalFeedback != "需要更多细节" {
		t.Fatalf("feedback = %q", s.PrincipalFeedback)
	}
	if s.SynthesisSummary != "" {
		t.Fatal("synthesis should be cleared after reject")
	}

	// One additional round after reject
	next := append(base,
		env(14, event.TypeRoundStarted, mustPayload(t, event.RoundStartedPayload{RoundNumber: 2, Order: order}), event.ActorModerator),
	)
	s, err = Fold("mtg-1", next)
	if err != nil {
		t.Fatalf("Fold post-reject round: %v", err)
	}
	if s.CurrentRound != 2 {
		t.Fatalf("current round = %d, want 2", s.CurrentRound)
	}
}

func TestFold_confirmationRejected_resetCycle(t *testing.T) {
	t.Parallel()

	order := []string{"p1"}
	base := []event.Envelope{
		env(1, event.TypeMeetingCreated, mustPayload(t, event.MeetingCreatedPayload{Topic: "x"}), event.ActorPrincipal),
		env(2, event.TypeParticipantInvited, mustPayload(t, event.ParticipantInvitedPayload{
			ParticipantID: "p1", Role: "Expert",
		}), event.ActorModerator),
	}
	base = append(base, preMeetingEvents(3, order)...)
	base = append(base,
		env(7, event.TypeRoundStarted, mustPayload(t, event.RoundStartedPayload{RoundNumber: 1, Order: order}), event.ActorModerator),
		env(8, event.TypeParticipantResponded, mustPayload(t, event.ParticipantRespondedPayload{
			ParticipantID: "p1", RoundNumber: 1, Content: "v1", Stance: event.StanceAgree,
		}), event.ActorParticipant),
		env(9, event.TypeRoundCompleted, mustPayload(t, event.RoundCompletedPayload{RoundNumber: 1, Summary: "s1"}), event.ActorModerator),
		env(10, event.TypeConsensusReached, mustPayload(t, event.ConsensusReachedPayload{
			Strategy: "no_objection", ResolvedBy: "strategy",
		}), event.ActorModerator),
		env(11, event.TypeConfirmationPrepared, mustPayload(t, event.ConfirmationPreparedPayload{
			Cycle: 3, Brief: event.ConfirmationBrief{ExecutiveSummary: "brief"},
		}), event.ActorModerator),
		env(12, event.TypeConfirmationPresented, mustPayload(t, event.ConfirmationPresentedPayload{Cycle: 3}), event.ActorModerator),
		env(13, event.TypeConfirmationRejected, mustPayload(t, event.ConfirmationRejectedPayload{
			Cycle: 3, Feedback: "继续研讨", ResetCycle: true,
		}), event.ActorPrincipal),
	)

	s, err := Fold("mtg-1", base)
	if err != nil {
		t.Fatalf("Fold: %v", err)
	}
	if s.ConfirmationCycle != 0 {
		t.Fatalf("confirmation cycle = %d, want 0 after reset", s.ConfirmationCycle)
	}
	if s.Status != StatusRunning {
		t.Fatalf("status = %s", s.Status)
	}
}

func TestFold_consensusVetoed(t *testing.T) {
	t.Parallel()

	events := []event.Envelope{
		env(1, event.TypeMeetingCreated, mustPayload(t, event.MeetingCreatedPayload{Topic: "x"}), event.ActorPrincipal),
		env(2, event.TypeParticipantInvited, mustPayload(t, event.ParticipantInvitedPayload{ParticipantID: "p1", Role: "E"}), event.ActorModerator),
	}
	events = append(events, preMeetingEvents(3, []string{"p1"})...)
	events = append(events,
		env(7, event.TypeRoundStarted, mustPayload(t, event.RoundStartedPayload{RoundNumber: 1, Order: []string{"p1"}}), event.ActorModerator),
		env(8, event.TypeParticipantResponded, mustPayload(t, event.ParticipantRespondedPayload{
			ParticipantID: "p1", RoundNumber: 1, Content: "c", Stance: event.StanceAgree,
		}), event.ActorParticipant),
		env(9, event.TypeRoundCompleted, mustPayload(t, event.RoundCompletedPayload{RoundNumber: 1, Summary: "s"}), event.ActorModerator),
		env(10, event.TypeConsensusReached, mustPayload(t, event.ConsensusReachedPayload{Strategy: "no_objection", ResolvedBy: "strategy"}), event.ActorModerator),
		env(11, event.TypeConsensusVetoed, mustPayload(t, event.ConsensusVetoedPayload{Reason: "再讨论"}), event.ActorPrincipal),
	)

	s, err := Fold("mtg-1", events)
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != StatusRunning {
		t.Fatalf("status = %s", s.Status)
	}
	if s.Consensus != nil {
		t.Fatal("consensus should be cleared")
	}
}

func TestFold_pauseResume(t *testing.T) {
	t.Parallel()

	events := []event.Envelope{
		env(1, event.TypeMeetingCreated, mustPayload(t, event.MeetingCreatedPayload{Topic: "x"}), event.ActorPrincipal),
		env(2, event.TypeParticipantInvited, mustPayload(t, event.ParticipantInvitedPayload{ParticipantID: "p1", Role: "E"}), event.ActorModerator),
	}
	events = append(events, preMeetingEvents(3, []string{"p1"})...)
	events = append(events,
		env(7, event.TypeRoundStarted, mustPayload(t, event.RoundStartedPayload{RoundNumber: 1, Order: []string{"p1"}}), event.ActorModerator),
		env(8, event.TypeMeetingPaused, mustPayload(t, event.MeetingPausedPayload{Reason: "break"}), event.ActorPrincipal),
		env(9, event.TypeMeetingResumed, nil, event.ActorPrincipal),
		env(10, event.TypeParticipantResponded, mustPayload(t, event.ParticipantRespondedPayload{
			ParticipantID: "p1", RoundNumber: 1, Content: "c", Stance: event.StanceAgree,
		}), event.ActorParticipant),
	)

	s, err := Fold("mtg-1", events)
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != StatusRunning {
		t.Fatalf("status = %s", s.Status)
	}
}

func TestApply_illegalAfterCompleted(t *testing.T) {
	t.Parallel()

	payload := mustPayload(t, event.MeetingCreatedPayload{Topic: "x"})
	s := State{ID: "mtg-1", Status: StatusCompleted}
	_, err := Apply(s, env(1, event.TypeMeetingCreated, payload, event.ActorPrincipal))
	if err == nil {
		t.Fatal("expected error applying to Completed meeting")
	}
}

func TestApply_participantResponded_duplicate(t *testing.T) {
	t.Parallel()

	s := State{
		ID:                  "mtg-1",
		Status:              StatusRunning,
		CurrentRound:        1,
		PreMeetingCompleted: true,
		RoundOrder:          []string{"p1"},
		RoundResponses: map[int]map[string]RoundResponse{
			1: {"p1": {Content: "already", Stance: event.StanceAgree}},
		},
	}
	payload := mustPayload(t, event.ParticipantRespondedPayload{
		ParticipantID: "p1", RoundNumber: 1, Content: "again", Stance: event.StanceAgree,
	})
	_, err := Apply(s, env(1, event.TypeParticipantResponded, payload, event.ActorParticipant))
	if err == nil {
		t.Fatal("expected duplicate response error")
	}
}

func TestApply_artifactAndActionItem(t *testing.T) {
	t.Parallel()

	s := State{ID: "mtg-1", Status: StatusRunning, Participants: map[string]ParticipantState{}}

	s, err := Apply(s, env(1, event.TypeArtifactProduced, mustPayload(t, event.ArtifactProducedPayload{
		ArtifactID: "a1", Type: "doc", Ref: "r",
	}), event.ActorModerator))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Artifacts) != 1 {
		t.Fatalf("artifacts = %d", len(s.Artifacts))
	}

	s, err = Apply(s, env(2, event.TypeActionItemCreated, mustPayload(t, event.ActionItemCreatedPayload{
		ActionItemID: "t1", Description: "follow up",
	}), event.ActorModerator))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.ActionItems) != 1 {
		t.Fatalf("action items = %d", len(s.ActionItems))
	}
}

func TestApply_multiRoundIncrement(t *testing.T) {
	t.Parallel()

	s := State{
		ID:                  "mtg-1",
		Status:              StatusRunning,
		CurrentRound:        1,
		PreMeetingCompleted: true,
		Participants:        map[string]ParticipantState{"p1": {ID: "p1"}},
		RoundOrder:   []string{"p1"},
		RoundResponses: map[int]map[string]RoundResponse{
			1: {"p1": {Stance: event.StanceObject, Content: "no"}},
		},
		Minutes: MinutesDraft{Rounds: []RoundSummary{{RoundNumber: 1, Summary: "s1"}}},
	}

	s, err := Apply(s, env(1, event.TypeRoundCompleted, mustPayload(t, event.RoundCompletedPayload{
		RoundNumber: 1, Summary: "s1",
	}), event.ActorModerator))
	if err != nil {
		t.Fatal(err)
	}

	s, err = Apply(s, env(2, event.TypeRoundStarted, mustPayload(t, event.RoundStartedPayload{
		RoundNumber: 2, Order: []string{"p1"},
	}), event.ActorModerator))
	if err != nil {
		t.Fatal(err)
	}
	if s.CurrentRound != 2 {
		t.Fatalf("round = %d", s.CurrentRound)
	}
}

func TestApply_moderatorSummarized_currentRound(t *testing.T) {
	t.Parallel()

	s := State{
		ID:                  "mtg-1",
		Status:              StatusRunning,
		CurrentRound:        1,
		PreMeetingCompleted: true,
	}
	s, err := Apply(s, env(1, event.TypeModeratorSummarized, mustPayload(t, event.ModeratorSummarizedPayload{
		RoundNumber: 1, Summary: "summary",
	}), event.ActorModerator))
	if err != nil {
		t.Fatal(err)
	}
	if s.ModeratorSummaries[1] != "summary" {
		t.Fatalf("got %q", s.ModeratorSummaries[1])
	}
}
