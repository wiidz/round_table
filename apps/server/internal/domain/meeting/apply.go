package meeting

import (
	"encoding/json"
	"fmt"

	"round_table/apps/server/internal/domain/event"
)

const (
	defaultMaxRoundsPerSegment      = 5
	defaultMaxConfirmationCycles    = 3
	defaultConsensusStrategy        = "no_objection"
	defaultFreeDialogueMaxQuestions   = 1
	defaultMinRoundsBeforeSynthesis   = 2
)

// Apply folds one event into state. Returns error if transition is illegal.
func Apply(s State, env event.Envelope) (State, error) {
	if s.isTerminal() {
		return s, fmt.Errorf("meeting %s: event %s not allowed in terminal status %s", s.ID, env.Type, s.Status)
	}

	switch env.Type {
	case event.TypeMeetingCreated:
		return applyMeetingCreated(s, env)
	case event.TypeParticipantInvited:
		return applyParticipantInvited(s, env)
	case event.TypeRoundStarted:
		return applyRoundStarted(s, env)
	case event.TypeParticipantResponded:
		return applyParticipantResponded(s, env)
	case event.TypeRoundCompleted:
		return applyRoundCompleted(s, env)
	case event.TypeModeratorSummarized:
		return applyModeratorSummarized(s, env)
	case event.TypeDeliberationReadinessChecked:
		return applyDeliberationReadinessChecked(s, env)
	case event.TypeFreeDialogueStarted:
		return applyFreeDialogueStarted(s, env)
	case event.TypeFreeDialogueQuestion:
		return applyFreeDialogueQuestionAsked(s, env)
	case event.TypeFreeDialogueAnswer:
		return applyFreeDialogueAnswered(s, env)
	case event.TypeFreeDialogueCompleted:
		return applyFreeDialogueCompleted(s, env)
	case event.TypeConsensusReached:
		return applyConsensusReached(s, env)
	case event.TypeSynthesisCompleted:
		return applySynthesisCompleted(s, env)
	case event.TypeSynthesisForced:
		return applySynthesisForced(s, env)
	case event.TypeConsensusVetoed:
		return applyConsensusVetoed(s, env)
	case event.TypeConsensusForced:
		return applyConsensusForced(s, env)
	case event.TypeConfirmationPrepared:
		return applyConfirmationPrepared(s, env)
	case event.TypeConfirmationPresented:
		return applyConfirmationPresented(s, env)
	case event.TypeConfirmationApproved:
		return applyConfirmationApproved(s, env)
	case event.TypeConfirmationRejected:
		return applyConfirmationRejected(s, env)
	case event.TypeConfirmationSkipped:
		return applyConfirmationSkipped(s, env)
	case event.TypeConfirmationForced:
		return applyConfirmationForced(s, env)
	case event.TypeMeetingPaused:
		return applyMeetingPaused(s, env)
	case event.TypeMeetingResumed:
		return applyMeetingResumed(s, env)
	case event.TypeMeetingFinished:
		return applyMeetingFinished(s, env)
	case event.TypeArtifactProduced:
		return applyArtifactProduced(s, env)
	case event.TypeActionItemCreated:
		return applyActionItemCreated(s, env)
	default:
		return s, fmt.Errorf("meeting %s: unsupported event type %s", s.ID, env.Type)
	}
}

// Fold replays events in order.
func Fold(meetingID string, events []event.Envelope) (State, error) {
	s := NewState(meetingID)
	for _, env := range events {
		var err error
		s, err = Apply(s, env)
		if err != nil {
			return s, err
		}
	}
	return s, nil
}

func decodePayload[T any](s State, env event.Envelope, eventName string) (T, error) {
	var p T
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return p, fmt.Errorf("meeting %s: decode %s: %w", s.ID, eventName, err)
	}
	return p, nil
}

