package profile_test

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/profile"
)

func TestRenderAndParseUserMD(t *testing.T) {
	raw := profile.RenderUserMD(profile.UserProfile{
		Language:     "en-US",
		Confirmation: "review lists",
		Context:      "Game studio",
	})
	got := profile.ParseUserMD(raw)
	if got.Language != "en-US" || got.Confirmation != "review lists" || got.Context != "Game studio" {
		t.Fatalf("parse: %+v", got)
	}
	again := profile.RenderUserMD(got)
	if strings.TrimSpace(again) != strings.TrimSpace(raw) {
		t.Fatalf("round-trip mismatch:\n%s\n---\n%s", raw, again)
	}
}
