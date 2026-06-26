package knowledge

import "errors"

var (
	ErrNotFound     = errors.New("knowledge: not found")
	ErrInvalidScope = errors.New("knowledge: invalid scope")
	ErrInvalidOwner = errors.New("knowledge: invalid owner id")
)

// Scope is the knowledge isolation boundary (ADR-0006).
type Scope string

const (
	ScopeParticipant Scope = "participants"
	ScopePrincipal   Scope = "principals"
	ScopeShared      Scope = "shared"
)

// Port manages long-term Knowledge files (MEMORY.md + daily logs).
type Port interface {
	Ensure(scope Scope, ownerID string) error
	ReadMemory(scope Scope, ownerID string) ([]byte, error)
	WriteMemory(scope Scope, ownerID string, data []byte) error
	AppendDailyLog(scope Scope, ownerID, date string, data []byte) error
	ReadDailyLog(scope Scope, ownerID, date string) ([]byte, error)
	ListDailyLogs(scope Scope, ownerID string) ([]string, error)
}

const (
	FileMemory    = "MEMORY.md"
	DirMemoryLogs = "memory"
)
