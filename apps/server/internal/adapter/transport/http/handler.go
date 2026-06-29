package httptransport

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"round_table/apps/server/internal/adapter/profile"
	profFS "round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/storage"
	"round_table/apps/server/internal/adapter/workspace"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	wsfs "round_table/apps/server/internal/adapter/workspace/fs"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/engine"
	"round_table/apps/server/internal/platform/config"
	"round_table/apps/server/internal/platform/discordsvc"
)

// Handler serves REST endpoints for web clients.
type Handler struct {
	workspace  *wsfs.Store
	profile    *profFS.Store
	bindings   *principalbind.Registry
	meetings   storage.MeetingCatalog
	events     storage.Store
	config     *config.Service
	discordSvc *discordsvc.Supervisor
}

func NewHandler(cfg config.Config, catalog storage.MeetingCatalog, events storage.Store, configSvc *config.Service, discordSvc *discordsvc.Supervisor) (*Handler, error) {
	reg, err := principalbind.NewRegistry(cfg.Transport.Discord.BindingsFile)
	if err != nil {
		return nil, err
	}
	return &Handler{
		workspace:  wsfs.NewStore(cfg.Workspace.Root),
		profile:    profFS.NewStore(cfg.Profile.Root, cfg.Profile.Templates),
		bindings:   reg,
		meetings:   catalog,
		events:     events,
		config:     configSvc,
		discordSvc: discordSvc,
	}, nil
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /api/system/runtime", h.handleGetRuntime)
	mux.HandleFunc("GET /api/meetings", h.handleListMeetings)
	mux.HandleFunc("GET /api/meetings/{id}", h.handleGetMeeting)
	mux.HandleFunc("GET /api/principals", h.handleListPrincipals)
	mux.HandleFunc("GET /api/principals/{id}", h.handleGetPrincipal)
	mux.HandleFunc("PUT /api/principals/{id}/files/{filename}", h.handlePutPrincipalFile)
	mux.HandleFunc("GET /api/participants", h.handleListParticipants)
	mux.HandleFunc("POST /api/participants", h.handleCreateParticipant)
	mux.HandleFunc("GET /api/participants/{id}", h.handleGetParticipant)
	mux.HandleFunc("PUT /api/participants/{id}", h.handleUpdateParticipant)
	mux.HandleFunc("DELETE /api/participants/{id}", h.handleDeleteParticipant)
	mux.HandleFunc("PUT /api/participants/{id}/files/{filename}", h.handlePutParticipantFile)
	mux.HandleFunc("GET /api/settings", h.handleGetSettings)
	mux.HandleFunc("PUT /api/settings", h.handlePutSettings)
	mux.HandleFunc("PUT /api/settings/discord-bots", h.handlePutDiscordBots)
	mux.HandleFunc("PUT /api/settings/meet-presets", h.handlePutMeetPresets)
	mux.HandleFunc("POST /api/settings/meet-presets/reset", h.handleResetMeetPresets)
	mux.HandleFunc("PUT /api/settings/meet-casts", h.handlePutMeetCasts)
	mux.HandleFunc("POST /api/settings/meet-casts/reset", h.handleResetMeetCasts)
	mux.HandleFunc("POST /api/settings/discord-bots/refresh-profiles", h.handleRefreshDiscordBotProfiles)
	mux.HandleFunc("GET /api/settings/discord-transport/status", h.handleGetDiscordTransportStatus)
	mux.HandleFunc("POST /api/settings/discord-transport/start", h.handlePostDiscordTransportStart)
	mux.HandleFunc("POST /api/settings/discord-transport/stop", h.handlePostDiscordTransportStop)
	mux.HandleFunc("GET /api/settings/discord-transport/logs", h.handleGetDiscordTransportLogs)
	mux.HandleFunc("POST /api/settings/discord-transport/logs/clear", h.handlePostDiscordTransportLogsClear)
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleListMeetings(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "page_size", 10)

	// Workspace dirs are the source of truth for listing (ADR-0009). SQLite meeting_index
	// may lag behind (e.g. sync without WAL checkpoint, or legacy runs before indexing).
	if h.workspace != nil {
		result, err := h.workspace.ListMeetingsPage(page, pageSize)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	if h.meetings != nil {
		result, err := h.meetings.ListMeetingsPage(r.Context(), page, pageSize)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	writeJSON(w, http.StatusOK, workspace.PaginatedMeetings{
		Meetings: []workspace.MeetingIndex{},
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *Handler) handleGetMeeting(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detail, err := h.workspace.ReadMeetingDetail(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == workspace.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	if h.events != nil {
		if envs, err := h.events.List(r.Context(), id); err == nil && len(envs) > 0 {
			if s, err := meeting.Fold(id, envs); err == nil {
				doc := engine.RenderMeetingDoc(s)
				if detail.Files == nil {
					detail.Files = make(map[string]string)
				}
				detail.Files[workspace.FileMeeting] = doc
				wsfs.EnrichFromMeetingDoc(&detail.MeetingIndex, doc)
			}
		}
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) handleListPrincipals(w http.ResponseWriter, _ *http.Request) {
	list, err := h.profile.ListPrincipals()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	names := h.bindingDisplayNames()
	for i := range list {
		if name, ok := names[list[i].ID]; ok {
			list[i].DisplayName = name
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"principals": list,
	})
}

func (h *Handler) handleGetPrincipal(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detail, err := h.profile.ReadPrincipalDetail(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == profile.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	if names := h.bindingDisplayNames(); names != nil {
		detail.DisplayName = names[detail.ID]
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) handlePutPrincipalFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	content, err := readFileContent(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if err := h.profile.WritePrincipalFile(id, filename, []byte(content)); err != nil {
		status := http.StatusBadRequest
		if err == profile.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleListParticipants(w http.ResponseWriter, _ *http.Request) {
	h.ensureConfiguredParticipants()
	list, err := h.listParticipantsForAPI()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	for i := range list {
		list[i] = h.enrichParticipantIndex(list[i])
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"participants": list,
	})
}

func (h *Handler) handleGetParticipant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.ensureConfiguredParticipants()
	detail, err := h.profile.ReadParticipantDetail(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == profile.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, h.enrichParticipantDetail(detail))
}

func (h *Handler) handlePutParticipantFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	content, err := readFileContent(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.profile.WriteParticipantFile(id, filename, []byte(content)); err != nil {
		status := http.StatusBadRequest
		if err == profile.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readFileContent(r *http.Request) (string, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if len(body) > 0 && body[0] == '{' {
		var payload struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return "", err
		}
		return payload.Content, nil
	}
	return string(body), nil
}

func (h *Handler) bindingDisplayNames() map[string]string {
	if h.bindings == nil {
		return nil
	}
	out := make(map[string]string)
	for _, b := range h.bindings.AllBindings() {
		if b.DisplayName != "" {
			out[b.PrincipalID] = b.DisplayName
		}
	}
	return out
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WithCORS wraps a handler for local web dev (Vite on another origin).
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
