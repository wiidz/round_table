package meeting

import (
	"time"

	"round_table/apps/server/internal/domain/event"
)

// Status is the Meeting lifecycle state.
type Status string

const (
	StatusCreated      Status = "Created"
	StatusPreparing    Status = "Preparing"
	StatusRunning      Status = "Running"
	StatusPaused       Status = "Paused"
	StatusConsensus    Status = "Consensus"
	StatusConfirmation Status = "Confirmation"
	StatusCompleted    Status = "Completed"
	StatusArchived     Status = "Archived"
)

// State is the folded Meeting projection (ADR-0003).
type State struct {
	ID                    string
	Status                Status
	Topic                 string
	Agenda                []event.AgendaItem
	ConsensusStrategy     string
	MaxRoundsPerSegment   int
	ConfirmationMode      string
	MaxConfirmationCycles int

	StartedAt             time.Time
	Goal                  string
	Participants     map[string]ParticipantState
	ParticipantOrder []string

	CurrentRound     int
	RoundOrder       []string
	RoundResponses   map[int]map[string]RoundResponse
	PreMeetingCompleted bool
	PreMeetingSummary   string
	ModeratorSummaries  map[int]string

	FreeDialogueMaxQuestions int
	FreeDialogueCompleted    bool
	FreeDialogueSummary      string
	FreeDialogueExchanges    []FreeDialogueExchange
	InFreeDialogue           bool
	FreeDialogueQuestionIndex int
	FreeDialogueAskerIndex    int
	PendingFreeDialogue      *PendingFreeDialogue

	TokenUsageLog    []TokenUsageRecord
	TokenUsageTotals TokenUsageTotals

	RunningSegment   int
	ConfirmationCycle int
	PrincipalFeedback string

	Consensus     *ConsensusState
	Confirmation  *ConfirmationState
	PausedFrom    Status

	Minutes     MinutesDraft
	Artifacts   []ArtifactRef
	ActionItems []ActionItem
	Outcome     string
}

// NewState returns initial Meeting state before any events.
func NewState(id string) State {
	return State{
		ID:             id,
		Status:         StatusCreated,
		Participants:   make(map[string]ParticipantState),
		RoundResponses: make(map[int]map[string]RoundResponse),
	}
}

func (s State) isTerminal() bool {
	return s.Status == StatusCompleted || s.Status == StatusArchived
}

// DebateRoundCount returns completed debate rounds (1+), excluding pre-meeting round 0.
func (s State) DebateRoundCount() int {
	n := 0
	for _, r := range s.Minutes.Rounds {
		if r.RoundNumber > 0 {
			n++
		}
	}
	return n
}
