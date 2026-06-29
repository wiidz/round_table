package discord

import (
	"strings"
	"testing"
)

func TestLocalizeProgressZH(t *testing.T) {
	got := localizeProgressZH("★ synthesis completed resolved_by=synthesis open_questions=8")
	if !strings.Contains(got, "设计草案合成完成") || !strings.Contains(got, "研讨合成") || !strings.Contains(got, "8") {
		t.Fatalf("got=%q", got)
	}

	got = localizeProgressZH("■ debate round 2 completed")
	if got != "✅ 第 2 轮研讨结束" {
		t.Fatalf("got=%q", got)
	}

	got = localizeProgressZH("■ meeting finished outcome=completed")
	if !strings.Contains(got, "已完成") {
		t.Fatalf("got=%q", got)
	}

	got = localizeProgressZH("↩ confirmation rejected cycle=2 — starting round 5")
	if got != "↩ 确认呈报第 2 次驳回 · 即将开始第 5 轮研讨" {
		t.Fatalf("got=%q", got)
	}
}

func TestFormatStreamStartZH(t *testing.T) {
	got := formatStreamStart(streamMeta{
		ParticipantID: "designer",
		Phase:         "deliberation",
		Detail:        "turn (1/4) · round 1",
	}, LocaleZH)
	if !strings.Contains(got, "方案") || !strings.Contains(got, "研讨") || !strings.Contains(got, "1/4") {
		t.Fatalf("got=%q", got)
	}
}

func TestParseLocale(t *testing.T) {
	if ParseLocale("") != LocaleEN {
		t.Fatal("default en")
	}
	if ParseLocale("zh-CN") != LocaleZH {
		t.Fatal("expected zh")
	}
}
