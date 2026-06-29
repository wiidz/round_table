package httptransport

import (
	"net/http"

	"round_table/apps/server/internal/platform/procstats"
)

type runtimeResponse struct {
	Server           serverRuntime      `json:"server"`
	DiscordTransport procstats.Snapshot `json:"discord_transport,omitempty"`
}

type serverRuntime struct {
	procstats.Snapshot
	ListenAddr string `json:"listen_addr,omitempty"`
}

func (h *Handler) handleGetRuntime(w http.ResponseWriter, _ *http.Request) {
	serverSnap := serverRuntime{Snapshot: procstats.ServerSnapshot()}
	if h.config != nil {
		serverSnap.ListenAddr = h.config.Current().Addr()
	}
	resp := runtimeResponse{
		Server: serverSnap,
	}
	if h.discordSvc != nil {
		st := h.discordSvc.Status()
		if st.Running && st.PID > 0 {
			resp.DiscordTransport = procstats.Snapshot{
				PID:           st.PID,
				UptimeSeconds: st.UptimeSeconds,
				MemoryBytes:   st.MemoryBytes,
				MemorySource:  st.MemorySource,
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
