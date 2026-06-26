package workspace

import "errors"

var (
	// ErrNotFound indicates the requested path does not exist.
	ErrNotFound = errors.New("workspace: not found")
	// ErrOutsideRoot indicates a path escapes the meeting workspace.
	ErrOutsideRoot = errors.New("workspace: path outside meeting root")
)

// Port is the file workspace for a Meeting's Markdown outputs (ADR-0009).
// Implementations must jail relative paths to the meeting directory.
type Port interface {
	// EnsureMeeting creates {root}/{meetingID}/ and seeds MEETING.md if missing.
	EnsureMeeting(meetingID, topic string) error
	Read(meetingID, relPath string) ([]byte, error)
	Write(meetingID, relPath string, data []byte) error
	List(meetingID, relPath string) ([]string, error)
	Resolve(meetingID, relPath string) (string, error)
}

const (
	FileMeeting    = "MEETING.md"
	FileMinutes    = "MINUTES.md"
	FileActionItems = "action-items.md"
	DirArtifacts   = "artifacts"
	DirRounds      = "rounds"
)
