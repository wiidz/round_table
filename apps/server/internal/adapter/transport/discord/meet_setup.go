package discord

import (
	"fmt"
	"strings"
	"sync"

	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/platform/config"
)

// meetLaunchConfig is the resolved configuration for starting a meeting.
type meetLaunchConfig struct {
	Topic                    string
	Brief                    meetBrief
	Mode                     string
	MaxRounds                int
	MinRoundsBeforeSynthesis int
	Confirmation             string
	FreeDialogueQuestions    int
	ParticipantsSummary      string
	ParticipantIDs           []string // empty = full roster
}

type setupStep int

const (
	setupStepAskTopic setupStep = iota
	setupStepPickParticipants
	setupStepBriefGoal
	setupStepBriefAgenda
	setupStepBriefScope
	setupStepPresetMenu
	setupStepCustomMode
	setupStepCustomRounds
	setupStepCustomConfirmation
	setupStepCustomFree
	setupStepCustomConfirm
	setupStepPickTemplate
)

type meetSetupSession struct {
	channelID            string
	authorID             string
	config               meetLaunchConfig
	step                 setupStep
	briefTemplateID      string
	templateLocksMeeting bool
}

type meetSetupSessions struct {
	mu        sync.Mutex
	byChannel map[string]meetSetupSession
}

func (s *meetSetupSessions) put(channelID string, sess meetSetupSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byChannel == nil {
		s.byChannel = make(map[string]meetSetupSession)
	}
	s.byChannel[channelID] = sess
}

func (s *meetSetupSessions) get(channelID string) (meetSetupSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.byChannel[channelID]
	return sess, ok
}

func (s *meetSetupSessions) clear(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byChannel, channelID)
}

func (s *meetSetupSessions) pending(channelID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.byChannel[channelID]
	return ok
}

func (r *MeetRunner) meetPresets(loc Locale) []meetPreset {
	return buildMeetPresets(r.activeCfg().Meeting.MeetPresets, loc)
}

func (r *MeetRunner) defaultLaunchConfig(topic, modeOverride string) meetLaunchConfig {
	presets := r.activeCfg().Meeting.MeetPresets
	if len(presets) == 0 {
		presets = config.DefaultMeetPresets(r.activeCfg())
	}
	for _, p := range presets {
		if p.ID == "1" {
			cfg := launchConfigFromPreset(p, topic)
			if modeOverride != "" {
				cfg.Mode = modeOverride
			}
			cfg.ParticipantsSummary = summarizeParticipants(r.dc().MeetParticipants)
			return cfg
		}
	}
	cfg := launchConfigFromPreset(presets[0], topic)
	if modeOverride != "" {
		cfg.Mode = modeOverride
	}
	cfg.ParticipantsSummary = summarizeParticipants(r.dc().MeetParticipants)
	return cfg
}

func (r *MeetRunner) meetCasts() []config.MeetCastConfig {
	if casts := r.activeCfg().Meeting.MeetCasts; len(casts) > 0 {
		return casts
	}
	return nil
}

func summarizeParticipants(raw string) string {
	var parts []string
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		first := strings.Index(item, ":")
		if first <= 0 {
			parts = append(parts, item)
			continue
		}
		id := item[:first]
		rest := item[first+1:]
		last := strings.LastIndex(rest, ":")
		if last <= 0 {
			parts = append(parts, id+"·"+rest)
			continue
		}
		parts = append(parts, id+"·"+rest[:last])
	}
	return strings.Join(parts, ", ")
}

func formatConfigShort(cfg meetLaunchConfig, loc Locale) string {
	free := freeDialogueLabel(cfg.FreeDialogueQuestions, loc)
	if loc == LocaleZH {
		return fmt.Sprintf("%s · %d轮 · 确认%s · 自由对话%s",
			meetingModeLabel(cfg.Mode, loc), cfg.MaxRounds,
			confirmationModeLabel(cfg.Confirmation, loc), free)
	}
	return fmt.Sprintf("%s · %d rounds · confirm %s · free dialogue %s",
		cfg.Mode, cfg.MaxRounds, cfg.Confirmation, free)
}

func freeDialogueLabel(n int, loc Locale) string {
	if n <= 0 {
		if loc == LocaleZH {
			return "关"
		}
		return "off"
	}
	if loc == LocaleZH {
		return fmt.Sprintf("%d轮/人", n)
	}
	return fmt.Sprintf("%d/participant", n)
}

func formatAskTopicPrompt(loc Locale) string {
	if loc == LocaleZH {
		return `🎙️ **主持人** — 请输入会议主题

直接发送主题文字即可（无需前缀）。
取消：发送 **取消会议**`
	}
	return `🎙️ **Moderator** — enter the meeting topic

Send the topic as plain text (no prefix).
Cancel: send **取消会议**`
}

