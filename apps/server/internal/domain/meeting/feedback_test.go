package meeting

import (
	"strings"
	"testing"
)

func TestFormatPrincipalFeedback_itemNotes(t *testing.T) {
	got := FormatPrincipalFeedback("", map[int]string{2: "技能树需重算"})
	if got == "" || !strings.Contains(got, "Item 2") || !strings.Contains(got, "技能树需重算") {
		t.Fatalf("got=%q", got)
	}
	got = FormatPrincipalFeedback("整体方向 OK", map[int]string{1: "通过"})
	if !strings.Contains(got, "整体方向 OK") || !strings.Contains(got, "Item 1") {
		t.Fatalf("got=%q", got)
	}
}
