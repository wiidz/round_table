package discord

import (
	"fmt"
	"strings"

	"round_table/apps/server/internal/domain/event"
)

// meetBrief is the Principal-authored meeting task book (Goal + Agenda + scope).
type meetBrief struct {
	Goal         string
	AgendaTitles []string
	InScope      string
	OutOfScope   string
	DoneCriteria string
}

func (b meetBrief) hasContent() bool {
	return strings.TrimSpace(b.Goal) != "" ||
		len(b.AgendaTitles) > 0 ||
		strings.TrimSpace(b.InScope) != "" ||
		strings.TrimSpace(b.OutOfScope) != "" ||
		strings.TrimSpace(b.DoneCriteria) != ""
}

func (c meetLaunchConfig) engineGoal() string {
	return formatBriefForEngineGoal(c.Brief)
}

func (c meetLaunchConfig) engineAgenda() []event.AgendaItem {
	return agendaTitlesToItems(c.Brief.AgendaTitles)
}

func formatBriefForEngineGoal(b meetBrief) string {
	if !b.hasContent() {
		return ""
	}
	var parts []string
	if g := strings.TrimSpace(b.Goal); g != "" {
		parts = append(parts, g)
	}
	if d := strings.TrimSpace(b.DoneCriteria); d != "" {
		parts = append(parts, "完成标准："+d)
	}
	if in := strings.TrimSpace(b.InScope); in != "" {
		parts = append(parts, "讨论范围："+in)
	}
	if out := strings.TrimSpace(b.OutOfScope); out != "" {
		parts = append(parts, "不在范围："+out)
	}
	return strings.Join(parts, "\n\n")
}

func agendaTitlesToItems(titles []string) []event.AgendaItem {
	var out []event.AgendaItem
	for i, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		id := slugAgendaID(title, i+1)
		out = append(out, event.AgendaItem{ID: id, Title: title})
	}
	return out
}

func slugAgendaID(title string, index int) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			if b.Len() > 0 && b.String()[b.Len()-1] != '_' {
				b.WriteByte('_')
			}
		}
	}
	s := strings.Trim(b.String(), "_")
	if s == "" {
		return fmt.Sprintf("agenda_%d", index)
	}
	if len(s) > 24 {
		s = s[:24]
	}
	return s
}

func isBriefSkipToken(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == "-" || s == "—" || strings.EqualFold(s, "skip") || s == "跳过"
}

func parseAgendaLines(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	repl := strings.NewReplacer("；", "\n", ";", "\n", "，", "\n", ",", "\n")
	text = repl.Replace(text)
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = stripListPrefix(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func stripListPrefix(s string) string {
	s = strings.TrimSpace(s)
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return s
	}
	rest := strings.TrimSpace(s[i:])
	for _, sep := range []string{".", "、", ")", "）"} {
		if strings.HasPrefix(rest, sep) {
			return strings.TrimSpace(strings.TrimPrefix(rest, sep))
		}
	}
	return s
}

func (r *MeetRunner) advanceBriefGoal(sess meetSetupSession, content string, loc Locale) (meetSetupSession, string) {
	if !isBriefSkipToken(content) {
		sess.config.Brief.Goal = strings.TrimSpace(content)
	}
	sess.step = setupStepBriefAgenda
	return sess, formatAskBriefAgendaPrompt(loc)
}

func (r *MeetRunner) advanceBriefAgenda(sess meetSetupSession, content string, loc Locale) (meetSetupSession, string) {
	if !isBriefSkipToken(content) {
		sess.config.Brief.AgendaTitles = parseAgendaLines(content)
	}
	sess.step = setupStepBriefScope
	return sess, formatAskBriefScopePrompt(loc)
}

func (r *MeetRunner) advanceBriefScope(sess meetSetupSession, content string, loc Locale) (meetSetupSession, string) {
	if !isBriefSkipToken(content) {
		in, out, done := parseBriefScope(content)
		sess.config.Brief.InScope = in
		sess.config.Brief.OutOfScope = out
		sess.config.Brief.DoneCriteria = done
	}
	sess.step = setupStepPresetMenu
	return sess, r.promptPresetMenu(loc, sess.config)
}

func parseBriefScope(text string) (inScope, outScope, done string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", "", ""
	}
	var other []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(line, "讨论范围") || strings.HasPrefix(line, "讨论：") || strings.HasPrefix(line, "讨论:") || strings.HasPrefix(lower, "in scope"):
			inScope = afterLabelColon(line)
		case strings.HasPrefix(line, "不在范围") || strings.HasPrefix(line, "不讨论") || strings.HasPrefix(line, "排除") || strings.HasPrefix(lower, "out of scope"):
			outScope = afterLabelColon(line)
		case strings.HasPrefix(line, "完成标准") || strings.HasPrefix(line, "产出") || strings.HasPrefix(lower, "done"):
			done = afterLabelColon(line)
		default:
			other = append(other, line)
		}
	}
	if inScope == "" && outScope == "" && done == "" && len(other) == 1 {
		inScope = other[0]
	} else if inScope == "" && len(other) > 0 {
		inScope = strings.Join(other, "；")
	}
	return inScope, outScope, done
}