func formatCustomStepPrompt(loc Locale, step setupStep) string {
	if loc == LocaleZH {
		switch step {
		case setupStepCustomMode:
			return `🎙️ **自定义 · 1/4 模式**
**1** — 研讨（出方案草案）
**J** — 裁决（投票共识）
**0** — 返回预设菜单`
		case setupStepCustomRounds:
			return `🎙️ **自定义 · 2/4 轮次**
**1** — 1 轮（闪电）
**2** — 2 轮
**3** — 3 轮
**4** — 5 轮
**0** — 返回预设菜单`
		case setupStepCustomConfirmation:
			return `🎙️ **自定义 · 3/4 Principal 确认**
**1** — 跳过（合成后直接结束）
**2** — 需要（合成后等你审阅）
**0** — 返回预设菜单`
		case setupStepCustomFree:
			return `🎙️ **自定义 · 4/4 自由对话**
**1** — 关闭
**2** — 第 1 轮后每人 1 轮提问
**0** — 返回预设菜单`
		case setupStepCustomConfirm:
			return `🎙️ **请确认配置**`
		}
	}
	switch step {
	case setupStepCustomMode:
		return `🎙️ **Custom · 1/4 mode**
**1** — Deliberation (design draft)
**J** — Decision (vote consensus)
**0** — Back to preset menu`
	case setupStepCustomRounds:
		return `🎙️ **Custom · 2/4 rounds**
**1** — 1 round (flash)
**2** — 2 rounds
**3** — 3 rounds
**4** — 5 rounds
**0** — Back to preset menu`
	case setupStepCustomConfirmation:
		return `🎙️ **Custom · 3/4 confirmation**
**1** — Skip (finish after synthesis)
**2** — Required (Principal review)
**0** — Back to preset menu`
	case setupStepCustomFree:
		return `🎙️ **Custom · 4/4 free dialogue**
**1** — Off
**2** — 1 question/participant after round 1
**0** — Back to preset menu`
	case setupStepCustomConfirm:
		return `🎙️ **Confirm setup**`
	}
	return ""
}

