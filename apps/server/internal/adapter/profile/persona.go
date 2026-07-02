package profile

import "time"

// PrincipalPersonaMeta summarizes one Principal preference profile variant.
type PrincipalPersonaMeta struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PrincipalPersonaManifest tracks persona variants and the active selection.
type PrincipalPersonaManifest struct {
	ActivePersonaID string                 `json:"active_persona_id"`
	Personas        []PrincipalPersonaMeta `json:"personas"`
}