func applyMeetingCreated(s State, env event.Envelope) (State, error) {
	if s.Status != StatusCreated {
		return s, fmt.Errorf("meeting %s: MeetingCreated not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.MeetingCreatedPayload](s, env, "MeetingCreated")
	if err != nil {
		return s, err
	}
	s.Topic = p.Topic
	s.Agenda = p.Agenda
	s.MeetingMode = p.MeetingMode
	if s.MeetingMode == "" {
		s.MeetingMode = MeetingModeDecision
	}
	s.ConsensusStrategy = p.ConsensusStrategy
	if s.ConsensusStrategy == "" {
		s.ConsensusStrategy = defaultConsensusStrategy
	}
	s.ConfirmationMode = p.ConfirmationMode
	if s.ConfirmationMode == "" {
		s.ConfirmationMode = ConfirmationModeRequired
	}
	s.MaxRoundsPerSegment = p.MaxRoundsPerSegment
	if s.MaxRoundsPerSegment <= 0 {
		s.MaxRoundsPerSegment = defaultMaxRoundsPerSegment
	}
	if p.MinRoundsBeforeSynthesis != nil {
		s.MinRoundsBeforeSynthesis = *p.MinRoundsBeforeSynthesis
	} else {
		s.MinRoundsBeforeSynthesis = defaultMinRoundsBeforeSynthesis
	}
	if s.MinRoundsBeforeSynthesis <= 0 {
		s.MinRoundsBeforeSynthesis = defaultMinRoundsBeforeSynthesis
	}
	s.MaxConfirmationCycles = p.MaxConfirmationCycles
	if s.MaxConfirmationCycles <= 0 {
		s.MaxConfirmationCycles = defaultMaxConfirmationCycles
	}
	if p.FreeDialogueMaxQuestions != nil {
		s.FreeDialogueMaxQuestions = *p.FreeDialogueMaxQuestions
	} else {
		s.FreeDialogueMaxQuestions = defaultFreeDialogueMaxQuestions
	}
	s.StartedAt = env.OccurredAt
	s.Goal = p.Goal
	if s.Goal == "" {
		s.Goal = defaultMeetingGoal(s.Topic, s.MeetingMode)
	}
	s.Status = StatusPreparing
	return s, nil
}

func defaultMeetingGoal(topic, mode string) string {
	if mode == MeetingModeDeliberation {
		return "围绕「" + topic + "」形成可评审的方案草案，并列出待决事项。"
	}
	return "围绕「" + topic + "」达成可执行的共识，并明确后续行动项。"
}

func applyParticipantInvited(s State, env event.Envelope) (State, error) {
	if s.Status != StatusPreparing {
		return s, fmt.Errorf("meeting %s: ParticipantInvited not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.ParticipantInvitedPayload](s, env, "ParticipantInvited")
	if err != nil {
		return s, err
	}
	if p.ParticipantID == "" {
		return s, fmt.Errorf("meeting %s: ParticipantInvited missing participant_id", s.ID)
	}
	if _, exists := s.Participants[p.ParticipantID]; exists {
		return s, fmt.Errorf("meeting %s: participant %s already invited", s.ID, p.ParticipantID)
	}
	s.Participants[p.ParticipantID] = ParticipantState{
		ID:        p.ParticipantID,
		Role:      p.Role,
		Expertise: p.Expertise,
		Goal:      p.Goal,
	}
	s.ParticipantOrder = append(s.ParticipantOrder, p.ParticipantID)
	return s, nil
}

func applyRoundStarted(s State, env event.Envelope) (State, error) {
	if s.Status == StatusPaused {
		return s, fmt.Errorf("meeting %s: RoundStarted not allowed while Paused", s.ID)
	}
	if s.Status != StatusPreparing && s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: RoundStarted not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.RoundStartedPayload](s, env, "RoundStarted")
	if err != nil {
		return s, err
	}
	if p.RoundNumber < 0 {
		return s, fmt.Errorf("meeting %s: RoundStarted invalid round_number %d", s.ID, p.RoundNumber)
	}
	if len(p.Order) == 0 {
		return s, fmt.Errorf("meeting %s: RoundStarted empty order", s.ID)
	}
	expected, err := expectedRoundStart(s)
	if err != nil {
		return s, err
	}
	if p.RoundNumber != expected {
		return s, fmt.Errorf("meeting %s: RoundStarted round %d, want %d", s.ID, p.RoundNumber, expected)
	}
	for _, id := range p.Order {
		if _, ok := s.Participants[id]; !ok {
			return s, fmt.Errorf("meeting %s: RoundStarted unknown participant %s", s.ID, id)
		}
	}

	s.CurrentRound = p.RoundNumber
	s.RoundOrder = append([]string(nil), p.Order...)
	s.RoundResponses[p.RoundNumber] = make(map[string]RoundResponse)
	s.Status = StatusRunning
	return s, nil
}

func applyParticipantResponded(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: ParticipantResponded not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.ParticipantRespondedPayload](s, env, "ParticipantResponded")
	if err != nil {
		return s, err
	}
	if p.RoundNumber != s.CurrentRound {
		return s, fmt.Errorf("meeting %s: ParticipantResponded round %d, current %d", s.ID, p.RoundNumber, s.CurrentRound)
	}
	if !participantInOrder(s.RoundOrder, p.ParticipantID) {
		return s, fmt.Errorf("meeting %s: ParticipantResponded unknown participant %s in round", s.ID, p.ParticipantID)
	}
	responses := s.RoundResponses[p.RoundNumber]
	if responses == nil {
		responses = make(map[string]RoundResponse)
		s.RoundResponses[p.RoundNumber] = responses
	}
	if _, dup := responses[p.ParticipantID]; dup {
		return s, fmt.Errorf("meeting %s: participant %s already responded in round %d", s.ID, p.ParticipantID, p.RoundNumber)
	}
	if p.RoundNumber == 0 {
		if p.Stance != "" && p.Stance != event.StanceNone {
			return s, fmt.Errorf("meeting %s: pre-meeting round 0 requires stance none", s.ID)
		}
	} else if s.IsDeliberation() {
		if p.Stance != "" && p.Stance != event.StanceNone {
			return s, fmt.Errorf("meeting %s: deliberation round %d requires stance none", s.ID, p.RoundNumber)
		}
		if p.Stance == "" {
			p.Stance = event.StanceNone
		}
	} else if p.Stance == event.StanceNone || p.Stance == "" {
		return s, fmt.Errorf("meeting %s: debate round %d requires agree, object, or abstain", s.ID, p.RoundNumber)
	} else if p.Stance == event.StanceObject && p.ObjectReason == "" {
		return s, fmt.Errorf("meeting %s: object stance requires object_reason", s.ID)
	}
	responses[p.ParticipantID] = RoundResponse{
		Content:      p.Content,
		Stance:       p.Stance,
		ObjectReason: p.ObjectReason,
	}
	if p.TokenUsage != nil {
		s = recordTokenUsage(s, *p.TokenUsage)
	}
	return s, nil
}

func applyRoundCompleted(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: RoundCompleted not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.RoundCompletedPayload](s, env, "RoundCompleted")
	if err != nil {
		return s, err
	}
	if p.RoundNumber != s.CurrentRound {
		return s, fmt.Errorf("meeting %s: RoundCompleted round %d, current %d", s.ID, p.RoundNumber, s.CurrentRound)
	}
	s.Minutes.Rounds = append(s.Minutes.Rounds, RoundSummary{
		RoundNumber: p.RoundNumber,
		Summary:     p.Summary,
	})
	if p.RoundNumber == 0 {
		s.PreMeetingCompleted = true
		s.PreMeetingSummary = p.Summary
	}
	return s, nil
}

func applyModeratorSummarized(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: ModeratorSummarized not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.ModeratorSummarizedPayload](s, env, "ModeratorSummarized")
	if err != nil {
		return s, err
	}
	if p.RoundNumber <= 0 || p.RoundNumber > s.CurrentRound {
		return s, fmt.Errorf("meeting %s: ModeratorSummarized invalid round %d for current %d", s.ID, p.RoundNumber, s.CurrentRound)
	}
	if s.ModeratorSummaries == nil {
		s.ModeratorSummaries = make(map[int]string)
	}
	s.ModeratorSummaries[p.RoundNumber] = p.Summary
	return s, nil
}

