package web

// Frame is a WebSocket JSON envelope between browser and server.
type Frame struct {
	Type       string `json:"type"`
	ID         string `json:"id,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
	Role       string `json:"role,omitempty"`
	AuthorID   string `json:"author_id,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
	At         string `json:"at,omitempty"`
	Content    string `json:"content,omitempty"`
	Error      string `json:"error,omitempty"`
}

const (
	FrameConnected = "connected"
	FrameMessage   = "message"
	FrameError     = "error"
	FrameTyping    = "typing"

	RoleUser        = "user"
	RoleModerator   = "moderator"
	RoleParticipant = "participant"
	RoleSystem      = "system"
)
