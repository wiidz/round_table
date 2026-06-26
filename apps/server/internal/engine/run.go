package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/knowledge"
	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/scheduler"
	"round_table/apps/server/internal/stream"
)

func (e *Engine) startRound(ctx context.Context, s meeting.State) (meeting.State, error) {
	order := append([]string(nil), s.ParticipantOrder...)
	var roundNum int
	if !s.PreMeetingCompleted {
		roundNum = 0
	} else {
		roundNum = s.CurrentRound + 1
	}
	return e.append(ctx, s, eventRoundStarted(roundNum, order))
}

func (e *Engine) advanceRunning(ctx context.Context, s meeting.State) (meeting.State, error) {
	if s.CurrentRound > 0 {
		if ns, handled, err := e.maybePrincipalRunningAction(ctx, s); err != nil || handled {
			return ns, err
		}
	}
	spoken := spokenInRound(s)
	if next, ok := scheduler.FixedOrder(s.RoundOrder, spoken); ok {
		return e.inviteSpeak(ctx, s, next)
	}
	return e.completeRound(ctx, s)
}

func (e *Engine) inviteSpeak(ctx context.Context, s meeting.State, participantID string) (meeting.State, error) {
	var prompt string
	if s.CurrentRound == 0 {
		prompt = e.buildPreMeetingPrompt(s, participantID)
	} else if s.IsDeliberation() {
		prompt = e.buildDeliberationPrompt(s, participantID)
	} else {
		prompt = e.buildPrompt(s, participantID)
	}
	phase := PhaseDebate
	if s.CurrentRound == 0 {
		phase = PhasePreMeeting
	} else if s.IsDeliberation() {
		phase = PhaseDeliberation
	}
	phaseLabel := strings.TrimPrefix(phase, "Phase: ")
	e.logLLMWaiting(phaseLabel, participantID, "turn="+debateTurnLabel(s, participantID)+" round="+strconv.Itoa(s.CurrentRound))
	ctx = e.withStreamCtx(ctx, stream.Meta{
		ParticipantID: participantID,
		Phase:         phaseLabel,
		Detail:        "turn " + debateTurnLabel(s, participantID) + " · round " + strconv.Itoa(s.CurrentRound),
	})
	start := time.Now()
	resp, err := e.Participant.Respond(ctx, s.ID, participantID, prompt)
	elapsed := time.Since(start)
	if err != nil {
		return s, err
	}
	stance := event.Stance(resp.Stance)
	if s.CurrentRound == 0 || s.IsDeliberation() {
		stance = event.StanceNone
	} else if stance == "" {
		stance = event.StanceAgree
	}
	e.logLLMDone(phaseLabel, participantID, string(stance), resp, elapsed)
	return e.append(ctx, s, eventParticipantResponded(
		participantID, s.CurrentRound, resp.Content, stance, resp.ObjectReason,
		tokenUsageFromResponse(phase, participantID, s.CurrentRound, 0, resp),
	))
}

