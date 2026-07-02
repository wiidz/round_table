package httptransport

import (
	"encoding/json"
	"net/http"

	"round_table/apps/server/internal/adapter/profile"
)

func (h *Handler) handlePutPrincipalUserProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	manifest, err := h.profile.EnsurePrincipalPersonas(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	active := manifest.ActivePersonaID
	if active == "" && len(manifest.Personas) > 0 {
		active = manifest.Personas[0].ID
	}
	if active == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no active persona"})
		return
	}

	var payload profile.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	userProfile, err := h.profile.WritePrincipalPersonaUserProfile(id, active, payload)
	if err != nil {
		status := http.StatusBadRequest
		if err == profile.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"user_profile": userProfile,
	})
}