func applyDeliberationReadinessChecked(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: DeliberationReadinessChecked not allowed in status %s", s.ID, s.Status)
	}
	if !s.IsDeliberation() {
		return s, fmt.Errorf("meeting %s: DeliberationReadinessChecked requires meeting_mode deliberation", s.ID)
	}
	p, err := decodePayload[event.DeliberationReadinessCheckedPayload](s, env, "DeliberationReadinessChecked")
	if err != nil {
		return s, err
	}
	if p.RoundNumber <= 0 || p.RoundNumber > s.CurrentRound {
		return s, fmt.Errorf("meeting %s: DeliberationReadinessChecked invalid round %d for current %d", s.ID, p.RoundNumber, s.CurrentRound)
	}
	if p.TokenUsage != nil {
		s = recordTokenUsage(s, *p.TokenUsage)
	}
	return s, nil
}

func applyFreeDialogueStarted(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: FreeDialogueStarted not allowed in status %s", s.ID, s.Status)
	}
	if s.CurrentRound != 1 {
		return s, fmt.Errorf("meeting %s: FreeDialogueStarted only after round 1, current %d", s.ID, s.CurrentRound)
	}
	if s.FreeDialogueCompleted || s.InFreeDialogue {
		return s, fmt.Errorf("meeting %s: FreeDialogueStarted already started or completed", s.ID)
	}
	p, err := decodePayload[event.FreeDialogueStartedPayload](s, env, "FreeDialogueStarted")
	if err != nil {
		return s, err
	}
	if p.AfterRound != 1 {
		return s, fmt.Errorf("meeting %s: FreeDialogueStarted after_round %d, want 1", s.ID, p.AfterRound)
	}
	if p.MaxQuestions <= 0 {
		return s, fmt.Errorf("meeting %s: FreeDialogueStarted max_questions must be positive", s.ID)
	}
	s.InFreeDialogue = true
	s.FreeDialogueQuestionIndex = 0
	s.FreeDialogueAskerIndex = 0
	s.FreeDialogueMaxQuestions = p.MaxQuestions
	return s, nil
}

