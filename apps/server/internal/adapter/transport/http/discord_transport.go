package httptransport

import "net/http"

func (h *Handler) handleGetDiscordTransportStatus(w http.ResponseWriter, _ *http.Request) {
	if h.discordSvc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "discord transport supervisor unavailable",
		})
		return
	}
	writeJSON(w, http.StatusOK, h.discordSvc.Status())
}

func (h *Handler) handlePostDiscordTransportStart(w http.ResponseWriter, r *http.Request) {
	if h.discordSvc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "discord transport supervisor unavailable",
		})
		return
	}
	if err := h.discordSvc.Start(r.Context(), h.config.Current()); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, h.discordSvc.Status())
}

func (h *Handler) handlePostDiscordTransportStop(w http.ResponseWriter, _ *http.Request) {
	if h.discordSvc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "discord transport supervisor unavailable",
		})
		return
	}
	if err := h.discordSvc.Stop(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, h.discordSvc.Status())
}

func (h *Handler) handleGetDiscordTransportLogs(w http.ResponseWriter, r *http.Request) {
	if h.discordSvc == nil || h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "discord transport supervisor unavailable",
		})
		return
	}
	maxLines := queryInt(r, "lines", 200)
	logs, err := h.discordSvc.Logs(h.config.Current(), maxLines)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) handlePostDiscordTransportLogsClear(w http.ResponseWriter, r *http.Request) {
	if h.discordSvc == nil || h.config == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "discord transport supervisor unavailable",
		})
		return
	}
	if err := h.discordSvc.ClearLogs(h.config.Current()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	logs, err := h.discordSvc.Logs(h.config.Current(), 200)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, logs)
}
