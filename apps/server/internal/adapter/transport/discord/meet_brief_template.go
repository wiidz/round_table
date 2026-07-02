package discord

import (
	"fmt"
	"strconv"
	"strings"

	"round_table/apps/server/internal/adapter/brief"
	"round_table/apps/server/internal/domain/meeting"
)

func (r *MeetRunner) briefStore() brief.Port {
	if r == nil {
		return nil
	}
	return r.Briefs
}

func (r *MeetRunner) listBriefTemplates() ([]brief.TemplateIndex, error) {
	store := r.briefStore()
	if store == nil {
		return nil, nil
	}
	return store.ListTemplates()
}

func (r *MeetRunner) readBriefTemplate(id string) (brief.TemplateDetail, error) {
	store := r.briefStore()
	if store == nil {
		return brief.TemplateDetail{}, brief.ErrNotFound
	}
	return store.ReadTemplate(strings.TrimSpace(id))
}

func launchDraftToMeetConfig(base meetLaunchConfig, draft brief.LaunchDraft) meetLaunchConfig {
	out := base

	if topic := strings.TrimSpace(draft.Topic); topic != "" && strings.TrimSpace(out.Topic) == "" {
		out.Topic = topic
	}

	if goal := strings.TrimSpace(draft.Brief.Goal); goal != "" {
		out.Brief.Goal = goal
	}
	if len(draft.Brief.Agenda) > 0 {
		out.Brief.AgendaTitles = append([]string(nil), draft.Brief.Agenda...)
	}
	if v := strings.TrimSpace(draft.Brief.InScope); v != "" {
		out.Brief.InScope = v
	}
	if v := strings.TrimSpace(draft.Brief.OutOfScope); v != "" {
		out.Brief.OutOfScope = v
	}
	if v := strings.TrimSpace(draft.Brief.DoneCriteria); v != "" {
		out.Brief.DoneCriteria = v
	}

	applyMeetingDefaults(&out, draft.Meeting)
	return out
}

func applyMeetingDefaults(cfg *meetLaunchConfig, m brief.MeetingDefaults) {
	if mode := strings.TrimSpace(m.Mode); mode != "" {
		cfg.Mode = mode
	}
	if m.MaxRounds > 0 {
		cfg.MaxRounds = m.MaxRounds
	}
	if m.MinRoundsBeforeSynthesis > 0 {
		cfg.MinRoundsBeforeSynthesis = m.MinRoundsBeforeSynthesis
	} else if m.MaxRounds > 0 && cfg.MinRoundsBeforeSynthesis > cfg.MaxRounds {
		cfg.MinRoundsBeforeSynthesis = cfg.MaxRounds
	}
	if cm := strings.TrimSpace(m.ConfirmationMode); cm != "" {
		cfg.Confirmation = cm
	}
	if m.FreeDialogueMaxQuestions > 0 {
		cfg.FreeDialogueQuestions = m.FreeDialogueMaxQuestions
	}
	if len(m.ParticipantIDs) > 0 {
		cfg.ParticipantIDs = append([]string(nil), m.ParticipantIDs...)
	}
}

func launchDraftLocksMeeting(draft brief.LaunchDraft) bool {
	return strings.TrimSpace(draft.Meeting.Mode) != ""
}

func (r *MeetRunner) applyBriefTemplate(cfg meetLaunchConfig, templateID string) (meetLaunchConfig, brief.LaunchDraft, bool, error) {
	detail, err := r.readBriefTemplate(templateID)
	if err != nil {
		return cfg, brief.LaunchDraft{}, false, err
	}
	out := launchDraftToMeetConfig(cfg, detail.Launch)
	if len(out.ParticipantIDs) > 0 {
		out.ParticipantsSummary = summarizeParticipantIDs(r.dc().MeetParticipants, out.ParticipantIDs)
	}
	return out, detail.Launch, launchDraftLocksMeeting(detail.Launch), nil
}

func nextStepAfterTemplateApply(cfg meetLaunchConfig, templateLocksMeeting bool, loc Locale, r *MeetRunner) (setupStep, string) {
	if strings.TrimSpace(cfg.Topic) == "" {
		return setupStepAskTopic, formatAskTopicPrompt(loc)
	}
	if len(cfg.ParticipantIDs) == 0 && strings.TrimSpace(cfg.ParticipantsSummary) == "" {
		return setupStepPickParticipants, formatPickParticipantsPrompt(loc, r.dc().MeetParticipants, r.meetCasts())
	}
	return setupStepBriefGoal, formatAskBriefGoalPrompt(loc, cfg.Topic, cfg.Brief, true)
}

