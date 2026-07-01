package webchat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"round_table/apps/server/internal/adapter/transport"
	webtransport "round_table/apps/server/internal/adapter/transport/web"
	"round_table/apps/server/internal/platform/config"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

// Service hosts browser chat sessions (no Principal binding).
type Service struct {
	hub     *webtransport.Hub
	handler *webtransport.Handler
}

// New wires a web-only handler; reg is intentionally omitted — browser chat has no Principal scope.
func New(cfg config.Config, configSvc *config.Service) *Service {
	hub := webtransport.NewHub()
	return &Service{
		hub:     hub,
		handler: webtransport.NewHandler(cfg, configSvc, hub),
	}
}

// HandleWebSocket upgrades HTTP to the chat protocol.
func (s *Service) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	if s == nil || s.handler == nil {
		http.Error(w, "chat unavailable", http.StatusServiceUnavailable)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("web chat: upgrade: %v", err)
		return
	}

	sessionID, outbound := s.hub.Register()
	defer func() {
		s.hub.Unregister(sessionID)
		_ = conn.Close()
	}()

	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	connectedPayload, err := webtransport.MarshalConnected(sessionID)
	if err != nil {
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, connectedPayload); err != nil {
		return
	}

	var writeMu sync.Mutex
	writeFrame := func(frame webtransport.Frame) error {
		data, err := json.Marshal(frame)
		if err != nil {
			return err
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteMessage(websocket.TextMessage, data)
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case frame, ok := <-outbound:
				if !ok {
					return
				}
				if err := writeFrame(frame); err != nil {
					cancel()
					return
				}
			case <-ticker.C:
				writeMu.Lock()
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				err := conn.WriteMessage(websocket.PingMessage, nil)
				writeMu.Unlock()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))

		var in webtransport.Frame
		if err := json.Unmarshal(data, &in); err != nil {
			_ = writeFrame(webtransport.Frame{
				Type:  webtransport.FrameError,
				Error: "invalid message format",
			})
			continue
		}
		if in.Type != webtransport.FrameMessage {
			continue
		}
		content := strings.TrimSpace(in.Content)
		if content == "" {
			continue
		}

		msg := transport.Inbound{
			Platform:   "web",
			ChannelID:  sessionID,
			AuthorID:   sessionID,
			AuthorName: "Web",
			Content:    content,
		}

		now := time.Now().UTC().Format(time.RFC3339Nano)
		if err := s.hub.SendOutbound(r.Context(), sessionID, webtransport.Outbound{
			ID:         in.ID,
			Role:       webtransport.RoleUser,
			Content:    content,
			AuthorID:   sessionID,
			AuthorName: "Web",
			At:         now,
		}); err != nil {
			return
		}

		s.hub.SendTyping(r.Context(), sessionID, webtransport.RoleModerator, "moderator", "司仪")
		reply, handleErr := s.handler.Handle(r.Context(), msg)
		if handleErr != nil {
			_ = writeFrame(webtransport.Frame{
				Type:  webtransport.FrameError,
				Error: handleErr.Error(),
			})
			continue
		}
		if strings.TrimSpace(reply.Content) != "" {
			role := webtransport.RoleSystem
			authorID := ""
			authorName := ""
			if reply.AsModerator {
				role = webtransport.RoleModerator
				authorID = "moderator"
				authorName = "司仪"
			}
			if err := s.hub.SendOutbound(r.Context(), sessionID, webtransport.Outbound{
				Role:       role,
				Content:    reply.Content,
				AuthorID:   authorID,
				AuthorName: authorName,
				At:         now,
			}); err != nil {
				return
			}
		}
	}
}
