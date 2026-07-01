package brief_test

import (
	"testing"

	"round_table/apps/server/internal/adapter/brief"
)

func TestSlugTemplateID(t *testing.T) {
	if got := brief.SlugTemplateID("裁决型评审"); got != "裁决型评审" {
		t.Fatalf("got %q", got)
	}
	if got := brief.SlugTemplateID("Game Balance Review"); got != "game_balance_review" {
		t.Fatalf("got %q", got)
	}
}

func TestNextAvailableTemplateID(t *testing.T) {
	taken := map[string]bool{"foo": true, "foo-2": true}
	id := brief.NextAvailableTemplateID("foo", func(id string) bool { return taken[id] })
	if id != "foo-3" {
		t.Fatalf("got %q", id)
	}
}
