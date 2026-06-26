package engine

import (
	"log"
	"strconv"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/participant"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

// ProgressLogger receives meeting progress lines (for CLI / debug).
type ProgressLogger interface {
	Logf(format string, args ...any)
}

// DiscardProgressLogger suppresses progress output (tests).
type DiscardProgressLogger struct{}

func (DiscardProgressLogger) Logf(string, ...any) {}

// StdProgressLogger writes progress to the standard logger with a meet: prefix.
type StdProgressLogger struct{}

func (StdProgressLogger) Logf(format string, args ...any) {
	log.Printf("meet: "+format, args...)
}

func (e *Engine) logf(format string, args ...any) {
	if e.Progress == nil {
		return
	}
	e.Progress.Logf(format, args...)
}

func (e *Engine) logProgress(env event.Envelope, s meeting.State) {
	switch env.Type {
	case event.TypeRoundStarted:
		p, _ := decodePayload[event.RoundStartedPayload](env)
		if p.RoundNumber == 0 {
			e.logf("▶ pre-meeting started order=%s", strings.Join(p.Order, " → "))
		} else {
			e.logf("▶ debate round %d started order=%s", p.RoundNumber, strings.Join(p.Order, " → "))
		}
	case event.TypeRoundCompleted:
		p, _ := decodePayload[event.RoundCompletedPayload](env)
		if p.RoundNumber == 0 {
			e.logf("■ pre-meeting completed → starting debate round 1")
		} else {
			e.logf("■ debate round %d completed", p.RoundNumber)
		}
	case event.TypeFreeDialogueStarted:
		p, _ := decodePayload[event.FreeDialogueStartedPayload](env)
		total := p.MaxQuestions * len(s.ParticipantOrder)
		e.logf("▶ free dialogue after round %d (%d Q&A pairs, max_questions=%d/person)",
			p.AfterRound, total, p.MaxQuestions)
	case event.TypeFreeDialogueQuestion:
		p, _ := decodePayload[event.FreeDialogueQuestionAskedPayload](env)
		total := freeDialogueTotal(s)
		e.logf("  ? question %d/%d %s → %s",
			p.QuestionIndex+1, total, p.AskerID, p.AnswererID)
	case event.TypeFreeDialogueAnswer:
		p, _ := decodePayload[event.FreeDialogueAnsweredPayload](env)
		total := freeDialogueTotal(s)
		e.logf("  ✓ answer %d/%d %s answered %s",
			p.QuestionIndex+1, total, p.AnswererID, p.AskerID)
	case event.TypeFreeDialogueCompleted:
		e.logf("■ free dialogue completed (%d exchanges)", len(s.FreeDialogueExchanges))
	case event.TypeModeratorSummarized:
		p, _ := decodePayload[event.ModeratorSummarizedPayload](env)
		e.logf("◆ moderator summary written for round %d", p.RoundNumber)
	case event.TypeDeliberationReadinessChecked:
		p, _ := decodePayload[event.DeliberationReadinessCheckedPayload](env)
		if p.Ready {
			e.logf("◆ synthesis readiness round=%d ready=true (%s)", p.RoundNumber, p.Rationale)
		} else {
			e.logf("◆ synthesis readiness round=%d ready=false (%s)", p.RoundNumber, p.Rationale)
		}
	case event.TypeConsensusReached:
		p, _ := decodePayload[event.ConsensusReachedPayload](env)
		e.logf("★ consensus reached strategy=%s resolved_by=%s", p.Strategy, p.ResolvedBy)
	case event.TypeSynthesisCompleted:
		p, _ := decodePayload[event.SynthesisCompletedPayload](env)
		e.logf("★ synthesis completed resolved_by=%s open_questions=%d", p.ResolvedBy, len(p.OpenQuestions))
	case event.TypeSynthesisForced:
		p, _ := decodePayload[event.SynthesisForcedPayload](env)
		e.logf("◇ synthesis forced by principal (%s)", p.Reason)
	case event.TypeConsensusForced:
		p, _ := decodePayload[event.ConsensusForcedPayload](env)
		e.logf("◇ consensus forced by principal (%s)", p.Reason)
	case event.TypeConfirmationPrepared:
		p, _ := decodePayload[event.ConfirmationPreparedPayload](env)
		e.logf("▶ confirmation prepared cycle=%d", p.Cycle)
	case event.TypeConfirmationPresented:
		p, _ := decodePayload[event.ConfirmationPresentedPayload](env)
		e.logf("… waiting for principal confirmation cycle=%d", p.Cycle)
	case event.TypeConfirmationApproved:
		p, _ := decodePayload[event.ConfirmationApprovedPayload](env)
		e.logf("★ confirmation approved cycle=%d", p.Cycle)
	case event.TypeConfirmationRejected:
		p, _ := decodePayload[event.ConfirmationRejectedPayload](env)
		e.logf("↩ confirmation rejected cycle=%d — resuming debate", p.Cycle)
	case event.TypeConfirmationForced:
		p, _ := decodePayload[event.ConfirmationForcedPayload](env)
		e.logf("★ confirmation forced cycle=%d (%s)", p.Cycle, p.Reason)
	case event.TypeMeetingFinished:
		p, _ := decodePayload[event.MeetingFinishedPayload](env)
		e.logf("■ meeting finished outcome=%s", p.Outcome)
	}
}

func (e *Engine) logLLMWaiting(phase, participantID string, detail string) {
	e.logf("… LLM %s participant=%s %s", phase, participantID, detail)
}

func (e *Engine) logLLMDone(phase, participantID, stance string, resp participant.Response, elapsed time.Duration) {
	tokens := resp.Usage.TotalTokens
	if tokens == 0 {
		e.logf("✓ LLM %s participant=%s stance=%s elapsed=%s",
			phase, participantID, stance, elapsed.Round(time.Millisecond))
		return
	}
	e.logf("✓ LLM %s participant=%s stance=%s tokens=%d elapsed=%s",
		phase, participantID, stance, tokens, elapsed.Round(time.Millisecond))
}

func debateTurnLabel(s meeting.State, participantID string) string {
	spoken := len(s.RoundResponses[s.CurrentRound])
	total := len(s.RoundOrder)
	return formatTurn(spoken+1, total)
}

func formatTurn(current, total int) string {
	if total <= 0 {
		return ""
	}
	return "(" + strconv.Itoa(current) + "/" + strconv.Itoa(total) + ")"
}

func freeDialogueTurnLabel(s meeting.State) string {
	total := freeDialogueTotal(s)
	idx := s.FreeDialogueQuestionIndex + 1
	if s.PendingFreeDialogue != nil {
		idx = s.PendingFreeDialogue.QuestionIndex + 1
	}
	return formatTurn(idx, total)
}
