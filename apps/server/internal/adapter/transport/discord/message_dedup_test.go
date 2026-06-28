package discord

import "testing"

func TestClaimInboundMessage(t *testing.T) {
	dir := t.TempDir()
	if !ClaimInboundMessage(dir, "msg-1") {
		t.Fatal("first claim should succeed")
	}
	if ClaimInboundMessage(dir, "msg-1") {
		t.Fatal("duplicate claim should fail")
	}
	if !ClaimInboundMessage(dir, "msg-2") {
		t.Fatal("second message should succeed")
	}
}
