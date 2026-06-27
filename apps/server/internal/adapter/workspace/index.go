package workspace

import "time"

// MeetingIndex is a lightweight listing entry derived from workspace files.
type MeetingIndex struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
