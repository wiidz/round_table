package meeting

import (
	"fmt"
	"sort"
	"strings"
)

// FormatPrincipalFeedback merges overall feedback with numbered Item notes for debate prompts.
func FormatPrincipalFeedback(feedback string, itemNotes map[int]string) string {
	feedback = strings.TrimSpace(feedback)
	if len(itemNotes) == 0 {
		return feedback
	}
	var b strings.Builder
	if feedback != "" {
		b.WriteString(feedback)
		b.WriteString("\n\n")
	}
	b.WriteString("Item-specific notes:\n")
	keys := make([]int, 0, len(itemNotes))
	for k := range itemNotes {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "- Item %d: %s\n", k, strings.TrimSpace(itemNotes[k]))
	}
	return strings.TrimSpace(b.String())
}