func applyFreeDialogueQuestionAsked(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning || !s.InFreeDialogue {
		return s, fmt.Errorf("meeting %s: FreeDialogueQuestionAsked not in free dialogue", s.ID)
	}
	if s.PendingFreeDialogue != nil {
		return s, fmt.Errorf("meeting %s: FreeDialogueQuestionAsked while answer pending", s.ID)
	}
	p, err := decodePayload[event.FreeDialogueQuestionAskedPayload](s, env, "FreeDialogueQuestionAsked")
	if err != nil {
		return s, err
	}
	if p.QuestionIndex != s.FreeDialogueQuestionIndex {
		return s, fmt.Errorf("meeting %s: FreeDialogueQuestionAsked index %d, want %d", s.ID, p.QuestionIndex, s.FreeDialogueQuestionIndex)
	}
	if p.PrincipalMediated {
		if p.AskerID != PrincipalRelayAskerID {
			return s, fmt.Errorf("meeting %s: FreeDialogueQuestionAsked principal_mediated asker %q, want %q", s.ID, p.AskerID, PrincipalRelayAskerID)
		}
		if !participantInOrder(s.ParticipantOrder, p.AnswererID) {
			return s, fmt.Errorf("meeting %s: FreeDialogueQuestionAsked unknown answerer", s.ID)
		}
	} else if !participantInOrder(s.ParticipantOrder, p.AskerID) || !participantInOrder(s.ParticipantOrder, p.AnswererID) {
		return s, fmt.Errorf("meeting %s: FreeDialogueQuestionAsked unknown participant", s.ID)
	}
	if p.Content == "" {
		return s, fmt.Errorf("meeting %s: FreeDialogueQuestionAsked empty content", s.ID)
	}
	s.PendingFreeDialogue = &PendingFreeDialogue{
		AskerID:           p.AskerID,
		AnswererID:        p.AnswererID,
		QuestionIndex:     p.QuestionIndex,
		Question:          p.Content,
		PrincipalMediated: p.PrincipalMediated,
	}
	if p.TokenUsage != nil {
		s = recordTokenUsage(s, *p.TokenUsage)
	}
	return s, nil
}

