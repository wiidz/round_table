package discord

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitDiscordMessages_cjkNotByBytes(t *testing.T) {
	// 700 CJK chars ≈ 2100 bytes — must not be treated as over limit when splitting for send.
	var b strings.Builder
	for i := 0; i < 700; i++ {
		b.WriteRune('测')
	}
	text := b.String()
	if utf8.RuneCountInString(text) != 700 {
		t.Fatal("setup")
	}
	parts := splitDiscordMessages(text, maxMessageRunes)
	if len(parts) != 1 {
		t.Fatalf("expected single part for 700 runes, got %d", len(parts))
	}
}

func TestSplitDiscordMessages_longCjk(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 2500; i++ {
		b.WriteString("开放问题条目。")
	}
	parts := splitDiscordMessages(b.String(), maxMessageRunes)
	if len(parts) < 2 {
		t.Fatalf("expected multiple parts, got %d", len(parts))
	}
	rejoined := strings.Join(parts, "")
	if !strings.HasPrefix(rejoined, "开放问题条目。") {
		t.Fatal("lost prefix")
	}
	if strings.Contains(rejoined, "�") {
		t.Fatal("garbled replacement char in rejoined text")
	}
	for _, p := range parts {
		if utf8.RuneCountInString(p) > maxMessageRunes {
			t.Fatalf("part exceeds limit: %d runes", utf8.RuneCountInString(p))
		}
	}
}

func TestClipMessageRunes(t *testing.T) {
	var b strings.Builder
	for i := 0; i < maxMessageRunes+10; i++ {
		b.WriteRune('题')
	}
	clipped := clipMessageRunes(b.String())
	if utf8.RuneCountInString(clipped) != maxMessageRunes {
		t.Fatalf("runes=%d", utf8.RuneCountInString(clipped))
	}
	if strings.HasSuffix(clipped, "…") {
		// ok
	} else {
		t.Fatal("expected ellipsis suffix")
	}
}

func TestSplitDiscordMessages(t *testing.T) {
	short := "hello"
	if parts := splitDiscordMessages(short, 2000); len(parts) != 1 || parts[0] != short {
		t.Fatalf("got %v", parts)
	}

	var b strings.Builder
	for i := 0; i < 30; i++ {
		b.WriteString("这是第 ")
		b.WriteString(strings.Repeat("x", 80))
		b.WriteString(" 条开放问题。\n")
	}
	long := b.String()
	parts := splitDiscordMessages(long, 500)
	if len(parts) < 2 {
		t.Fatalf("expected split, got %d parts", len(parts))
	}
	rejoined := strings.Join(parts, "\n")
	if !strings.Contains(rejoined, "开放问题") {
		t.Fatal("lost content")
	}
}
