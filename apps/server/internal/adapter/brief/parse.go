package brief

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseDocument(raw []byte) (Document, error) {
	var doc Document
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return Document{}, fmt.Errorf("%w: %v", ErrInvalidYAML, err)
	}
	if strings.TrimSpace(doc.Meta.Title) == "" {
		return Document{}, fmt.Errorf("%w: meta.title is required", ErrInvalidYAML)
	}
	return doc, nil
}

func DocumentToLaunch(doc Document) LaunchDraft {
	return LaunchDraft{
		Topic:   strings.TrimSpace(doc.Topic),
		Brief:   doc.Brief,
		Meeting: doc.Meeting,
	}
}

func ParseMeetingDoc(doc string) (LaunchDraft, error) {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return LaunchDraft{}, fmt.Errorf("%w: empty meeting doc", ErrInvalidYAML)
	}

	out := LaunchDraft{}
	out.Topic = extractSection(doc, "会议主题")
	out.Brief.Goal = extractSection(doc, "会议目标")

	agendaSection := extractSection(doc, "讨论议题")
	if agendaSection == "" {
		agendaSection = extractSection(doc, "议程")
	}
	out.Brief.Agenda = parseNumberedList(agendaSection)
	out.Brief.InScope = extractSection(doc, "讨论范围")
	out.Brief.OutOfScope = extractSection(doc, "不在范围")
	out.Brief.DoneCriteria = extractSection(doc, "完成标准")

	if out.Topic == "" && out.Brief.Goal == "" && len(out.Brief.Agenda) == 0 {
		return LaunchDraft{}, fmt.Errorf("%w: no brief sections found in MEETING.md", ErrInvalidYAML)
	}
	return out, nil
}

func extractSection(doc, heading string) string {
	marker := "## " + heading
	idx := strings.Index(doc, marker)
	if idx < 0 {
		return ""
	}
	rest := doc[idx+len(marker):]
	rest = strings.TrimLeft(rest, " \t\r\n")
	if nl := strings.Index(rest, "\n## "); nl >= 0 {
		rest = rest[:nl]
	}
	if hr := strings.Index(rest, "\n---"); hr >= 0 {
		rest = rest[:hr]
	}
	return strings.TrimSpace(rest)
}

func parseNumberedList(body string) []string {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	var out []string
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if i := strings.Index(line, ". "); i > 0 && i <= 3 {
			prefix := line[:i]
			allDigits := true
			for _, r := range prefix {
				if r < '0' || r > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				line = strings.TrimSpace(line[i+2:])
			}
		}
		line = strings.TrimPrefix(line, "**")
		line = strings.TrimSuffix(line, "**")
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
