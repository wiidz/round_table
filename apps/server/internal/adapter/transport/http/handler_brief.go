package httptransport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"round_table/apps/server/internal/adapter/brief"
	"round_table/apps/server/internal/adapter/workspace"
)

func (h *Handler) handleListBriefTemplates(w http.ResponseWriter, _ *http.Request) {
	if h.briefs == nil {
		writeJSON(w, http.StatusOK, map[string]any{"templates": []brief.TemplateIndex{}})
		return
	}
	list, err := h.briefs.ListTemplates()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"templates": list})
}

func (h *Handler) handleGetBriefTemplate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if h.briefs == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": brief.ErrNotFound.Error()})
		return
	}
	detail, err := h.briefs.ReadTemplate(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == brief.ErrNotFound || err == brief.ErrInvalidID {
			status = http.StatusNotFound
		}
		if err == brief.ErrInvalidYAML {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *Handler) handlePostBriefTemplate(w http.ResponseWriter, r *http.Request) {
	if h.briefs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "brief store unavailable"})
		return
	}
	raw, err := readBriefTemplatePayload(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	id, err := h.briefs.CreateTemplate(raw)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "status": "ok"})
}

func (h *Handler) handlePutBriefTemplate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if h.briefs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "brief store unavailable"})
		return
	}
	raw, err := readBriefTemplatePayload(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.briefs.WriteTemplate(id, raw); err != nil {
		status := http.StatusBadRequest
		if err == brief.ErrBuiltinReadonly {
			status = http.StatusForbidden
		}
		if err == brief.ErrInvalidID {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readBriefTemplatePayload(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	var payload struct {
		Document *brief.Document `json:"document"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Document == nil {
		return nil, fmt.Errorf("document required")
	}
	return brief.MarshalDocument(*payload.Document)
}

func (h *Handler) handlePostCloneBrief(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		MeetingID string `json:"meeting_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.MeetingID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "meeting_id required"})
		return
	}
	if h.workspace == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "workspace unavailable"})
		return
	}
	detail, err := h.workspace.ReadMeetingDetail(payload.MeetingID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == workspace.ErrNotFound {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	doc := detail.Files[workspace.FileMeeting]
	if doc == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "MEETING.md not found"})
		return
	}
	if h.briefs == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "brief store unavailable"})
		return
	}
	launch, err := h.briefs.CloneFromMeetingDoc(doc)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"launch": launch})
}
