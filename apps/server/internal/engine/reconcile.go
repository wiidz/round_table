package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/meeting"
)

const defaultReconcileAbortReason = "启动 reconcile：进程曾中断"

// ReconcileItem describes one meeting processed by ReconcileMeetings.
type ReconcileItem struct {
	MeetingID string `json:"meeting_id"`
	Action    string `json:"action"` // aborted | synced | skipped | error
	Detail    string `json:"detail,omitempty"`
}

// ReconcileResult aggregates ReconcileMeetings output.
type ReconcileResult struct {
	Scanned int             `json:"scanned"`
	Items   []ReconcileItem `json:"items"`
}

// ReconcileMeetings aborts orphan non-terminal meetings and syncs MEETING.md for terminal ones.
func ReconcileMeetings(ctx context.Context, e *Engine, reason string) (ReconcileResult, error) {
	if e == nil || e.Store == nil || e.Workspace == nil {
		return ReconcileResult{}, fmt.Errorf("engine: reconcile requires store and workspace")
	}
	if reason == "" {
		reason = defaultReconcileAbortReason
	}

	listStore, ok := e.Workspace.(workspaceListStore)
	if !ok {
		return ReconcileResult{}, fmt.Errorf("engine: workspace does not support listing meetings")
	}
	meetings, err := listStore.ListMeetings()
	if err != nil {
		return ReconcileResult{}, err
	}

	result := ReconcileResult{Scanned: len(meetings)}
	for _, idx := range meetings {
		item := ReconcileItem{MeetingID: idx.ID}
		action, detail, err := e.reconcileOne(ctx, idx.ID, reason)
		item.Action = action
		item.Detail = detail
		if err != nil {
			item.Action = "error"
			item.Detail = err.Error()
		}
		result.Items = append(result.Items, item)
		if item.Action == "aborted" || item.Action == "synced" {
			log.Printf("reconcile meeting %s: %s %s", idx.ID, item.Action, item.Detail)
		}
	}
	return result, nil
}

type workspaceListStore interface {
	ListMeetings() ([]workspace.MeetingIndex, error)
}

func (e *Engine) reconcileOne(ctx context.Context, meetingID, reason string) (action, detail string, err error) {
	events, err := e.Store.List(ctx, meetingID)
	if err != nil {
		return "", "", err
	}
	if len(events) == 0 {
		return e.reconcileDocOnly(meetingID)
	}

	s, err := e.LoadState(ctx, meetingID)
	if err != nil {
		return "", "", err
	}

	if s.Status == meeting.StatusCompleted || s.Status == meeting.StatusArchived {
		stale, docStatus := e.meetingDocStatusStale(meetingID)
		if stale {
			if err := e.writeMeetingDoc(s); err != nil {
				return "", "", err
			}
			return "synced", fmt.Sprintf("%s -> %s", docStatus, renderMeetingStatusLabel(s)), nil
		}
		return "skipped", string(s.Status), nil
	}

	if !IsAbortableStatus(s.Status) {
		return "skipped", string(s.Status), nil
	}

	final, err := e.AbortMeeting(ctx, meetingID, reason)
	if err != nil {
		return "", "", err
	}
	return "aborted", renderMeetingStatusLabel(final), nil
}

func (e *Engine) reconcileDocOnly(meetingID string) (action, detail string, err error) {
	docStatus, err := e.meetingDocStatus(meetingID)
	if err != nil {
		return "", "", err
	}
	if !IsStaleRunningMeetingDocStatus(docStatus) {
		return "skipped", docStatus, nil
	}
	if err := e.patchMeetingDocStatus(meetingID, RenderAbortedMeetingDocStatus()); err != nil {
		return "", "", err
	}
	return "aborted", fmt.Sprintf("%s -> %s", docStatus, RenderAbortedMeetingDocStatus()), nil
}

func (e *Engine) meetingDocStatus(meetingID string) (string, error) {
	data, err := e.Workspace.Read(meetingID, workspace.FileMeeting)
	if err != nil {
		return "", err
	}
	return ParseMeetingDocStatus(string(data)), nil
}

func (e *Engine) patchMeetingDocStatus(meetingID, status string) error {
	data, err := e.Workspace.Read(meetingID, workspace.FileMeeting)
	if err != nil {
		return err
	}
	doc := PatchMeetingDocStatus(string(data), status)
	return e.Workspace.Write(meetingID, workspace.FileMeeting, []byte(doc))
}

// SyncMeetingDoc rewrites MEETING.md from the current folded event state.
func (e *Engine) SyncMeetingDoc(ctx context.Context, meetingID string) error {
	events, err := e.Store.List(ctx, meetingID)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return errors.New("engine: meeting has no events")
	}
	s, err := e.LoadState(ctx, meetingID)
	if err != nil {
		return err
	}
	return e.writeMeetingDoc(s)
}

// RenderAbortedMeetingDocStatus is the MEETING.md status label for aborted meetings.
func RenderAbortedMeetingDocStatus() string {
	return "已中断"
}

// ParseMeetingDocStatus reads the 会议状态 table cell from MEETING.md.
func ParseMeetingDocStatus(doc string) string {
	for _, line := range strings.Split(doc, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "| 会议状态 |") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			return ""
		}
		return strings.TrimSpace(parts[2])
	}
	return ""
}

// PatchMeetingDocStatus replaces the 会议状态 row in MEETING.md.
func PatchMeetingDocStatus(doc, status string) string {
	lines := strings.Split(doc, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "| 会议状态 |") {
			continue
		}
		lines[i] = fmt.Sprintf("| 会议状态 | %s |", status)
		return strings.Join(lines, "\n")
	}
	return doc
}

// IsStaleRunningMeetingDocStatus reports MEETING.md statuses that imply an unfinished meeting.
func IsStaleRunningMeetingDocStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "进行中", "Running", "准备中", "Preparing", "已暂停", "Paused",
		"Principal 确认中", "Confirmation", "共识达成", "Consensus":
		return true
	default:
		return false
	}
}

func (e *Engine) meetingDocStatusStale(meetingID string) (bool, string) {
	docStatus, err := e.meetingDocStatus(meetingID)
	if err != nil {
		return false, ""
	}
	return IsStaleRunningMeetingDocStatus(docStatus), docStatus
}
