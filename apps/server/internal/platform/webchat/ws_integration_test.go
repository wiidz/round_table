//go:build integration

package webchat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	webtransport "round_table/apps/server/internal/adapter/transport/web"
	"round_table/apps/server/internal/platform/config"
)

func TestWebSocketChatIntegration(t *testing.T) {
	svc := New(config.Load(), nil)
	srv := httptest.NewServer(http.HandlerFunc(svc.HandleWebSocket))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var connected webtransport.Frame
	if err := json.Unmarshal(data, &connected); err != nil {
		t.Fatal(err)
	}
	if connected.Type != webtransport.FrameConnected {
		t.Fatalf("frame=%+v", connected)
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"message","content":"会议状态"}`)); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, data, err = conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var reply webtransport.Frame
	if err := json.Unmarshal(data, &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Type != webtransport.FrameMessage {
		t.Fatalf("reply=%+v", reply)
	}
	if reply.Role != webtransport.RoleModerator && reply.Role != webtransport.RoleSystem {
		t.Fatalf("role=%q content=%q", reply.Role, reply.Content)
	}
}
