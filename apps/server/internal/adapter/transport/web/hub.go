package web

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Hub routes outbound chat messages to connected browser sessions.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*clientConn
}

type clientConn struct {
	sessionID string
	send      chan Frame
}

// NewHub returns an empty session hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[string]*clientConn)}
}

// Register adds a session and returns its ID and outbound channel.
func (h *Hub) Register() (sessionID string, outbound <-chan Frame) {
	id := uuid.NewString()
	ch := make(chan Frame, 32)
	h.mu.Lock()
	h.clients[id] = &clientConn{sessionID: id, send: ch}
	h.mu.Unlock()
	return id, ch
}

// Unregister removes a session.
func (h *Hub) Unregister(sessionID string) {
	h.mu.Lock()
	if c, ok := h.clients[sessionID]; ok {
		delete(h.clients, sessionID)
		close(c.send)
	}
	h.mu.Unlock()
}

// Outbound is metadata for a chat message pushed to the browser.
type Outbound struct {
	Role       string
	Content    string
	AuthorID   string
	AuthorName string
	At         string
}

// Send posts a message to one session.
func (h *Hub) Send(ctx context.Context, sessionID, role, content string) error {
	return h.SendOutbound(ctx, sessionID, Outbound{
		Role:    role,
		Content: content,
	})
}

// SendOutbound posts a fully attributed message to one session.
func (h *Hub) SendOutbound(_ context.Context, sessionID string, msg Outbound) error {
	if msg.Content == "" {
		return nil
	}
	at := stringsTrimOr(msg.At, time.Now().UTC().Format(time.RFC3339Nano))
	frame := Frame{
		Type:       FrameMessage,
		ID:         uuid.NewString(),
		SessionID:  sessionID,
		Role:       msg.Role,
		AuthorID:   msg.AuthorID,
		AuthorName: msg.AuthorName,
		At:         at,
		Content:    msg.Content,
	}
	h.mu.RLock()
	c, ok := h.clients[sessionID]
	h.mu.RUnlock()
	if !ok {
		return nil
	}
	select {
	case c.send <- frame:
	default:
		// drop if client is slow; avoid blocking handler
	}
	return nil
}

func stringsTrimOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// MarshalConnected returns the first frame after upgrade.
func MarshalConnected(sessionID string) ([]byte, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return json.Marshal(Frame{
		Type:      FrameConnected,
		SessionID: sessionID,
		Role:      RoleSystem,
		At:        now,
		Content:   "已连接 RoundTable 浏览器聊天。无需 Principal 绑定；可发送「会议状态」或自然语言提问。",
	})
}
