package httptransport

import (
	"encoding/json"
	"net/http"

	"round_table/apps/server/internal/platform/config"
)

func (h *Handler) handleGetSettings(w http.ResponseWriter, _ *http.Request) {
	if h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "config service unavailable",
		})
		return
	}
	writeJSON(w, http.StatusOK, h.config.SettingsView())
}

func (h *Handler) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "config service unavailable",
		})
		return
	}
	var body struct {
		Values map[string]string `json:"values"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}
	if body.Values == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "values required",
		})
		return
	}
	if err := h.config.Update(r.Context(), body.Values); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, h.config.SettingsView())
}

func (h *Handler) handlePutDiscordBots(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "config service unavailable",
		})
		return
	}
	var body config.DiscordBotsUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}
	if err := h.config.UpdateDiscordBots(r.Context(), body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, h.config.SettingsView())
}

func (h *Handler) handlePutMeetPresets(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "config service unavailable",
		})
		return
	}
	var body struct {
		Presets []config.MeetPresetConfig `json:"presets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}
	if len(body.Presets) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "presets required",
		})
		return
	}
	if err := h.config.UpdateMeetPresets(r.Context(), body.Presets); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, h.config.SettingsView())
}

func (h *Handler) handleResetMeetPresets(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "config service unavailable",
		})
		return
	}
	if err := h.config.ResetMeetPresets(r.Context()); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, h.config.SettingsView())
}

func (h *Handler) handlePutMeetCasts(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "config service unavailable",
		})
		return
	}
	var body struct {
		Casts []config.MeetCastConfig `json:"casts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}
	if body.Casts == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "casts required",
		})
		return
	}
	if err := h.config.UpdateMeetCasts(r.Context(), body.Casts); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, h.config.SettingsView())
}

func (h *Handler) handleResetMeetCasts(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "config service unavailable",
		})
		return
	}
	if err := h.config.ResetMeetCasts(r.Context()); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, h.config.SettingsView())
}

func (h *Handler) handleRefreshDiscordBotProfiles(w http.ResponseWriter, r *http.Request) {
	if h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "config service unavailable",
		})
		return
	}
	if err := h.config.RefreshDiscordBotProfiles(r.Context()); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, h.config.SettingsView())
}
