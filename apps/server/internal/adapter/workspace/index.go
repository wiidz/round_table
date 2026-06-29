package workspace

import "time"

// MeetingIndex is a lightweight listing entry derived from workspace files.
type MeetingIndex struct {
	ID                string    `json:"id"`
	Topic             string    `json:"topic"`
	Status            string    `json:"status"`
	Mode              string    `json:"mode,omitempty"`
	ModeKind          string    `json:"mode_kind,omitempty"` // decision | deliberation
	StartedAt         string    `json:"started_at,omitempty"`
	ParticipantCount  int       `json:"participant_count,omitempty"`
	MaxRounds         int       `json:"max_rounds,omitempty"`
	FreeDialogue      bool      `json:"free_dialogue,omitempty"`
	LLMCallCount      int       `json:"llm_call_count,omitempty"`
	TotalTokens       int       `json:"total_tokens,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// MeetingDetail includes markdown workspace files for one Meeting.
type MeetingDetail struct {
	MeetingIndex
	Files map[string]string `json:"files"`
}
