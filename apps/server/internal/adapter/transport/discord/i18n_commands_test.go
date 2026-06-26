package discord

import (
	"strings"
	"testing"
)

func TestHelpTextZH(t *testing.T) {
	h := &CommandHandler{Prefix: "!rt "}
	h.Meet = &MeetRunner{}
	h.Meet.Discord.Locale = "zh"
	got := h.helpText()
	for _, want := range []string{"RoundTable Discord 指令", "principal bind", "meet"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestHelpTextEN(t *testing.T) {
	h := &CommandHandler{Prefix: "!rt "}
	h.Meet = &MeetRunner{}
	h.Meet.Discord.Locale = "en"
	got := h.helpText()
	for _, want := range []string{"RoundTable Discord commands", "principal bind", "meet"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}
