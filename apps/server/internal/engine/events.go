package engine

import (
	"encoding/json"
	"errors"

	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/event"
)

var errConfirmationNotImplemented = errors.New("engine: confirmation flow not implemented in v0.1")

func eventMeetingCreated(topic, confirmationMode string, maxRounds int) event.Envelope {
	payload, _ := json.Marshal(event.MeetingCreatedPayload{
		Topic:               topic,
		ConfirmationMode:    confirmationMode,
		MaxRoundsPerSegment: maxRounds,
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

func eventParticipantResponded(id string, round int, content string, stance event.Stance, objectReason string) event.Envelope {
	payload, _ := json.Marshal(event.ParticipantRespondedPayload{
		ParticipantID: id,
		RoundNumber:   round,
		Content:       content,
		Stance:        stance,
		ObjectReason:  objectReason,
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
