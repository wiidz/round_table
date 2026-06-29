//go:build integration

package webchat

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	webtransport "round_table/apps/server/internal/adapter/transport/web"
)

// TestLiveWebSocketChat dials the running local server (make server-dev on :7777).
func TestLiveWebSocketChat(t *testing.T) {
	conn, resp, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:7777/api/chat/ws", nil)
	if err != nil {
		if resp != nil {
			t.Log("status", resp.Status)
		}
		t.Skip("local server not running:", err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("connected", string(data))

	_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"message","content":"会议状态"}`))
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, data, err = conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	var frame webtransport.Frame
	_ = json.Unmarshal(data, &frame)
	t.Logf("status reply role=%q content=%q", frame.Role, frame.Content)
}
