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
	return s == "" || s == "0" || strings.EqualFold(s, "skip") || s == "跳过"
}

func isBriefConfirmToken(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	switch lower {
	case "1", "确认", "confirm", "ok", "yes", "y", "是", "同意", "好", "好的":
		return true
	default:
		return false
	}
}

func briefStepActionHint(loc Locale, hasDraft, templateFlow bool) string {
	var b strings.Builder
	if loc == LocaleZH {
		b.WriteString("**回复方式：**\n")
		if hasDraft || templateFlow {
			if hasDraft {
				b.WriteString("**1** / **确认** — 采用以上内容\n")
			} else {
				b.WriteString("**1** / **确认** — 进入下一步（本步留空）\n")
			}
			b.WriteString("**0** / **跳过** — 本步留空\n")
			b.WriteString("**文字** — 直接发送以修改或填写")
			return b.String()
		}
		b.WriteString("**0** / **跳过** — 本步留空\n")
		b.WriteString("**文字** — 直接发送以填写")
		return b.String()
	}

	b.WriteString("**How to reply:**\n")
	if hasDraft || templateFlow {
		if hasDraft {
			b.WriteString("**1** / **confirm** — accept the content above\n")
		} else {
			b.WriteString("**1** / **confirm** — continue (leave this step empty)\n")
		}
		b.WriteString("**0** / **skip** — leave this step empty\n")
		b.WriteString("**text** — send a message to edit or fill in")
		return b.String()
	}
	b.WriteString("**0** / **skip** — leave this step empty\n")
	b.WriteString("**text** — send a message to fill in")
	return b.String()
}

func briefTemplatePickActionHint(loc Locale) string {
	if loc == LocaleZH {
		return `**回复方式：**
**编号** 或 **模板 ID** — 选用模板（如 **1** 或 ` + "`decision-review`" + `）
**0** / **跳过** — 不选模板，手动填写`
	}
	return `**How to reply:**
**number** or **template id** — pick a template (e.g. **1** or ` + "`decision-review`" + `)
**0** / **skip** — continue without a template`
}

func briefStepGoalTopicNote(loc Locale) string {
	if loc == LocaleZH {
		return "_（本步仅影响会议目标，**不会修改上方主题**）_"
	}
	return "_(This step edits the goal only; **topic above is unchanged**.)_"
}

func parseAgendaLines(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	// Only split on semicolons; commas appear inside topic text (e.g. parenthetical lists).
	repl := strings.NewReplacer("；", "\n", ";", "\n")
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
	sess.config.Brief.Goal = applyBriefGoalInput(sess.config.Brief.Goal, content)
	sess.step = setupStepBriefAgenda
	return sess, formatAskBriefAgendaPrompt(loc, sess.config.Brief, sess.briefTemplateID != "")
}

func (r *MeetRunner) advanceBriefAgenda(sess meetSetupSession, content string, loc Locale) (meetSetupSession, string) {
	sess.config.Brief.AgendaTitles = applyBriefAgendaInput(sess.config.Brief.AgendaTitles, content)
	sess.step = setupStepBriefScope
	return sess, formatAskBriefScopePrompt(loc, sess.config.Brief, sess.briefTemplateID != "")
}

func (r *MeetRunner) advanceBriefScope(sess meetSetupSession, content string, loc Locale) (meetSetupSession, string) {
	in, out, done := applyBriefScopeInput(sess.config.Brief, content)
	sess.config.Brief.InScope = in
	sess.config.Brief.OutOfScope = out
	sess.config.Brief.DoneCriteria = done
	if sess.templateLocksMeeting {
		ensureMeetingDefaults(&sess.config)
		sess.step = setupStepCustomConfirm
		return sess, formatTemplateMeetConfirmPrompt(loc, sess.config)
	}
	sess.step = setupStepPresetMenu
	return sess, r.promptPresetMenu(loc, sess.config)
}

func applyBriefGoalInput(current, content string) string {
	if isBriefSkipToken(content) {
		return ""
	}
	if isBriefConfirmToken(content) {
		return strings.TrimSpace(current)
	}
	return strings.TrimSpace(content)
}

func applyBriefAgendaInput(current []string, content string) []string {
	if isBriefSkipToken(content) {
		return nil
	}
	if isBriefConfirmToken(content) {
		if len(current) == 0 {
			return nil
		}
		return append([]string(nil), current...)
	}
	return parseAgendaLines(content)
}