func (r *MeetRunner) shouldOfferTemplatePick() bool {
	list, err := r.listBriefTemplates()
	return err == nil && len(list) > 0
}

func formatPickTemplatePrompt(loc Locale, templates []brief.TemplateIndex) string {
	var b strings.Builder
	if loc == LocaleZH {
		b.WriteString("📋 **简报模板** — 开会前可选预填\n\n")
	} else {
		b.WriteString("📋 **Brief templates** — optional prefill before meeting\n\n")
	}
	for i, item := range templates {
		n := i + 1
		desc := strings.TrimSpace(item.Description)
		if desc == "" {
			if loc == LocaleZH {
				desc = "无说明"
			} else {
				desc = "No description"
			}
		}
		fmt.Fprintf(&b, "**%d** %s · `%s`\n    └ %s\n\n", n, item.Title, item.ID, desc)
	}
	b.WriteString(briefTemplatePickActionHint(loc))
	return strings.TrimRight(b.String(), "\n")
}

func resolveTemplateChoice(content string, templates []brief.TemplateIndex) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" || isBriefSkipToken(content) {
		return "", nil
	}
	if n, err := strconv.Atoi(content); err == nil && n > 0 && n <= len(templates) {
		return templates[n-1].ID, nil
	}
	lower := strings.ToLower(content)
	for _, item := range templates {
		if strings.EqualFold(item.ID, content) || strings.EqualFold(item.ID, lower) {
			return item.ID, nil
		}
	}
	for _, item := range templates {
		if strings.Contains(strings.ToLower(item.Title), lower) {
			return item.ID, nil
		}
	}
	return "", errSetupInvalidTemplate
}

func formatTemplateMeetConfirmPrompt(loc Locale, cfg meetLaunchConfig) string {
	head := formatCustomStepPrompt(loc, setupStepCustomConfirm)
	if loc == LocaleZH {
		return fmt.Sprintf(`%s

📌 主题：%s
- 🎯 %s
- 🔄 %d 轮（最少 %d 轮再合成）
- ✅ 确认：%s
- 💬 自由对话：%s
- 👥 %s

（会议配置来自简报模板，可在确认前返回修改）

**1** — 开始
**0** — 重新选择预设`, head, cfg.Topic,
			meetingModeLabel(cfg.Mode, loc), cfg.MaxRounds, cfg.MinRoundsBeforeSynthesis,
			confirmationModeLabel(cfg.Confirmation, loc),
			freeDialogueLabel(cfg.FreeDialogueQuestions, loc),
			cfg.ParticipantsSummary)
	}
	return fmt.Sprintf(`%s

📌 Topic: %s
- 🎯 %s
- 🔄 %d rounds (min %d before synthesis)
- ✅ Confirmation: %s
- 💬 Free dialogue: %s
- 👥 %s

(Meeting config from brief template)

**1** — Start
**0** — Back to presets`, head, cfg.Topic,
		cfg.Mode, cfg.MaxRounds, cfg.MinRoundsBeforeSynthesis,
		cfg.Confirmation,
		freeDialogueLabel(cfg.FreeDialogueQuestions, LocaleEN),
		cfg.ParticipantsSummary)
}

func formatBriefTemplateApplied(loc Locale, title, id string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("✅ 已加载简报模板：**%s**（`%s`）", title, id)
	}
	return fmt.Sprintf("✅ Loaded brief template: **%s** (`%s`)", title, id)
}

func meetTemplateNotFoundText(loc Locale, id string, err error) string {
	if loc == LocaleZH {
		return fmt.Sprintf("❌ 未找到简报模板 `%s`：%v", id, err)
	}
	return fmt.Sprintf("❌ Brief template `%s` not found: %v", id, err)
}

func meetTemplateUnavailableText(loc Locale) string {
	if loc == LocaleZH {
		return "❌ 简报模板服务不可用。"
	}
	return "❌ Brief template store is unavailable."
}

func ensureMeetingDefaults(cfg *meetLaunchConfig) {
	if strings.TrimSpace(cfg.Mode) == "" {
		cfg.Mode = meeting.MeetingModeDecision
	}
	if cfg.MaxRounds <= 0 {
		cfg.MaxRounds = 3
	}
	if cfg.MinRoundsBeforeSynthesis <= 0 {
		cfg.MinRoundsBeforeSynthesis = 2
		if cfg.MaxRounds < cfg.MinRoundsBeforeSynthesis {
			cfg.MinRoundsBeforeSynthesis = cfg.MaxRounds
		}
	}
	if strings.TrimSpace(cfg.Confirmation) == "" {
		cfg.Confirmation = meeting.ConfirmationModeRequired
	}
}
