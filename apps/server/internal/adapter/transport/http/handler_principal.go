package httptransport

import (
	"encoding/json"
	"net/http"

	"round_table/apps/server/internal/adapter/profile"
)

func (h *Handler) handlePutPrincipalUserProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.profile.EnsurePrincipal(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var payload profile.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	content := profile.RenderUserMD(payload)
	if err := h.profile.WritePrincipalFile(id, profile.FileUser, []byte(content)); err != nil {
		status := http.StatusBadRequest
		if err == profile.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"user_profile": profile.ParseUserMD(content),
	})
}
