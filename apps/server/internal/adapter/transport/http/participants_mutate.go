package httptransport

import (
	"encoding/json"
	"net/http"
	"strings"

	"round_table/apps/server/internal/platform/config"
)

func (h *Handler) handleCreateParticipant(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeServiceUnavailable(w)
		return
	}
	var body config.ParticipantRosterItem
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := h.config.CreateParticipant(r.Context(), body); err != nil {
		writeParticipantError(w, err)
		return
	}
	if err := h.profile.EnsureParticipant(strings.TrimSpace(body.ID)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.writeParticipantList(w)
}

func (h *Handler) handleUpdateParticipant(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeServiceUnavailable(w)
		return
	}
	oldID := strings.TrimSpace(r.PathValue("id"))
	var body config.ParticipantRosterItem
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	newID := strings.TrimSpace(body.ID)
	if newID == "" {
		newID = oldID
	}
	body.ID = newID

	if err := h.config.UpdateParticipant(r.Context(), oldID, body); err != nil {
		writeParticipantError(w, err)
		return
	}
	if newID != oldID {
		if err := h.profile.RenameParticipant(oldID, newID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	h.writeParticipantList(w)
}

func (h *Handler) handleDeleteParticipant(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeServiceUnavailable(w)
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if err := h.config.DeleteParticipant(r.Context(), id); err != nil {
		writeParticipantError(w, err)
		return
	}
	if err := h.profile.DeleteParticipant(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.writeParticipantList(w)
}

func (h *Handler) writeParticipantList(w http.ResponseWriter) {
	h.ensureConfiguredParticipants()
	list, err := h.listParticipantsForAPI()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	for i := range list {
		list[i] = h.enrichParticipantIndex(list[i])
	}
	writeJSON(w, http.StatusOK, map[string]any{"participants": list})
}

func writeServiceUnavailable(w http.ResponseWriter) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{
		"error": "config service unavailable",
	})
}

func writeParticipantError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
}
