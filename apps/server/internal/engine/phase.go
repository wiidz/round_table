package engine

import "strings"

// Prompt phase markers — debate prompts may reference pre-meeting context; only this header selects pre-meeting mode.
const (
	PhasePreMeeting         = "Phase: pre-meeting"
	PhaseDebate             = "Phase: debate"
	PhaseDeliberation       = "Phase: deliberation"
	PhaseFreeDialogueAsk        = "Phase: free-dialogue-ask"
	PhaseFreeDialogueAnswer     = "Phase: free-dialogue-answer"
	PhaseDeliberationSynthesis  = "Phase: deliberation-synthesis"
	PhaseDeliberationReadiness  = "Phase: deliberation-readiness"
)

func isPreMeetingPhase(prompt string) bool {
	for _, line := range strings.Split(prompt, "\n") {
		if strings.TrimSpace(line) == PhasePreMeeting {
			return true
		}
	}
	return false
}