func applyFreeDialogueAnswered(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning || !s.InFreeDialogue {
		return s, fmt.Errorf("meeting %s: FreeDialogueAnswered not in free dialogue", s.ID)
	}
	if s.PendingFreeDialogue == nil {
		return s, fmt.Errorf("meeting %s: FreeDialogueAnswered without pending question", s.ID)
	}
	p, err := decodePayload[event.FreeDialogueAnsweredPayload](s, env, "FreeDialogueAnswered")
	if err != nil {
		return s, err
	}
	pending := s.PendingFreeDialogue
	if p.AskerID != pending.AskerID || p.AnswererID != pending.AnswererID || p.QuestionIndex != pending.QuestionIndex {
		return s, fmt.Errorf("meeting %s: FreeDialogueAnswered mismatch with pending question", s.ID)
	}
	if p.Answer == "" {
		return s, fmt.Errorf("meeting %s: FreeDialogueAnswered empty answer", s.ID)
	}
	s.FreeDialogueExchanges = append(s.FreeDialogueExchanges, FreeDialogueExchange{
		QuestionIndex:     p.QuestionIndex,
		AskerID:           p.AskerID,
		AnswererID:        p.AnswererID,
		Question:          p.Question,
		Answer:            p.Answer,
		PrincipalMediated: pending.PrincipalMediated,
	})
	s.PendingFreeDialogue = nil
	s.FreeDialogueQuestionIndex++
	s.FreeDialogueAskerIndex = (s.FreeDialogueAskerIndex + 1) % len(s.ParticipantOrder)
	if p.TokenUsage != nil {
		s = recordTokenUsage(s, *p.TokenUsage)
	}
	return s, nil
}

func applyFreeDialogueCompleted(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning || !s.InFreeDialogue {
		return s, fmt.Errorf("meeting %s: FreeDialogueCompleted not in free dialogue", s.ID)
	}
	if s.PendingFreeDialogue != nil {
		return s, fmt.Errorf("meeting %s: FreeDialogueCompleted with pending question", s.ID)
	}
	p, err := decodePayload[event.FreeDialogueCompletedPayload](s, env, "FreeDialogueCompleted")
	if err != nil {
		return s, err
	}
	if p.AfterRound != 1 {
		return s, fmt.Errorf("meeting %s: FreeDialogueCompleted after_round %d, want 1", s.ID, p.AfterRound)
	}
	s.InFreeDialogue = false
	s.FreeDialogueCompleted = true
	s.FreeDialogueSummary = p.Summary
	return s, nil
}

func expectedRoundStart(s State) (int, error) {
	if !s.PreMeetingCompleted {
		return 0, nil
	}
	return s.CurrentRound + 1, nil
}

func applyConsensusReached(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: ConsensusReached not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.ConsensusReachedPayload](s, env, "ConsensusReached")
	if err != nil {
		return s, err
	}
	s.Consensus = &ConsensusState{
		Strategy:   p.Strategy,
		ResolvedBy: p.ResolvedBy,
		Dissent:    append([]event.DissentingOpinion(nil), p.Dissent...),
	}
	s.Status = StatusConsensus
	return s, nil
}

