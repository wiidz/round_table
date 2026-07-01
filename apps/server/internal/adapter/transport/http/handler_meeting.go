package httptransport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

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

func (h *Handler) newMaintenanceEngine() *engine.Engine {
	return engine.New(h.events, consensus.NoObjection{}, nil, nil, h.workspace, nil, nil)
}
