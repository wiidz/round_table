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
		if frame.Turn != 1 {
			t.Errorf("turn=%d want 1", frame.Turn)
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

func TestHubTurnAssignment(t *testing.T) {
	h := NewHub()
	sessionID, outbound := h.Register()
	defer h.Unregister(sessionID)

	wantTurns := []int{1, 2, 0, 3}
	contents := []string{"user", "mod", "system", "mod2"}
	roles := []string{RoleUser, RoleModerator, RoleSystem, RoleModerator}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i, want := range wantTurns {
			frame := <-outbound
			if frame.Content != contents[i] {
				t.Errorf("content[%d]=%q", i, frame.Content)
			}
			if frame.Turn != want {
				t.Errorf("turn[%d]=%d want %d", i, frame.Turn, want)
			}
		}
	}()

	for i, role := range roles {
		out := Outbound{Role: role, Content: contents[i]}
		if role == RoleParticipant {
			out.AuthorID = "dev"
		}
		if role == RoleUser {
			out.AuthorID = "web"
			out.AuthorName = "Web"
		}
		if err := h.SendOutbound(t.Context(), sessionID, out); err != nil {
			t.Fatal(err)
		}
	}
	<-done
}

func TestRoleAssignsTurn(t *testing.T) {
	if !roleAssignsTurn(RoleModerator) || !roleAssignsTurn(RoleParticipant) || !roleAssignsTurn(RoleUser) {
		t.Fatal("moderator/participant/user should assign turn")
	}
	if roleAssignsTurn(RoleSystem) {
		t.Fatal("system should not assign turn")
	}
}