func applySynthesisCompleted(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: SynthesisCompleted not allowed in status %s", s.ID, s.Status)
	}
	if !s.IsDeliberation() {
		return s, fmt.Errorf("meeting %s: SynthesisCompleted requires meeting_mode deliberation", s.ID)
	}
	p, err := decodePayload[event.SynthesisCompletedPayload](s, env, "SynthesisCompleted")
	if err != nil {
		return s, err
	}
	s.SynthesisSummary = p.Summary
	s.SynthesisOpenQuestions = append([]string(nil), p.OpenQuestions...)
	s.SynthesisSections = append([]event.SynthesisAgendaSectionPayload(nil), p.Sections...)
	if p.CrossCutting != nil {
		cc := *p.CrossCutting
		s.SynthesisCrossCutting = &cc
	} else {
		s.SynthesisCrossCutting = nil
	}
	if p.TokenUsage != nil {
		s = recordTokenUsage(s, *p.TokenUsage)
	}
	resolvedBy := p.ResolvedBy
	if resolvedBy == "" {
		resolvedBy = "synthesis"
	}
	s.Consensus = &ConsensusState{
		Strategy:   MeetingModeDeliberation,
		ResolvedBy: resolvedBy,
	}
	s.Status = StatusConsensus
	return s, nil
}

func applyConsensusVetoed(s State, env event.Envelope) (State, error) {
	if s.Status != StatusConsensus {
		return s, fmt.Errorf("meeting %s: ConsensusVetoed not allowed in status %s", s.ID, s.Status)
	}
	if _, err := decodePayload[event.ConsensusVetoedPayload](s, env, "ConsensusVetoed"); err != nil {
		return s, err
	}
	s.Consensus = nil
	s.Confirmation = nil
	s.Status = StatusRunning
	return s, nil
}

func applyConsensusForced(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: ConsensusForced not allowed in status %s", s.ID, s.Status)
	}
	if s.IsDeliberation() {
		return s, fmt.Errorf("meeting %s: ConsensusForced not allowed in deliberation mode; use SynthesisForced", s.ID)
	}
	p, err := decodePayload[event.ConsensusForcedPayload](s, env, "ConsensusForced")
	if err != nil {
		return s, err
	}
	s.Consensus = &ConsensusState{
		Strategy:   s.ConsensusStrategy,
		ResolvedBy: "principal",
		Dissent:    nil,
	}
	_ = p
	s.Status = StatusConsensus
	return s, nil
}

func applySynthesisForced(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning {
		return s, fmt.Errorf("meeting %s: SynthesisForced not allowed in status %s", s.ID, s.Status)
	}
	if !s.IsDeliberation() {
		return s, fmt.Errorf("meeting %s: SynthesisForced requires meeting_mode deliberation", s.ID)
	}
	if _, err := decodePayload[event.SynthesisForcedPayload](s, env, "SynthesisForced"); err != nil {
		return s, err
	}
	return s, nil
}

func applyConfirmationPrepared(s State, env event.Envelope) (State, error) {
	if s.Status != StatusConsensus {
		return s, fmt.Errorf("meeting %s: ConfirmationPrepared not allowed in status %s", s.ID, s.Status)
	}
	if s.ConfirmationMode == ConfirmationModeSkip {
		return s, fmt.Errorf("meeting %s: ConfirmationPrepared with confirmation_mode skip", s.ID)
	}
	p, err := decodePayload[event.ConfirmationPreparedPayload](s, env, "ConfirmationPrepared")
	if err != nil {
		return s, err
	}
	s.Confirmation = &ConfirmationState{
		Cycle: p.Cycle,
		Brief: p.Brief,
	}
	return s, nil
}

