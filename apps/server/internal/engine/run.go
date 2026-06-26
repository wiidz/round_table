package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/knowledge"
	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/scheduler"
)

func (e *Engine) startRound(ctx context.Context, s meeting.State) (meeting.State, error) {
	order := append([]string(nil), s.ParticipantOrder...)
	return e.append(ctx, s, eventRoundStarted(s.CurrentRound+1, order))
}

func (e *Engine) advanceRunning(ctx context.Context, s meeting.State) (meeting.State, error) {
	spoken := spokenInRound(s)
	if next, ok := scheduler.FixedOrder(s.RoundOrder, spoken); ok {
		return e.inviteSpeak(ctx, s, next)
	}
	return e.completeRound(ctx, s)
}

func (e *Engine) inviteSpeak(ctx context.Context, s meeting.State, participantID string) (meeting.State, error) {
	prompt := e.buildPrompt(s, participantID)
	resp, err := e.Participant.Respond(ctx, s.ID, participantID, prompt)
	if err != nil {
		return s, err
	}
	stance := event.Stance(resp.Stance)
	if stance == "" {
		stance = event.StanceAgree
	}
	return e.append(ctx, s, eventParticipantResponded(participantID, s.CurrentRound, resp.Content, stance, resp.ObjectReason))
}

func (e *Engine) completeRound(ctx context.Context, s meeting.State) (meeting.State, error) {
	summary := summarizeRound(s)
	s, err := e.append(ctx, s, eventRoundCompleted(s.CurrentRound, summary))
	if err != nil {
		return s, err
	}

	result, err := e.Strategy.Evaluate(consensus.Context{Meeting: s})
	if err != nil {
		return s, err
	}
	if result.Reached {
		return e.append(ctx, s, eventConsensusReached(s.ConsensusStrategy, result))
	}

	if s.CurrentRound >= s.MaxRoundsPerSegment {
		return e.append(ctx, s, eventModeratorDecision(s.ConsensusStrategy))
	}
	return e.startRound(ctx, s)
}

func spokenInRound(s meeting.State) map[string]bool {
	spoken := make(map[string]bool)
	for id := range s.RoundResponses[s.CurrentRound] {
		spoken[id] = true
	}
	return spoken
}

func summarizeRound(s meeting.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Round %d\n\n", s.CurrentRound)
	for _, id := range s.RoundOrder {
		r := s.RoundResponses[s.CurrentRound][id]
		role := s.Participants[id].Role
		fmt.Fprintf(&b, "- **%s** (%s): %s _[%s]_\n", id, role, r.Content, r.Stance)
	}
	return b.String()
}

func (e *Engine) buildPrompt(s meeting.State, participantID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\nRound: %d\nYou are %s (%s).\n", s.Topic, s.CurrentRound, participantID, s.Participants[participantID].Role)
	if s.PrincipalFeedback != "" {
		fmt.Fprintf(&b, "\nPrincipal feedback (address this round):\n%s\n", s.PrincipalFeedback)
	}
	if e.Workspace != nil {
		if data, err := e.Workspace.Read(s.ID, workspace.FileMeeting); err == nil {
			b.WriteString("\n--- MEETING.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (e *Engine) append(ctx context.Context, s meeting.State, env event.Envelope) (meeting.State, error) {
	env.MeetingID = s.ID
	if env.Sequence == 0 {
		seq, err := e.nextSequence(ctx, s.ID)
		if err != nil {
			return s, err
		}
		env.Sequence = seq
	}
	if env.ID == "" {
		env.ID = fmt.Sprintf("%s-%d", s.ID, env.Sequence)
	}
	if env.OccurredAt.IsZero() {
		env.OccurredAt = time.Now().UTC()
	}
	if env.Version == 0 {
		env.Version = 1
	}

	next, err := meeting.Apply(s, env)
	if err != nil {
		return s, err
	}
	if err := e.Store.Append(ctx, env); err != nil {
		return s, err
	}
	if err := e.project(ctx, next, env); err != nil {
		return s, err
	}
	return next, nil
}

func (e *Engine) nextSequence(ctx context.Context, meetingID string) (int, error) {
	events, err := e.Store.List(ctx, meetingID)
	if err != nil {
		return 0, err
	}
	return len(events) + 1, nil
}

func (e *Engine) project(ctx context.Context, s meeting.State, env event.Envelope) error {
	_ = ctx
	switch env.Type {
	case event.TypeMeetingCreated:
		if e.Workspace != nil {
			return e.Workspace.EnsureMeeting(s.ID, s.Topic)
		}
	case event.TypeParticipantInvited:
		if e.Profile != nil {
			p, _ := decodePayload[event.ParticipantInvitedPayload](env)
			if err := e.Profile.EnsureParticipant(p.ParticipantID); err != nil {
				return err
			}
		}
		if e.Knowledge != nil {
			p, _ := decodePayload[event.ParticipantInvitedPayload](env)
			_ = e.Knowledge.Ensure(knowledge.ScopeParticipant, p.ParticipantID)
		}
	case event.TypeRoundCompleted:
		p, _ := decodePayload[event.RoundCompletedPayload](env)
		if e.Workspace != nil {
			name := fmt.Sprintf("rounds/round-%03d.md", p.RoundNumber)
			body := fmt.Sprintf("# Round %d\n\n%s\n", p.RoundNumber, p.Summary)
			if err := e.Workspace.Write(s.ID, name, []byte(body)); err != nil {
				return err
			}
			if err := e.Workspace.Write(s.ID, workspace.FileMinutes, []byte(renderMinutes(s))); err != nil {
				return err
			}
		}
	case event.TypeConfirmationPrepared:
		if e.Workspace != nil {
			p, _ := decodePayload[event.ConfirmationPreparedPayload](env)
			body := renderConfirmationBrief(p.Brief, p.Cycle)
			if err := e.Workspace.Write(s.ID, "confirmation/brief.md", []byte(body)); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderConfirmationBrief(brief event.ConfirmationBrief, cycle int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Confirmation Brief (cycle %d)\n\n%s\n\n", cycle, brief.ExecutiveSummary)
	for _, item := range brief.Items {
		fmt.Fprintf(&b, "## %d. %s\n\n%s\n", item.Index, item.Title, item.Description)
		if item.Source != "" {
			fmt.Fprintf(&b, "_Source: %s_\n\n", item.Source)
		}
	}
	return b.String()
}

func (e *Engine) writeArtifactFile(meetingID, ref string, body []byte) error {
	if e.Workspace == nil {
		return nil
	}
	return e.Workspace.Write(meetingID, ref, body)
}

func renderMinutes(s meeting.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Minutes\n\n**Topic:** %s\n\n", s.Topic)
	for _, r := range s.Minutes.Rounds {
		fmt.Fprintf(&b, "## Round %d\n\n%s\n\n", r.RoundNumber, r.Summary)
	}
	if s.Consensus != nil {
		fmt.Fprintf(&b, "## Consensus\n\nStrategy: %s (resolved by %s)\n", s.Consensus.Strategy, s.Consensus.ResolvedBy)
	}
	return b.String()
}

func decodePayload[T any](env event.Envelope) (T, error) {
	var p T
	err := json.Unmarshal(env.Payload, &p)
	return p, err
}
