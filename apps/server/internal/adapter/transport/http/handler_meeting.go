package httptransport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"round_table/apps/server/internal/adapter/storage"
	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/consensus"
	"round_table/apps/server/internal/engine"
)

func (h *Handler) handlePostMeetingAbort(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "meeting id required"})
		return
	}
	if h.events == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "event store unavailable"})
		return
	}

	reason := "手动中止"
	if r.Body != nil {
		body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
			return
		}
		if len(body) > 0 {
			var payload struct {
				Reason string `json:"reason"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
				return
			}
			if trimmed := payload.Reason; trimmed != "" {
				reason = trimmed
			}
		}
	}

	events, err := h.events.List(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if len(events) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "meeting not found"})
		return
	}

	eng := h.newMaintenanceEngine()
	final, err := eng.AbortMeeting(r.Context(), id, reason)
	if errors.Is(err, engine.ErrMeetingAlreadyFinished) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	if errors.Is(err, engine.ErrMeetingNotAbortable) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "ok",
		"meeting_id":     id,
		"meeting_status": string(final.Status),
		"outcome":        final.Outcome,
	})
}

func (h *Handler) handleGetMeetingArchive(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "meeting id required"})
		return
	}
	if h.workspace == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "workspace unavailable"})
		return
	}

	var buf bytes.Buffer
	if err := h.workspace.WriteMeetingArchive(id, &buf); err != nil {
		if errors.Is(err, workspace.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "meeting not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, sanitizeDownloadFilename(id)))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, &buf)
}

func (h *Handler) handleDeleteMeeting(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "meeting id required"})
		return
	}

	exists, err := h.meetingExists(r, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if !exists {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "meeting not found"})
		return
	}

	if deleter, ok := h.events.(storage.MeetingDeleter); ok && h.events != nil {
		if err := deleter.DeleteMeeting(r.Context(), id); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	if h.workspace != nil {
		if err := h.workspace.DeleteMeeting(id); err != nil && !errors.Is(err, workspace.ErrNotFound) {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "ok",
		"meeting_id": id,
	})
}

func (h *Handler) meetingExists(r *http.Request, id string) (bool, error) {
	if h.workspace != nil {
		if _, err := h.workspace.ReadMeetingDetail(id); err == nil {
			return true, nil
		} else if !errors.Is(err, workspace.ErrNotFound) {
			return false, err
		}
	}
	if h.events != nil {
		envs, err := h.events.List(r.Context(), id)
		if err != nil {
			return false, err
		}
		if len(envs) > 0 {
			return true, nil
		}
	}
	return false, nil
}

func sanitizeDownloadFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "meeting"
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "meeting"
	}
	return out
}

func (h *Handler) newMaintenanceEngine() *engine.Engine {
	return engine.New(h.events, consensus.NoObjection{}, nil, nil, h.workspace, nil, nil)
}