func afterLabelColon(line string) string {
	for _, sep := range []string{"：", ":"} {
		if idx := strings.Index(line, sep); idx >= 0 {
			return strings.TrimSpace(line[idx+len(sep):])
		}
	}
	return strings.TrimSpace(line)
}

func formatAskBriefGoalPrompt(loc Locale, topic string) string {
	if loc == LocaleZH {
		return fmt.Sprintf(`📋 **会议简报 · 1/3 目标**

📌 主题：%s

这场会要**交付什么**？（例如：输出技能框架草案 + 3 条待决项）
发送 **-** 跳过（使用默认目标）`, topic)
	}
	return fmt.Sprintf(`📋 **Meeting brief · 1/3 goal**

📌 Topic: %s

What should this meeting deliver? Send **-** to skip.`, topic)
}

func formatAskBriefAgendaPrompt(loc Locale) string {
	if loc == LocaleZH {
		return `📋 **会议简报 · 2/3 议程**

请列出要覆盖的子问题（每行一条，或 1、2、3 编号），例如：
1）职业定位
2）核心技能循环
3）与战士的差异

发送 **-** 跳过`
	}
	return `📋 **Meeting brief · 2/3 agenda**

List sub-questions (one per line). Send **-** to skip.`
}

func formatAskBriefScopePrompt(loc Locale) string {
	if loc == LocaleZH {
		return `📋 **会议简报 · 3/3 范围**

可选，一行或多行，例如：
讨论范围：定位、核心循环
不在范围：数值表、全职业平衡
完成标准：每个议程至少 1 条结论或待决

发送 **-** 跳过`
	}
	return `📋 **Meeting brief · 3/3 scope**

Optional lines: 讨论范围 / 不在范围 / 完成标准. Send **-** to skip.`
}

func formatBriefSummary(loc Locale, cfg meetLaunchConfig) string {
	if !cfg.Brief.hasContent() {
		return ""
	}
	var lines []string
	if loc == LocaleZH {
		lines = append(lines, "📋 **已记录简报**")
	} else {
		lines = append(lines, "📋 **Brief recorded**")
	}
	if body := formatBriefSummaryBody(loc, cfg.Brief); body != "" {
		lines = append(lines, body)
	}
	return strings.Join(lines, "\n")
}

func formatBriefLaunchBlock(loc Locale, cfg meetLaunchConfig) string {
	if !cfg.Brief.hasContent() {
		return ""
	}
	var lines []string
	if loc == LocaleZH {
		lines = append(lines, "📋 **会议简报**")
	} else {
		lines = append(lines, "📋 **Meeting brief**")
	}
	if body := formatBriefSummaryBody(loc, cfg.Brief); body != "" {
		lines = append(lines, body)
	}
	if loc == LocaleZH {
		lines = append(lines, "_完整版见 MEETING.md_")
	} else {
		lines = append(lines, "_Full version in MEETING.md_")
	}
	return strings.Join(lines, "\n")
}

func formatBriefSummaryBody(loc Locale, b meetBrief) string {
	var lines []string
	if g := strings.TrimSpace(b.Goal); g != "" {
		if loc == LocaleZH {
			lines = append(lines, "- 🎯 目标："+g)
		} else {
			lines = append(lines, "- 🎯 Goal: "+g)
		}
	}
	if len(b.AgendaTitles) > 0 {
		if loc == LocaleZH {
			lines = append(lines, "- 📑 议程："+strings.Join(b.AgendaTitles, " · "))
		} else {
			lines = append(lines, "- 📑 Agenda: "+strings.Join(b.AgendaTitles, " · "))
		}
	}
	if in := strings.TrimSpace(b.InScope); in != "" {
		if loc == LocaleZH {
			lines = append(lines, "- ✅ 讨论："+in)
		} else {
			lines = append(lines, "- ✅ In scope: "+in)
		}
	}
	if out := strings.TrimSpace(b.OutOfScope); out != "" {
		if loc == LocaleZH {
			lines = append(lines, "- ⛔ 不讨论："+out)
		} else {
			lines = append(lines, "- ⛔ Out of scope: "+out)
		}
	}
	if d := strings.TrimSpace(b.DoneCriteria); d != "" {
		if loc == LocaleZH {
			lines = append(lines, "- ✔️ 完成标准："+d)
		} else {
			lines = append(lines, "- ✔️ Done when: "+d)
		}
	}
	return strings.Join(lines, "\n")
}

func formatModeratorSetupWithBrief(loc Locale, prefix string, all []meetPreset, cfg meetLaunchConfig) string {
	head := formatBriefSummary(loc, cfg)
	body := formatModeratorSetupPrompt(loc, prefix, all)
	if head == "" {
		return body
	}
	return head + "\n\n" + body
}

func mergePresetLaunchConfig(base, preset meetLaunchConfig) meetLaunchConfig {
	out := preset
	out.ParticipantsSummary = base.ParticipantsSummary
	out.ParticipantIDs = append([]string(nil), base.ParticipantIDs...)
	out.Brief = base.Brief
	return out
}

func (r *MeetRunner) promptPresetMenu(loc Locale, cfg meetLaunchConfig) string {
	prefix := strings.TrimSpace(r.dc().CommandPrefix)
	if prefix == "" {
		prefix = "!rt"
	}
	return formatModeratorSetupWithBrief(loc, prefix+" ", r.meetPresets(loc), cfg)
}
