package profile

import "errors"

var ErrNotFound = errors.New("profile: not found")

// Port manages Participant / Principal / Moderator identity files (ADR-0010).
type Port interface {
	EnsureParticipant(id string) error
	EnsurePrincipal(id string) error
	EnsureModerator() error
	ReadParticipant(id, filename string) ([]byte, error)
	WriteParticipant(id, filename string, data []byte) error
	ReadPrincipal(id, filename string) ([]byte, error)
	WritePrincipal(id, filename string, data []byte) error
	ReadModerator(filename string) ([]byte, error)
}

const (
	FileSoul   = "SOUL.md"
	FileAgents = "AGENTS.md"
	FileTools  = "TOOLS.md"
	FileUser   = "USER.md"
)