func applyBriefScopeInput(current meetBrief, content string) (inScope, outScope, done string) {
	if isBriefSkipToken(content) {
		return "", "", ""
	}
	if isBriefConfirmToken(content) {
		return strings.TrimSpace(current.InScope),
			strings.TrimSpace(current.OutOfScope),
			strings.TrimSpace(current.DoneCriteria)
	}
	return parseBriefScope(content)
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

func formatAskBriefGoalPrompt(loc Locale, topic string, b meetBrief, templateFlow bool) string {
	draft := strings.TrimSpace(b.Goal)
	hasDraft := draft != ""
	var draftBlock string
	if hasDraft {
		if loc == LocaleZH {
			draftBlock = "🎯 **模板目标：**\n" + draft + "\n\n"
		} else {
			draftBlock = "🎯 **Template goal:**\n" + draft + "\n\n"
		}
	} else if templateFlow {
		if loc == LocaleZH {
			draftBlock = "🎯 模板未预填会议目标，可填写或确认留空进入议程。\n\n"
		} else {
			draftBlock = "🎯 Template has no meeting goal; fill in or confirm to continue with an empty goal.\n\n"
		}
	}
	if loc == LocaleZH {
		return fmt.Sprintf(`📋 **会议简报 · 1/3 目标**

📌 主题：%s

%s这场会要**交付什么**？（例如：输出方案草案 + 待决清单）
%s
%s`, topic, draftBlock, briefStepActionHint(loc, hasDraft, templateFlow), briefStepGoalTopicNote(loc))
	}
	return fmt.Sprintf(`📋 **Meeting brief · 1/3 goal**

📌 Topic: %s

%sWhat should this meeting deliver?
%s
%s`, topic, draftBlock, briefStepActionHint(loc, hasDraft, templateFlow), briefStepGoalTopicNote(loc))
}

func formatAskBriefAgendaPrompt(loc Locale, b meetBrief, templateFlow bool) string {
	hasDraft := len(b.AgendaTitles) > 0
	var draftBlock string
	if hasDraft {
		if loc == LocaleZH {
			draftBlock = "📑 **模板讨论议题：**\n" + formatBriefAgendaLines(loc, b.AgendaTitles) + "\n\n"
		} else {
			draftBlock = "📑 **Template topics:**\n" + formatBriefAgendaLines(loc, b.AgendaTitles) + "\n\n"
		}
	}
	if loc == LocaleZH {
		return fmt.Sprintf(`📋 **会议简报 · 2/3 讨论议题**

%s请列出本场要**讨论覆盖的子议题**（每行一条，或 1、2、3 编号）。

说明：此处是**要聊的具体议题**，与 MEETING.md 里的「会议流程」（Pre-meeting → 研讨轮次 → 合成）不是同一概念。

%s`, draftBlock, briefStepActionHint(loc, hasDraft, templateFlow))
	}
	return fmt.Sprintf(`📋 **Meeting brief · 2/3 discussion topics**

%sList sub-topics to cover (one per line). These are **what to discuss**, not the meeting run-of-show in MEETING.md.

%s`, draftBlock, briefStepActionHint(loc, hasDraft, templateFlow))
}

func formatAskBriefScopePrompt(loc Locale, b meetBrief, templateFlow bool) string {
	hasDraft := briefScopeHasDraft(b)
	var draftBlock string
	if hasDraft {
		if loc == LocaleZH {
			draftBlock = "📌 **模板边界：**\n" + formatBriefScopeDraft(loc, b) + "\n\n"
		} else {
			draftBlock = "📌 **Template boundaries:**\n" + formatBriefScopeDraft(loc, b) + "\n\n"
		}
	}
	if loc == LocaleZH {
		return fmt.Sprintf(`📋 **会议简报 · 3/3 边界与完成标准**

%s可选；约束讨论深度、排除项与收工条件。一行或多行，例如：
讨论范围：概念层与取舍方向
不在范围：实施排期、详细成本表
完成标准：每个讨论议题至少 1 条结论或待决

%s`, draftBlock, briefStepActionHint(loc, hasDraft, templateFlow))
	}
	return fmt.Sprintf(`📋 **Meeting brief · 3/3 boundaries & done criteria**

%sOptional — set depth limits, exclusions, and when to wrap up.

%s`, draftBlock, briefStepActionHint(loc, hasDraft, templateFlow))
}

func briefScopeHasDraft(b meetBrief) bool {
	return strings.TrimSpace(b.InScope) != "" ||
		strings.TrimSpace(b.OutOfScope) != "" ||
		strings.TrimSpace(b.DoneCriteria) != ""
}

func formatBriefAgendaLines(loc Locale, titles []string) string {
	var lines []string
	for i, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		if loc == LocaleZH {
			lines = append(lines, fmt.Sprintf("%d）%s", i+1, title))
		} else {
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, title))
		}
	}
	return strings.Join(lines, "\n")
}

func formatBriefScopeDraft(loc Locale, b meetBrief) string {
	var lines []string
	if in := strings.TrimSpace(b.InScope); in != "" {
		if loc == LocaleZH {
			lines = append(lines, "- 讨论范围："+in)
		} else {
			lines = append(lines, "- In scope: "+in)
		}
	}
	if out := strings.TrimSpace(b.OutOfScope); out != "" {
		if loc == LocaleZH {
			lines = append(lines, "- 不在范围："+out)
		} else {
			lines = append(lines, "- Out of scope: "+out)
		}
	}
	if done := strings.TrimSpace(b.DoneCriteria); done != "" {
		if loc == LocaleZH {
			lines = append(lines, "- 完成标准："+done)
		} else {
			lines = append(lines, "- Done when: "+done)
		}
	}
	return strings.Join(lines, "\n")
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
		// Blank line breaks Discord list continuation from the goal bullet above.
		lines = append(lines, "")
		if loc == LocaleZH {
			lines = append(lines, "📑 **讨论议题**：")
		} else {
			lines = append(lines, "📑 **Topics**：")
		}
		for i, title := range b.AgendaTitles {
			// Plain 「N）」 lines — Discord eats markdown list numbers (- 1. …) in display/copy.
			if loc == LocaleZH {
				lines = append(lines, fmt.Sprintf("%d）%s", i+1, title))
			} else {
				lines = append(lines, fmt.Sprintf("%d. %s", i+1, title))
			}
		}
		lines = append(lines, "")
	}
	if in := strings.TrimSpace(b.InScope); in != "" {
		if loc == LocaleZH {
			lines = append(lines, "- ✅ 讨论范围："+in)
		} else {
			lines = append(lines, "- ✅ In scope: "+in)
		}
	}
	if out := strings.TrimSpace(b.OutOfScope); out != "" {
		if loc == LocaleZH {
			lines = append(lines, "- ⛔ 不在范围："+out)
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
