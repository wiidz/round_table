package profile

import (
	"time"

	"round_table/apps/server/internal/platform/config"
)

// PrincipalIndex summarizes one Principal profile directory.
type PrincipalIndex struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name,omitempty"`
	Files       []string  `json:"files"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PrincipalDetail includes active persona preferences and all persona metadata.
type PrincipalDetail struct {
	ID              string                 `json:"id"`
	DisplayName     string                 `json:"display_name,omitempty"`
	ActivePersonaID string                 `json:"active_persona_id"`
	Personas        []PrincipalPersonaMeta `json:"personas"`
	UserProfile     UserProfile            `json:"user_profile"`
	Files           map[string]string      `json:"files,omitempty"`
}

// ParticipantIndex summarizes one Participant profile directory.
type ParticipantIndex struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name,omitempty"`
	Expertise   string    `json:"expertise,omitempty"`
	IMBindings  []config.ParticipantIMBind `json:"im_bindings,omitempty"`
	InRoster    bool                       `json:"in_roster,omitempty"`
	Files       []string  `json:"files"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ParticipantDetail includes markdown file contents for one Participant.
type ParticipantDetail struct {
	ID          string            `json:"id"`
	DisplayName string            `json:"display_name,omitempty"`
	Expertise   string            `json:"expertise,omitempty"`
	IMBindings  []config.ParticipantIMBind `json:"im_bindings,omitempty"`
	Files       map[string]string `json:"files"`
}