func formatCustomConfirmPrompt(loc Locale, cfg meetLaunchConfig) string {
	head := formatCustomStepPrompt(loc, setupStepCustomConfirm)
	if loc == LocaleZH {
		return fmt.Sprintf(`%s

📌 主题：%s
- 🎯 %s
- 🔄 %d 轮（最少 %d 轮再合成）
- ✅ 确认：%s
- 💬 自由对话：%s
- 👥 %s

**1** — 开始
**0** — 重新选择`, head, cfg.Topic,
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

**1** — Start
**0** — Choose again`, head, cfg.Topic,
		cfg.Mode, cfg.MaxRounds, cfg.MinRoundsBeforeSynthesis,
		cfg.Confirmation,
		freeDialogueLabel(cfg.FreeDialogueQuestions, LocaleEN),
		cfg.ParticipantsSummary)
}

type setupHandleResult struct {
	reply  string
	launch bool
	config meetLaunchConfig
	step   setupStep
}

func normalizeSetupChoice(content string) string {
	s := strings.TrimSpace(content)
	if s == "" {
		return ""
	}
	s = normalizeASCIIForms(s)
	lower := strings.ToLower(s)
	if lower == "开始" || lower == "默认" || lower == "start" || lower == "default" || lower == "ok" {
		return "1"
	}
	fields := strings.Fields(s)
	if len(fields) == 1 {
		c := strings.ToUpper(fields[0])
		if c == "J" {
			return "J"
		}
		return fields[0]
	}
	return s
}

func handleSetupStep(sess meetSetupSession, choice string, loc Locale, prefix string, all []meetPreset) (setupHandleResult, error) {
	if sess.step == setupStepPresetMenu {
		choice = normalizePresetChoice(choice)
	} else {
		choice = normalizeSetupChoice(choice)
	}
	if choice == "" {
		return setupHandleResult{}, errSetupReplyEmpty
	}

	switch sess.step {
	case setupStepPresetMenu:
		return handlePresetMenu(sess, choice, loc, prefix, all)
	case setupStepCustomMode:
		return handleCustomMode(sess, choice, loc, prefix, all)
	case setupStepCustomRounds:
		return handleCustomRounds(sess, choice, loc, prefix, all)
	case setupStepCustomConfirmation:
		return handleCustomConfirmation(sess, choice, loc, prefix, all)
	case setupStepCustomFree:
		return handleCustomFree(sess, choice, loc, prefix, all)
	case setupStepCustomConfirm:
		return handleCustomConfirm(sess, choice, loc, prefix, all)
	default:
		return setupHandleResult{}, errSetupReplyUnrecognized
	}
}

func handlePresetMenu(sess meetSetupSession, choice string, loc Locale, prefix string, all []meetPreset) (setupHandleResult, error) {
	topic := sess.config.Topic
	participants := sess.config.ParticipantsSummary

	if choice == "0" {
		sess.step = setupStepCustomMode
		return setupHandleResult{
			reply:  formatCustomStepPrompt(loc, setupStepCustomMode),
			step:   setupStepCustomMode,
			config: sess.config,
		}, nil
	}

	preset, ok := lookupPreset(choice, all)
	if !ok {
		return setupHandleResult{}, errSetupInvalidChoice
	}

	cfg := mergePresetLaunchConfig(sess.config, preset.Make(topic))
	cfg.ParticipantsSummary = participants
	if err := validateLaunchConfig(cfg); err != nil {
		return setupHandleResult{}, err
	}
	return setupHandleResult{launch: true, config: cfg}, nil
}

func handleCustomMode(sess meetSetupSession, choice string, loc Locale, prefix string, all []meetPreset) (setupHandleResult, error) {
	switch choice {
	case "0":
		sess.step = setupStepPresetMenu
		return setupHandleResult{
			reply:  formatModeratorSetupWithBrief(loc, prefix, all, sess.config),
			step:   setupStepPresetMenu,
			config: sess.config,
		}, nil
	case "1":
		sess.config.Mode = meeting.MeetingModeDeliberation
	case "J":
		sess.config.Mode = meeting.MeetingModeDecision
	default:
		return setupHandleResult{}, errSetupInvalidChoice
	}
	sess.step = setupStepCustomRounds
	return setupHandleResult{
		reply:  formatCustomStepPrompt(loc, setupStepCustomRounds),
		step:   setupStepCustomRounds,
		config: sess.config,
	}, nil
}

func handleCustomRounds(sess meetSetupSession, choice string, loc Locale, prefix string, all []meetPreset) (setupHandleResult, error) {
	var rounds int
	switch choice {
	case "0":
		sess.step = setupStepPresetMenu
		return setupHandleResult{
			reply:  formatModeratorSetupWithBrief(loc, prefix, all, sess.config),
			step:   setupStepPresetMenu,
			config: sess.config,
		}, nil
	case "1":
		rounds = 1
	case "2":
		rounds = 2
	case "3":
		rounds = 3
	case "4":
		rounds = 5
	default:
		return setupHandleResult{}, errSetupInvalidChoice
	}
	sess.config.MaxRounds = rounds
	sess.config.MinRoundsBeforeSynthesis = 2
	if rounds < 2 {
		sess.config.MinRoundsBeforeSynthesis = rounds
	}
	sess.step = setupStepCustomConfirmation
	return setupHandleResult{
		reply:  formatCustomStepPrompt(loc, setupStepCustomConfirmation),
		step:   setupStepCustomConfirmation,
		config: sess.config,
	}, nil
}

func handleCustomConfirmation(sess meetSetupSession, choice string, loc Locale, prefix string, all []meetPreset) (setupHandleResult, error) {
	switch choice {
	case "0":
		sess.step = setupStepPresetMenu
		return setupHandleResult{
			reply:  formatModeratorSetupWithBrief(loc, prefix, all, sess.config),
			step:   setupStepPresetMenu,
			config: sess.config,
		}, nil
	case "1":
		sess.config.Confirmation = meeting.ConfirmationModeSkip
	case "2":
		sess.config.Confirmation = meeting.ConfirmationModeRequired
	default:
		return setupHandleResult{}, errSetupInvalidChoice
	}
	sess.step = setupStepCustomFree
	return setupHandleResult{
		reply:  formatCustomStepPrompt(loc, setupStepCustomFree),
		step:   setupStepCustomFree,
		config: sess.config,
	}, nil
}

func handleCustomFree(sess meetSetupSession, choice string, loc Locale, prefix string, all []meetPreset) (setupHandleResult, error) {
	switch choice {
	case "0":
		sess.step = setupStepPresetMenu
		return setupHandleResult{
			reply:  formatModeratorSetupWithBrief(loc, prefix, all, sess.config),
			step:   setupStepPresetMenu,
			config: sess.config,
		}, nil
	case "1":
		sess.config.FreeDialogueQuestions = 0
	case "2":
		sess.config.FreeDialogueQuestions = 1
	default:
		return setupHandleResult{}, errSetupInvalidChoice
	}
	if err := validateLaunchConfig(sess.config); err != nil {
		return setupHandleResult{}, err
	}
	sess.step = setupStepCustomConfirm
	return setupHandleResult{
		reply:  formatCustomConfirmPrompt(loc, sess.config),
		step:   setupStepCustomConfirm,
		config: sess.config,
	}, nil
}

func handleCustomConfirm(sess meetSetupSession, choice string, loc Locale, prefix string, all []meetPreset) (setupHandleResult, error) {
	switch choice {
	case "0":
		sess.step = setupStepPresetMenu
		return setupHandleResult{
			reply:  formatModeratorSetupWithBrief(loc, prefix, all, sess.config),
			step:   setupStepPresetMenu,
			config: sess.config,
		}, nil
	case "1":
		return setupHandleResult{launch: true, config: sess.config}, nil
	default:
		return setupHandleResult{}, errSetupInvalidChoice
	}
}

func confirmationModeLabel(mode string, loc Locale) string {
	if loc != LocaleZH {
		return mode
	}
	switch mode {
	case meeting.ConfirmationModeSkip:
		return "跳过"
	case meeting.ConfirmationModeRequired:
		return "需要"
	default:
		return mode
	}
}

func validateLaunchConfig(cfg meetLaunchConfig) error {
	switch cfg.Mode {
	case meeting.MeetingModeDecision, meeting.MeetingModeDeliberation:
	default:
		return errSetupInvalidMode
	}
	switch cfg.Confirmation {
	case meeting.ConfirmationModeSkip, meeting.ConfirmationModeRequired:
	default:
		return errSetupInvalidConfirmation
	}
	if cfg.MaxRounds <= 0 {
		return errSetupInvalidRounds
	}
	if cfg.MinRoundsBeforeSynthesis <= 0 {
		return errSetupInvalidMinRounds
	}
	if cfg.FreeDialogueQuestions < 0 {
		return errSetupInvalidFree
	}
	return nil
}

const participantSummarySep = "\u00b7" // id·display in ParticipantsSummary

func formatLaunchParticipantsSummary(summary string, loc Locale) string {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		if loc == LocaleZH {
			return "全员"
		}
		return "full roster"
	}
	var names []string
	for _, part := range strings.Split(summary, ",") {
		part = strings.TrimSpace(part)
		if _, display, ok := strings.Cut(part, participantSummarySep); ok && strings.TrimSpace(display) != "" {
			names = append(names, strings.TrimSpace(display))
			continue
		}
		names = append(names, part)
	}
	if loc == LocaleZH {
		return strings.Join(names, "、")
	}
	return strings.Join(names, ", ")
}

func formatMeetLaunchAck(loc Locale, meetingID string, cfg meetLaunchConfig, principalName string, maxConfirmationCycles int) string {
	participants := formatLaunchParticipantsSummary(cfg.ParticipantsSummary, loc)
	briefBlock := formatBriefLaunchBlock(loc, cfg)
	if briefBlock != "" {
		briefBlock = "\n\n" + briefBlock
	}
	if loc == LocaleZH {
		confirmLine := confirmationModeLabel(cfg.Confirmation, loc)
		if cfg.Confirmation == meeting.ConfirmationModeRequired && maxConfirmationCycles > 0 {
			confirmLine = fmt.Sprintf("%s（最多 %d 轮确认）", confirmLine, maxConfirmationCycles)
		}
		return fmt.Sprintf(`🚀 **会议已启动**
- 🆔 `+"`%s`"+`
- 📌 主题：%s
- 👥 参会：%s
- 🎯 模式：%s
- 🔄 轮次上限：%d · 最少 %d 轮再合成
- ✅ 确认：%s · 💬 自由对话：%s
- 👤 Principal：%s%s

进度将推送到本频道。`, meetingID, cfg.Topic, participants, meetingModeLabel(cfg.Mode, loc),
			cfg.MaxRounds, cfg.MinRoundsBeforeSynthesis,
			confirmLine,
			freeDialogueLabel(cfg.FreeDialogueQuestions, loc), principalName, briefBlock)
	}
	confirmLine := cfg.Confirmation
	if cfg.Confirmation == meeting.ConfirmationModeRequired && maxConfirmationCycles > 0 {
		confirmLine = fmt.Sprintf("%s (max %d confirmation cycles)", cfg.Confirmation, maxConfirmationCycles)
	}
	return fmt.Sprintf(`🚀 **Meeting started**
- 🆔 `+"`%s`"+`
- 📌 Topic: %s
- 👥 Participants: %s
- 🎯 Mode: %s
- 🔄 Max rounds: %d · min before synthesis: %d
- ✅ Confirmation: %s · 💬 Free dialogue: %s
- 👤 Principal: %s%s

Progress will post here.`, meetingID, cfg.Topic, participants, cfg.Mode,
		cfg.MaxRounds, cfg.MinRoundsBeforeSynthesis,
		confirmLine,
		freeDialogueLabel(cfg.FreeDialogueQuestions, LocaleEN), principalName, briefBlock)
}
