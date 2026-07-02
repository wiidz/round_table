package httptransport

import (
	"encoding/json"
	"net/http"

	"round_table/apps/server/internal/adapter/profile"
)

func (h *Handler) handlePostPrincipalPersona(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	persona, err := h.profile.CreatePrincipalPersona(id, body.Title)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"persona": persona})
}

func (h *Handler) handlePutPrincipalActivePersona(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		PersonaID string `json:"persona_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	manifest, err := h.profile.SetActivePrincipalPersona(id, body.PersonaID)
	if err != nil {
		status := http.StatusBadRequest
		if err == profile.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":            "ok",
		"active_persona_id": manifest.ActivePersonaID,
		"personas":          manifest.Personas,
	})
}

func (h *Handler) handlePutPrincipalPersonaUserProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	personaID := r.PathValue("personaId")
	var payload profile.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	userProfile, err := h.profile.WritePrincipalPersonaUserProfile(id, personaID, payload)
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

func (h *Handler) handleGetPrincipalPersona(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	personaID := r.PathValue("personaId")
	userProfile, err := h.profile.ReadPrincipalPersonaUserProfile(id, personaID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == profile.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":           id,
		"persona_id":   personaID,
		"user_profile": userProfile,
	})
}
