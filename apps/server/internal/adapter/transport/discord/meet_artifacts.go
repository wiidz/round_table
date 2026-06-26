package discord

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	wsfs "round_table/apps/server/internal/adapter/workspace/fs"
	"round_table/apps/server/internal/adapter/workspace"
	"round_table/apps/server/internal/domain/meeting"
)

const artifactExcerptRunes = 800
const artifactFetchExcerptRunes = 1800

func (r *MeetRunner) postMeetArtifacts(ctx context.Context, channelID string, final meeting.State, meetingID string, loc Locale) {
	if r.Bots == nil || r.Bots.Default == nil || r.Cfg.Workspace.Root == "" {
		return
	}
	ws := wsfs.NewStore(r.Cfg.Workspace.Root)

	type artifactSpec struct {
		title string
		path  string
	}
	specs := []artifactSpec{
		{artifactTitle(loc, "minutes"), workspace.FileMinutes},
	}
	if final.IsDeliberation() {
		specs = append(specs,
			artifactSpec{artifactTitle(loc, "draft"), "artifacts/design-draft.md"},
			artifactSpec{artifactTitle(loc, "open"), "artifacts/open-questions.md"},
		)
	} else {
		specs = append(specs, artifactSpec{artifactTitle(loc, "conclusion"), "artifacts/minutes.md"})
	}

	for _, spec := range specs {
		data, err := ws.Read(meetingID, spec.path)
		if err != nil || len(strings.TrimSpace(string(data))) == 0 {
			continue
		}
		header := formatArtifactHeader(loc, spec.title, spec.path, meetingID)
		body := excerptArtifact(string(data), artifactExcerptRunes, loc)
		footer := artifactFetchHint(loc)
		SendLong(r.Bots.Default, ctx, channelID, header+"\n\n"+body+"\n\n"+footer)
	}
}

func (r *MeetRunner) fetchArtifact(ctx context.Context, channelID, meetingID string, kind string, loc Locale) (string, error) {
	if r.Bots == nil || r.Bots.Default == nil || r.Cfg.Workspace.Root == "" {
		return "", nil
	}
	ws := wsfs.NewStore(r.Cfg.Workspace.Root)
	path, kindKey, ok := artifactPathForKind(kind)
	if !ok {
		return artifactFetchUsageText(loc), nil
	}
	data, err := ws.Read(meetingID, path)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return artifactFetchMissingText(loc, kind, meetingID), nil
	}
	displayTitle := artifactTitle(loc, kindKey)
	header := formatArtifactHeader(loc, displayTitle, path, meetingID)
	body := excerptArtifact(string(data), artifactFetchExcerptRunes, loc)
	SendLong(r.Bots.Default, ctx, channelID, header+"\n\n"+body)
	return artifactFetchSentText(loc, kindKey), nil
}

func artifactPathForKind(kind string) (path, title string, ok bool) {
	switch kind {
	case "minutes":
		return workspace.FileMinutes, "minutes", true
	case "draft":
		return "artifacts/design-draft.md", "draft", true
	case "open":
		return "artifacts/open-questions.md", "open", true
	case "conclusion":
		return "artifacts/minutes.md", "conclusion", true
	default:
		return "", "", false
	}
}

func artifactFetchHint(loc Locale) string {
	if loc == LocaleZH {
		return "📎 完整版：**获取纪要** · **获取草案** · **获取待决** · **获取结论**"
	}
	return "📎 Full text: **get minutes** · **get draft** · **get open** · **get conclusion**"
}

func artifactTitle(loc Locale, kind string) string {
	if loc == LocaleZH {
		switch kind {
		case "minutes":
			return "📋 会议纪要"
		case "draft":
			return "📐 方案草案"
		case "open":
			return "❓ 待决事项"
		case "conclusion":
			return "📋 会议结论"
		}
	}
	switch kind {
	case "minutes":
		return "📋 Minutes"
	case "draft":
		return "📐 Design draft"
	case "open":
		return "❓ Open questions"
	case "conclusion":
		return "📋 Conclusion"
	default:
		return kind
	}
}

func formatArtifactHeader(loc Locale, title, relPath, meetingID string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("%s · `%s`\n📁 `%s`", title, meetingID, relPath)
	}
	return fmt.Sprintf("%s · `%s`\n📁 `%s`", title, meetingID, relPath)
}

func excerptArtifact(body string, maxRunes int, loc Locale) string {
	body = strings.TrimSpace(body)
	if maxRunes <= 0 {
		return body
	}
	if utf8.RuneCountInString(body) <= maxRunes {
		return body
	}
	clipped := clipRunes(body, maxRunes-1)
	note := "\n\n… _(excerpt truncated)_"
	if loc == LocaleZH {
		note = "\n\n… _(内容过长，已节选；完整版见工作区)_"
	}
	return clipped + "…" + note
}

func clipRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}
