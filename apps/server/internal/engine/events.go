package engine

import (
	"encoding/json"

	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

type meetingCreatedParams struct {
	Topic                    string
	Goal                     string
	MeetingMode              string
	ConfirmationMode         string
	MaxRoundsPerSegment      int
	MinRoundsBeforeSynthesis int
	FreeDialogueMaxQuestions int
	MaxConfirmationCycles    int
	Agenda                   []event.AgendaItem
}

func eventMeetingCreated(p meetingCreatedParams) event.Envelope {
	q := p.FreeDialogueMaxQuestions
	minRounds := p.MinRoundsBeforeSynthesis
	payload, _ := json.Marshal(event.MeetingCreatedPayload{
		Topic:                    p.Topic,
		Goal:                     p.Goal,
		MeetingMode:              p.MeetingMode,
		Agenda:                   p.Agenda,
		ConfirmationMode:         p.ConfirmationMode,
		MaxRoundsPerSegment:      p.MaxRoundsPerSegment,
		MinRoundsBeforeSynthesis:   &minRounds,
		MaxConfirmationCycles:      p.MaxConfirmationCycles,
		FreeDialogueMaxQuestions: &q,
	})
	return event.Envelope{
		Type:    event.TypeMeetingCreated,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}

func eventParticipantInvited(p ParticipantInput) event.Envelope {
	payload, _ := json.Marshal(event.ParticipantInvitedPayload{
		ParticipantID: p.ID,
		Role:          p.Role,
		Expertise:     p.Expertise,
		Goal:          p.Goal,
	})
	return event.Envelope{
		Type:    event.TypeParticipantInvited,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventRoundStarted(roundNumber int, order []string) event.Envelope {
	payload, _ := json.Marshal(event.RoundStartedPayload{
		RoundNumber: roundNumber,
		Order:       order,
	})
	return event.Envelope{
		Type:    event.TypeRoundStarted,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventParticipantResponded(id string, round int, content string, stance event.Stance, objectReason string, usage *event.TokenUsage) event.Envelope {
	payload, _ := json.Marshal(event.ParticipantRespondedPayload{
		ParticipantID: id,
		RoundNumber:   round,
		Content:       content,
		Stance:        stance,
		ObjectReason:  objectReason,
		TokenUsage:    usage,
	})
	return event.Envelope{
		Type:    event.TypeParticipantResponded,
		Payload: payload,
		Actor:   event.ActorParticipant,
	}
}

func eventRoundCompleted(roundNumber int, summary string) event.Envelope {
	payload, _ := json.Marshal(event.RoundCompletedPayload{
		RoundNumber: roundNumber,
		Summary:     summary,
	})
	return event.Envelope{
		Type:    event.TypeRoundCompleted,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventModeratorSummarized(roundNumber int, summary string) event.Envelope {
	payload, _ := json.Marshal(event.ModeratorSummarizedPayload{
		RoundNumber: roundNumber,
		Summary:     summary,
	})
	return event.Envelope{
		Type:    event.TypeModeratorSummarized,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventDeliberationReadinessChecked(roundNumber int, ready bool, rationale string, gaps []string, usage *event.TokenUsage) event.Envelope {
	payload, _ := json.Marshal(event.DeliberationReadinessCheckedPayload{
		RoundNumber: roundNumber,
		Ready:       ready,
		Rationale:   rationale,
		Gaps:        gaps,
		TokenUsage:  usage,
	})
	return event.Envelope{
		Type:    event.TypeDeliberationReadinessChecked,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventFreeDialogueStarted(afterRound, maxQuestions int) event.Envelope {
	payload, _ := json.Marshal(event.FreeDialogueStartedPayload{
		AfterRound:   afterRound,
		MaxQuestions: maxQuestions,
	})
	return event.Envelope{
		Type:    event.TypeFreeDialogueStarted,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventFreeDialogueQuestionAsked(askerID, answererID string, questionIndex int, content string, usage *event.TokenUsage) event.Envelope {
	payload, _ := json.Marshal(event.FreeDialogueQuestionAskedPayload{
		AskerID:       askerID,
		AnswererID:    answererID,
		QuestionIndex: questionIndex,
		Content:       content,
		TokenUsage:    usage,
	})
	return event.Envelope{
		Type:    event.TypeFreeDialogueQuestion,
		Payload: payload,
		Actor:   event.ActorParticipant,
	}
}

func eventFreeDialogueQuestionAskedByPrincipal(answererID string, questionIndex int, content string) event.Envelope {
	payload, _ := json.Marshal(event.FreeDialogueQuestionAskedPayload{
		AskerID:           meeting.PrincipalRelayAskerID,
		AnswererID:        answererID,
		QuestionIndex:     questionIndex,
		Content:           content,
		PrincipalMediated: true,
	})
	return event.Envelope{
		Type:    event.TypeFreeDialogueQuestion,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}

func eventFreeDialogueAnswered(askerID, answererID string, questionIndex int, question, answer string, usage *event.TokenUsage) event.Envelope {
	payload, _ := json.Marshal(event.FreeDialogueAnsweredPayload{
		AskerID:       askerID,
		AnswererID:    answererID,
		QuestionIndex: questionIndex,
		Question:      question,
		Answer:        answer,
		TokenUsage:    usage,
	})
	return event.Envelope{
		Type:    event.TypeFreeDialogueAnswer,
		Payload: payload,
		Actor:   event.ActorParticipant,
	}
}

func eventFreeDialogueCompleted(afterRound int, summary string) event.Envelope {
	payload, _ := json.Marshal(event.FreeDialogueCompletedPayload{
		AfterRound: afterRound,
		Summary:    summary,
	})
	return event.Envelope{
		Type:    event.TypeFreeDialogueCompleted,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventConsensusReached(strategy string, res consensus.Result) event.Envelope {
	dissent := make([]event.DissentingOpinion, len(res.Dissent))
	for i, d := range res.Dissent {
		dissent[i] = event.DissentingOpinion{ParticipantID: d.ParticipantID, Reason: d.Reason}
	}
	payload, _ := json.Marshal(event.ConsensusReachedPayload{
		Strategy:   strategy,
		Dissent:    dissent,
		ResolvedBy: res.ResolvedBy,
	})
	return event.Envelope{
		Type:    event.TypeConsensusReached,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventModeratorDecision(strategy string) event.Envelope {
	payload, _ := json.Marshal(event.ConsensusReachedPayload{
		Strategy:   strategy,
		ResolvedBy: "moderator",
	})
	return event.Envelope{
		Type:    event.TypeConsensusReached,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventSynthesisCompleted(
	summary string,
	openQuestions []string,
	resolvedBy string,
	usage *event.TokenUsage,
	sections []event.SynthesisAgendaSectionPayload,
	cross *event.SynthesisCrossCuttingPayload,
) event.Envelope {
	payload, _ := json.Marshal(event.SynthesisCompletedPayload{
		Summary:       summary,
		OpenQuestions: openQuestions,
		ResolvedBy:    resolvedBy,
		TokenUsage:    usage,
		Sections:      sections,
		CrossCutting:  cross,
	})
	return event.Envelope{
		Type:    event.TypeSynthesisCompleted,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventSynthesisForced(reason string) event.Envelope {
	payload, _ := json.Marshal(event.SynthesisForcedPayload{Reason: reason})
	return event.Envelope{
		Type:    event.TypeSynthesisForced,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}

func eventConsensusForced(reason string) event.Envelope {
	payload, _ := json.Marshal(event.ConsensusForcedPayload{Reason: reason})
	return event.Envelope{
		Type:    event.TypeConsensusForced,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}

func eventArtifactProduced(id, typ, ref string) event.Envelope {
	payload, _ := json.Marshal(event.ArtifactProducedPayload{
		ArtifactID: id,
		Type:       typ,
		Ref:        ref,
	})
	return event.Envelope{
		Type:    event.TypeArtifactProduced,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventMeetingFinished(outcome string) event.Envelope {
	payload, _ := json.Marshal(event.MeetingFinishedPayload{Outcome: outcome})
	return event.Envelope{
		Type:    event.TypeMeetingFinished,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventMeetingAborted(_ string) event.Envelope {
	payload, _ := json.Marshal(event.MeetingFinishedPayload{Outcome: meeting.OutcomeAborted})
	return event.Envelope{
		Type:    event.TypeMeetingFinished,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}

func eventMeetingPaused(reason string) event.Envelope {
	payload, _ := json.Marshal(event.MeetingPausedPayload{Reason: reason})
	return event.Envelope{
		Type:    event.TypeMeetingPaused,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}

func eventMeetingResumed() event.Envelope {
	return event.Envelope{
		Type:  event.TypeMeetingResumed,
		Actor: event.ActorPrincipal,
	}
}

func eventConfirmationPrepared(cycle int, brief event.ConfirmationBrief) event.Envelope {
	payload, _ := json.Marshal(event.ConfirmationPreparedPayload{Cycle: cycle, Brief: brief})
	return event.Envelope{
		Type:    event.TypeConfirmationPrepared,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventConfirmationPresented(cycle int) event.Envelope {
	payload, _ := json.Marshal(event.ConfirmationPresentedPayload{Cycle: cycle})
	return event.Envelope{
		Type:    event.TypeConfirmationPresented,
		Payload: payload,
		Actor:   event.ActorModerator,
	}
}

func eventConfirmationApproved(cycle int, notes map[int]string) event.Envelope {
	payload, _ := json.Marshal(event.ConfirmationApprovedPayload{Cycle: cycle, ItemNotes: notes})
	return event.Envelope{
		Type:    event.TypeConfirmationApproved,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}

func eventConfirmationRejected(cycle int, feedback string, notes map[int]string) event.Envelope {
	payload, _ := json.Marshal(event.ConfirmationRejectedPayload{
		Cycle: cycle, Feedback: feedback, ItemNotes: notes,
	})
	return event.Envelope{
		Type:    event.TypeConfirmationRejected,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}

func eventConfirmationRejectedResetCycle(cycle int, feedback string, notes map[int]string) event.Envelope {
	payload, _ := json.Marshal(event.ConfirmationRejectedPayload{
		Cycle: cycle, Feedback: feedback, ItemNotes: notes, ResetCycle: true,
	})
	return event.Envelope{
		Type:    event.TypeConfirmationRejected,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}

func eventConfirmationForced(cycle int, reason string) event.Envelope {
	payload, _ := json.Marshal(event.ConfirmationForcedPayload{Cycle: cycle, Reason: reason})
	return event.Envelope{
		Type:    event.TypeConfirmationForced,
		Payload: payload,
		Actor:   event.ActorPrincipal,
	}
}