func (e *Engine) completeRound(ctx context.Context, s meeting.State) (meeting.State, error) {
	var summary string
	if s.CurrentRound == 0 {
		summary = summarizePreMeeting(s)
	} else {
		summary = summarizeRound(s)
	}
	s, err := e.append(ctx, s, eventRoundCompleted(s.CurrentRound, summary))
	if err != nil {
		return s, err
	}

	if s.CurrentRound == 0 {
		return e.startRound(ctx, s)
	}

	if s.CurrentRound == 1 && !s.FreeDialogueCompleted && s.FreeDialogueMaxQuestions > 0 && len(s.ParticipantOrder) >= 2 {
		return e.startFreeDialogue(ctx, s)
	}

	return e.continueAfterDebateRound(ctx, s)
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

func (e *Engine) buildPreMeetingPrompt(s meeting.State, participantID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n%s\nPre-meeting (Round 0)\nYou are %s (%s).\n", s.Topic, PhasePreMeeting, participantID, s.Participants[participantID].Role)
	b.WriteString("\nThis is a private pre-meeting turn. Other participants cannot see your response yet.\n")
	b.WriteString("Share 2–4 initial perspectives or angles you will use to evaluate this topic.\n")
	b.WriteString("Do not react to others — they have not spoken yet.\n")
	if e.Workspace != nil {
		if data, err := e.Workspace.Read(s.ID, workspace.FileMeeting); err == nil {
			b.WriteString("\n--- MEETING.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (e *Engine) buildPrompt(s meeting.State, participantID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Topic: %s\n%s\nRound: %d\nYou are %s (%s).\n", s.Topic, PhaseDebate, s.CurrentRound, participantID, s.Participants[participantID].Role)
	if s.PrincipalFeedback != "" {
		fmt.Fprintf(&b, "\nPrincipal feedback (address this round):\n%s\n", s.PrincipalFeedback)
	}
	if ctx := formatDiscussionContext(s, participantID); ctx != "" {
		b.WriteString("\n--- Discussion so far ---\n")
		b.WriteString(ctx)
		b.WriteByte('\n')
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
	e.logProgress(env, next)
	return next, nil
}

func (e *Engine) nextSequence(ctx context.Context, meetingID string) (int, error) {
	events, err := e.Store.List(ctx, meetingID)
	if err != nil {
		return 0, err
	}
	return len(events) + 1, nil
}

func (e *Engine) writeMeetingDoc(s meeting.State) error {
	if e.Workspace == nil {
		return nil
	}
	return e.Workspace.Write(s.ID, workspace.FileMeeting, []byte(renderMeetingDoc(s)))
}

func (e *Engine) project(ctx context.Context, s meeting.State, env event.Envelope) error {
	_ = ctx
	switch env.Type {
	case event.TypeMeetingCreated:
		if e.Workspace != nil {
			if err := e.Workspace.EnsureMeeting(s.ID, s.Topic); err != nil {
				return err
			}
			return e.writeMeetingDoc(s)
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
		return e.writeMeetingDoc(s)
	case event.TypeParticipantResponded, event.TypeFreeDialogueQuestion, event.TypeFreeDialogueAnswer, event.TypeMeetingFinished:
		if err := e.projectTokenUsage(s); err != nil {
			return err
		}
		if env.Type == event.TypeMeetingFinished {
			return e.writeMeetingDoc(s)
		}
	case event.TypeRoundStarted:
		return e.writeMeetingDoc(s)
	case event.TypeRoundCompleted:
		p, _ := decodePayload[event.RoundCompletedPayload](env)
		if e.Workspace != nil {
			name := fmt.Sprintf("rounds/round-%03d.md", p.RoundNumber)
			title := fmt.Sprintf("# Round %d\n\n", p.RoundNumber)
			if p.RoundNumber == 0 {
				name = "pre-meeting/perspectives.md"
				title = "# Pre-meeting (Round 0)\n\n"
			}
			body := title + p.Summary + "\n"
			if err := e.Workspace.Write(s.ID, name, []byte(body)); err != nil {
				return err
			}
			if err := e.Workspace.Write(s.ID, workspace.FileMinutes, []byte(renderMinutes(s))); err != nil {
				return err
			}
		}
	case event.TypeModeratorSummarized:
		p, _ := decodePayload[event.ModeratorSummarizedPayload](env)
		if e.Workspace != nil {
			name := fmt.Sprintf("moderator/round-%03d-summary.md", p.RoundNumber)
			body := fmt.Sprintf("# Moderator Summary — Round %d\n\n%s\n", p.RoundNumber, p.Summary)
			if err := e.Workspace.Write(s.ID, name, []byte(body)); err != nil {
				return err
			}
		}
	case event.TypeDeliberationReadinessChecked:
		if err := e.projectTokenUsage(s); err != nil {
			return err
		}
		p, _ := decodePayload[event.DeliberationReadinessCheckedPayload](env)
		if e.Workspace != nil {
			var body strings.Builder
			fmt.Fprintf(&body, "# Synthesis Readiness — Round %d\n\n", p.RoundNumber)
			fmt.Fprintf(&body, "Ready: **%v**\n\n", p.Ready)
			if p.Rationale != "" {
				fmt.Fprintf(&body, "## Rationale\n\n%s\n\n", p.Rationale)
			}
			if len(p.Gaps) > 0 {
				body.WriteString("## Gaps\n\n")
				for _, g := range p.Gaps {
					fmt.Fprintf(&body, "- %s\n", g)
				}
				body.WriteByte('\n')
			}
			name := fmt.Sprintf("moderator/round-%03d-readiness.md", p.RoundNumber)
			if err := e.Workspace.Write(s.ID, name, []byte(body.String())); err != nil {
				return err
			}
		}
	case event.TypeFreeDialogueCompleted:
		p, _ := decodePayload[event.FreeDialogueCompletedPayload](env)
		if e.Workspace != nil {
			body := fmt.Sprintf("# Free Dialogue — after Round %d\n\n%s\n", p.AfterRound, p.Summary)
			if err := e.Workspace.Write(s.ID, "free-dialogue/after-round-001.md", []byte(body)); err != nil {
				return err
			}
			if err := e.Workspace.Write(s.ID, workspace.FileMinutes, []byte(renderMinutes(s))); err != nil {
				return err
			}
		}
	case event.TypeSynthesisCompleted:
		if err := e.projectTokenUsage(s); err != nil {
			return err
		}
		if e.Workspace != nil && s.SynthesisSummary != "" {
			if err := e.Workspace.Write(s.ID, "artifacts/design-draft.md", []byte(s.SynthesisSummary+"\n")); err != nil {
				return err
			}
			if len(s.SynthesisOpenQuestions) > 0 {
				var ob strings.Builder
				ob.WriteString("# Open Questions\n\n")
				for _, q := range s.SynthesisOpenQuestions {
					fmt.Fprintf(&ob, "- %s\n", q)
				}
				if err := e.Workspace.Write(s.ID, "artifacts/open-questions.md", []byte(ob.String())); err != nil {
					return err
				}
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

func (e *Engine) projectTokenUsage(s meeting.State) error {
	if e.Workspace == nil {
		return nil
	}
	return writeTokenUsageFiles(s, func(name string, body []byte) error {
		return e.Workspace.Write(s.ID, name, body)
	})
}

func renderMinutes(s meeting.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Minutes\n\n**Topic:** %s\n\n", s.Topic)
	for _, r := range s.Minutes.Rounds {
		if r.RoundNumber == 0 {
			fmt.Fprintf(&b, "## Pre-meeting (Round 0)\n\n%s\n\n", r.Summary)
			continue
		}
		fmt.Fprintf(&b, "## Round %d\n\n%s\n\n", r.RoundNumber, r.Summary)
		if r.RoundNumber == 1 && s.FreeDialogueCompleted && s.FreeDialogueSummary != "" {
			fmt.Fprintf(&b, "### Free dialogue\n\n%s\n\n", s.FreeDialogueSummary)
		}
		if sum, ok := s.ModeratorSummaries[r.RoundNumber]; ok {
			fmt.Fprintf(&b, "### Moderator summary\n\n%s\n\n", sum)
		}
	}
	if s.Consensus != nil {
		if s.IsDeliberation() {
			fmt.Fprintf(&b, "## Synthesis\n\nMode: deliberation (resolved by %s)\n\n", s.Consensus.ResolvedBy)
			if s.SynthesisSummary != "" {
				b.WriteString(s.SynthesisSummary)
				b.WriteString("\n\n")
			}
		} else {
			fmt.Fprintf(&b, "## Consensus\n\nStrategy: %s (resolved by %s)\n", s.Consensus.Strategy, s.Consensus.ResolvedBy)
		}
	}
	if s.TokenUsageTotals.CallCount > 0 {
		fmt.Fprintf(&b, "\n## Token usage\n\nTotal tokens: **%d** (%d LLM calls — see `usage/summary.md`)\n",
			s.TokenUsageTotals.TotalTokens, s.TokenUsageTotals.CallCount)
	}
	return b.String()
}

func decodePayload[T any](env event.Envelope) (T, error) {
	var p T
	err := json.Unmarshal(env.Payload, &p)
	return p, err
}
