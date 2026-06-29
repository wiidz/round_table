package web

import (
	"testing"
)

func TestHubSendToRegisteredSession(t *testing.T) {
	h := NewHub()
	sessionID, outbound := h.Register()
	defer h.Unregister(sessionID)

	done := make(chan struct{})
	go func() {
		defer close(done)
		frame := <-outbound
		if frame.Type != FrameMessage {
			t.Errorf("type=%q", frame.Type)
		}
		if frame.Content != "hello" {
			t.Errorf("content=%q", frame.Content)
		}
	}()

	if err := h.Send(t.Context(), sessionID, RoleModerator, "hello"); err != nil {
		t.Fatal(err)
	}
	<-done
}

func TestHubSendUnknownSessionNoError(t *testing.T) {
	h := NewHub()
	if err := h.Send(t.Context(), "missing", RoleModerator, "x"); err != nil {
		t.Fatal(err)
	}
}
