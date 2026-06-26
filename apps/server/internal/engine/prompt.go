package engine

import (
	"fmt"
	"strings"

	"round_table/apps/server/internal/domain/meeting"
)

// formatDiscussionContext builds shared context for debate rounds (1+).
func formatDiscussionContext(s meeting.State, participantID string) string {
	if s.CurrentRound <= 0 {
		return ""
	}

	var b strings.Builder

	if s.PreMeetingSummary != "" {
		b.WriteString("## Pre-meeting (Round 0)\n\n")
		b.WriteString(strings.TrimSpace(s.PreMeetingSummary))
		b.WriteString("\n\n")
	}

	for round := 1; round < s.CurrentRound; round++ {
		if round == 1 && s.FreeDialogueCompleted && s.FreeDialogueSummary != "" {
			b.WriteString("## Free dialogue after Round 1\n\n")
			b.WriteString(strings.TrimSpace(s.FreeDialogueSummary))
			b.WriteString("\n\n")
		}
		if sum, ok := s.ModeratorSummaries[round]; ok {
			fmt.Fprintf(&b, "## Moderator summary after Round %d\n\n", round)
			b.WriteString(strings.TrimSpace(sum))
			b.WriteString("\n\n")
		}
	}

	responses := s.RoundResponses[s.CurrentRound]
	if len(responses) > 0 {
		var current strings.Builder
		for _, id := range s.RoundOrder {
			if id == participantID {
				continue
			}
			r, ok := responses[id]
			if !ok {
				continue
			}
			role := s.Participants[id].Role
			fmt.Fprintf(&current, "- **%s** (%s): %s _[%s]_\n", id, role, r.Content, r.Stance)
		}
		if current.Len() > 0 {
			fmt.Fprintf(&b, "## Round %d (in progress)\n\n", s.CurrentRound)
			b.WriteString(current.String())
		}
	}

	return strings.TrimSpace(b.String())
}