func applyConfirmationPresented(s State, env event.Envelope) (State, error) {
	if s.Status != StatusConsensus {
		return s, fmt.Errorf("meeting %s: ConfirmationPresented not allowed in status %s", s.ID, s.Status)
	}
	if s.ConfirmationMode == ConfirmationModeSkip {
		return s, fmt.Errorf("meeting %s: ConfirmationPresented with confirmation_mode skip", s.ID)
	}
	p, err := decodePayload[event.ConfirmationPresentedPayload](s, env, "ConfirmationPresented")
	if err != nil {
		return s, err
	}
	if s.Confirmation == nil {
		s.Confirmation = &ConfirmationState{Cycle: p.Cycle}
	} else if s.Confirmation.Cycle != p.Cycle {
		return s, fmt.Errorf("meeting %s: ConfirmationPresented cycle %d, want %d", s.ID, p.Cycle, s.Confirmation.Cycle)
	}
	s.Status = StatusConfirmation
	return s, nil
}

func applyConfirmationApproved(s State, env event.Envelope) (State, error) {
	if s.Status != StatusConfirmation {
		return s, fmt.Errorf("meeting %s: ConfirmationApproved not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.ConfirmationApprovedPayload](s, env, "ConfirmationApproved")
	if err != nil {
		return s, err
	}
	if s.Confirmation == nil {
		return s, fmt.Errorf("meeting %s: ConfirmationApproved without prepared confirmation", s.ID)
	}
	if s.Confirmation.Cycle != p.Cycle {
		return s, fmt.Errorf("meeting %s: ConfirmationApproved cycle %d, want %d", s.ID, p.Cycle, s.Confirmation.Cycle)
	}
	s.Confirmation.Approved = true
	s.Confirmation.ItemNotes = p.ItemNotes
	return s, nil
}

func applyConfirmationRejected(s State, env event.Envelope) (State, error) {
	if s.Status != StatusConfirmation {
		return s, fmt.Errorf("meeting %s: ConfirmationRejected not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.ConfirmationRejectedPayload](s, env, "ConfirmationRejected")
	if err != nil {
		return s, err
	}
	if p.Feedback == "" {
		return s, fmt.Errorf("meeting %s: ConfirmationRejected requires feedback", s.ID)
	}
	if s.Confirmation != nil && s.Confirmation.Cycle != p.Cycle {
		return s, fmt.Errorf("meeting %s: ConfirmationRejected cycle %d, want %d", s.ID, p.Cycle, s.Confirmation.Cycle)
	}

	if p.ResetCycle {
		s.ConfirmationCycle = 0
	} else {
		s.ConfirmationCycle++
	}
	s.PrincipalFeedback = p.Feedback
	s.Consensus = nil
	s.Confirmation = nil
	s.SynthesisSummary = ""
	s.SynthesisOpenQuestions = nil
	s.SynthesisSections = nil
	s.SynthesisCrossCutting = nil
	if s.CurrentRound >= s.MaxRoundsPerSegment {
		s.MaxRoundsPerSegment = s.CurrentRound + 1
	}
	s.Status = StatusRunning
	return s, nil
}

func applyConfirmationSkipped(s State, env event.Envelope) (State, error) {
	if s.Status != StatusConsensus && s.Status != StatusConfirmation {
		return s, fmt.Errorf("meeting %s: ConfirmationSkipped not allowed in status %s", s.ID, s.Status)
	}
	if _, err := decodePayload[event.ConfirmationSkippedPayload](s, env, "ConfirmationSkipped"); err != nil {
		return s, err
	}
	s.ConfirmationMode = ConfirmationModeSkip
	s.Confirmation = nil
	if s.Status == StatusConfirmation {
		s.Status = StatusConsensus
	}
	return s, nil
}

func applyConfirmationForced(s State, env event.Envelope) (State, error) {
	if s.Status != StatusConfirmation {
		return s, fmt.Errorf("meeting %s: ConfirmationForced not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.ConfirmationForcedPayload](s, env, "ConfirmationForced")
	if err != nil {
		return s, err
	}
	if s.Confirmation == nil {
		s.Confirmation = &ConfirmationState{Cycle: p.Cycle}
	}
	s.Confirmation.Approved = true
	return s, nil
}

func applyMeetingPaused(s State, env event.Envelope) (State, error) {
	if s.Status != StatusRunning && s.Status != StatusConfirmation {
		return s, fmt.Errorf("meeting %s: MeetingPaused not allowed in status %s", s.ID, s.Status)
	}
	if _, err := decodePayload[event.MeetingPausedPayload](s, env, "MeetingPaused"); err != nil {
		return s, err
	}
	s.PausedFrom = s.Status
	s.Status = StatusPaused
	return s, nil
}

func applyMeetingResumed(s State, env event.Envelope) (State, error) {
	if s.Status != StatusPaused {
		return s, fmt.Errorf("meeting %s: MeetingResumed not allowed in status %s", s.ID, s.Status)
	}
	if env.Payload != nil && len(env.Payload) > 0 {
		// MeetingResumed has empty payload in ADR-0003; tolerate {} .
	}
	if s.PausedFrom == "" {
		s.PausedFrom = StatusRunning
	}
	s.Status = s.PausedFrom
	s.PausedFrom = ""
	return s, nil
}

func applyMeetingFinished(s State, env event.Envelope) (State, error) {
	p, err := decodePayload[event.MeetingFinishedPayload](s, env, "MeetingFinished")
	if err != nil {
		return s, err
	}
	switch s.Status {
	case StatusConsensus, StatusConfirmation:
		// expected path
	case StatusRunning:
		if p.Outcome != OutcomeAborted {
			return s, fmt.Errorf("meeting %s: MeetingFinished from Running requires outcome aborted", s.ID)
		}
	default:
		return s, fmt.Errorf("meeting %s: MeetingFinished not allowed in status %s", s.ID, s.Status)
	}
	if p.Outcome == "" {
		p.Outcome = OutcomeCompleted
	}
	s.Outcome = p.Outcome
	s.Status = StatusCompleted
	return s, nil
}

func applyArtifactProduced(s State, env event.Envelope) (State, error) {
	if s.Status == StatusCreated || s.isTerminal() {
		return s, fmt.Errorf("meeting %s: ArtifactProduced not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.ArtifactProducedPayload](s, env, "ArtifactProduced")
	if err != nil {
		return s, err
	}
	s.Artifacts = append(s.Artifacts, ArtifactRef{
		ID:   p.ArtifactID,
		Type: p.Type,
		Ref:  p.Ref,
	})
	return s, nil
}

func applyActionItemCreated(s State, env event.Envelope) (State, error) {
	if s.Status == StatusCreated || s.isTerminal() {
		return s, fmt.Errorf("meeting %s: ActionItemCreated not allowed in status %s", s.ID, s.Status)
	}
	p, err := decodePayload[event.ActionItemCreatedPayload](s, env, "ActionItemCreated")
	if err != nil {
		return s, err
	}
	s.ActionItems = append(s.ActionItems, ActionItem{
		ID:          p.ActionItemID,
		Assignee:    p.Assignee,
		Description: p.Description,
	})
	return s, nil
}

func participantInOrder(order []string, id string) bool {
	for _, pid := range order {
		if pid == id {
			return true
		}
	}
	return false
}

func recordTokenUsage(s State, u event.TokenUsage) State {
	if u.TotalTokens == 0 && u.PromptTokens == 0 && u.CompletionTokens == 0 {
		return s
	}
	turn := len(s.TokenUsageLog) + 1
	s.TokenUsageLog = append(s.TokenUsageLog, TokenUsageRecord{
		Turn:             turn,
		Phase:            u.Phase,
		ParticipantID:    u.ParticipantID,
		Model:            u.Model,
		RoundNumber:      u.RoundNumber,
		QuestionIndex:    u.QuestionIndex,
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	})
	s.TokenUsageTotals.CallCount++
	s.TokenUsageTotals.PromptTokens += u.PromptTokens
	s.TokenUsageTotals.CompletionTokens += u.CompletionTokens
	s.TokenUsageTotals.TotalTokens += u.TotalTokens
	return s
}
