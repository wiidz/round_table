package discord

import (
	"fmt"
	"strings"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
)

// ChannelInputPhase is the Discord channel's expected Principal input mode.
type ChannelInputPhase string

const (
	InputPhaseIdle              ChannelInputPhase = "idle"
	InputPhaseSetupTopic        ChannelInputPhase = "setup_topic"
	InputPhaseSetupParticipants ChannelInputPhase = "setup_participants"
	InputPhaseSetupBrief        ChannelInputPhase = "setup_brief"
	InputPhaseSetupMenu         ChannelInputPhase = "setup_menu"
	InputPhaseSetupCustom       ChannelInputPhase = "setup_custom"
	InputPhaseMeetingRunning    ChannelInputPhase = "meeting_running"
	InputPhaseMeetingPaused     ChannelInputPhase = "meeting_paused"
	InputPhaseMeetingFreeDialogue ChannelInputPhase = "meeting_free_dialogue"
	InputPhaseMeetingConfirmation ChannelInputPhase = "meeting_confirmation"
	InputPhasePostMeeting       ChannelInputPhase = "post_meeting"
	InputPhaseExpertSetup       ChannelInputPhase = "expert_setup"
	InputPhaseReceptionConfirm  ChannelInputPhase = "reception_confirm"
	InputPhaseReceptionClarify  ChannelInputPhase = "reception_clarify"
)

func isInputStatusTrigger(content string) bool {
	s := strings.TrimSpace(content)
	if s == "" {
		return false
	}
	norm := normalizeASCIIForms(s)
	lower := strings.ToLower(norm)
	return matchExact(lower, "会议状态", "状态", "status", "meet status")
}

// InputPhase returns the current input phase for a Discord channel.
func (r *MeetRunner) InputPhase(channelID string) ChannelInputPhase {
	if r.setups.pending(channelID) {
		sess, _ := r.setups.get(channelID)
		switch sess.step {
		case setupStepPickTemplate:
			return InputPhaseSetupBrief
		case setupStepAskTopic:
			return InputPhaseSetupTopic
		case setupStepPickParticipants:
			return InputPhaseSetupParticipants
		case setupStepBriefGoal, setupStepBriefAgenda, setupStepBriefScope:
			return InputPhaseSetupBrief
		case setupStepPresetMenu:
			return InputPhaseSetupMenu
		default:
			return InputPhaseSetupCustom
		}
	}
	if r.Principal != nil {
		if _, _, ok := r.Principal.PendingConfirmation(channelID); ok {
			return InputPhaseMeetingConfirmation
		}
		if r.Principal.PendingPaused(channelID) {
			return InputPhaseMeetingPaused
		}
		if r.Principal.InFreeDialogue(channelID) {
			return InputPhaseMeetingFreeDialogue
		}
	}
	if _, active := r.sessions.active(channelID); active {
		return InputPhaseMeetingRunning
	}
	if _, ok := r.sessions.last(channelID); ok {
		return InputPhasePostMeeting
	}
	return InputPhaseIdle
}

// HandleInputStatus returns the current input phase hint for Principal.
func (r *MeetRunner) HandleInputStatus(msg transport.Inbound) (string, error) {
	if !isInputStatusTrigger(msg.Content) {
		return "", nil
	}
	loc := r.locale()
	phase := r.InputPhase(msg.ChannelID)
	return formatInputPhaseStatus(loc, phase, r.meetingIDForPhase(msg.ChannelID, phase)), nil
}

func (r *MeetRunner) meetingIDForPhase(channelID string, phase ChannelInputPhase) string {
	switch phase {
	case InputPhaseMeetingRunning, InputPhaseMeetingPaused,
		InputPhaseMeetingFreeDialogue, InputPhaseMeetingConfirmation:
		if id, ok := r.sessions.active(channelID); ok {
			return id
		}
	case InputPhasePostMeeting:
		if id, ok := r.sessions.last(channelID); ok {
			return id
		}
	}
	return ""
}

func formatInputPhaseStatus(loc Locale, phase ChannelInputPhase, meetingID string) string {
	return formatInputPhaseStatusForPlatform(loc, phase, meetingID, "")
}

func formatInputPhaseStatusForPlatform(loc Locale, phase ChannelInputPhase, meetingID, platform string) string {
	hint := inputPhaseHint(loc, phase)
	if platform == "web" && phase == InputPhaseIdle {
		hint = webIdleInputHint(loc)
	}
	if loc == LocaleZH {
		title := inputPhaseTitleZH(phase)
		if meetingID != "" {
			return fmt.Sprintf("📍 **当前输入态：%s**\n🆔 `%s`\n\n%s", title, meetingID, hint)
		}
		return fmt.Sprintf("📍 **当前输入态：%s**\n\n%s", title, hint)
	}
	title := inputPhaseTitleEN(phase)
	if meetingID != "" {
		return fmt.Sprintf("📍 **Input phase: %s**\n🆔 `%s`\n\n%s", title, meetingID, hint)
	}
	return fmt.Sprintf("📍 **Input phase: %s**\n\n%s", title, hint)
}

func inputPhaseTitleZH(phase ChannelInputPhase) string {
	switch phase {
	case InputPhaseSetupTopic:
		return "配置 · 等待主题"
	case InputPhaseSetupParticipants:
		return "配置 · 选择专家"
	case InputPhaseSetupBrief:
		return "配置 · 会议简报"
	case InputPhaseSetupMenu:
		return "配置 · 选择预设"
	case InputPhaseSetupCustom:
		return "配置 · 自定义向导"
	case InputPhaseMeetingRunning:
		return "会议进行中"
	case InputPhaseMeetingPaused:
		return "会议已暂停"
	case InputPhaseMeetingFreeDialogue:
		return "自由问答"
	case InputPhaseMeetingConfirmation:
		return "确认关"
	case InputPhasePostMeeting:
		return "会议已结束"
	case InputPhaseExpertSetup:
		return "专家 · 向导中"
	case InputPhaseReceptionConfirm:
		return "接待 · 等待确认"
	case InputPhaseReceptionClarify:
		return "接待 · 补充信息"
	default:
		return "空闲"
	}
}

func inputPhaseTitleEN(phase ChannelInputPhase) string {
	switch phase {
	case InputPhaseSetupTopic:
		return "Setup · awaiting topic"
	case InputPhaseSetupParticipants:
		return "Setup · choose participants"
	case InputPhaseSetupBrief:
		return "Setup · meeting brief"
	case InputPhaseSetupMenu:
		return "Setup · preset menu"
	case InputPhaseSetupCustom:
		return "Setup · custom wizard"
	case InputPhaseMeetingRunning:
		return "Meeting running"
	case InputPhaseMeetingPaused:
		return "Meeting paused"
	case InputPhaseMeetingFreeDialogue:
		return "Free dialogue"
	case InputPhaseMeetingConfirmation:
		return "Confirmation"
	case InputPhasePostMeeting:
		return "Meeting finished"
	case InputPhaseExpertSetup:
		return "Expert · wizard"
	case InputPhaseReceptionConfirm:
		return "Reception · confirm"
	case InputPhaseReceptionClarify:
		return "Reception · follow-up"
	default:
		return "Idle"
	}
}

func inputPhaseHint(loc Locale, phase ChannelInputPhase) string {
	if loc == LocaleZH {
		switch phase {
		case InputPhaseSetupTopic:
			return "请直接发送**会议主题**文字，或 **取消会议**。"
		case InputPhaseSetupParticipants:
			return "发送 **阵容编号**（如 C1）、**专家编号/名字**（如 1,2 或 designer,player），**0** 全员，或 **取消会议**。"
		case InputPhaseSetupBrief:
			return "按主持人上一条提示确认或修改 **简报**；**1** / **确认** 采用模板内容，**0** / **跳过** 留空，或直接发文字修改；**取消会议** 放弃。"
		case InputPhaseSetupMenu:
			return "发送预设编号 **1–6** / **J1–J5**，**0** 自定义，**取消会议** 放弃。"
		case InputPhaseSetupCustom:
			return "按主持人上一条提示回复数字/选项；**0** 返回上一级，**取消会议** 放弃。"
		case InputPhaseMeetingRunning:
			return "可发送：**暂停会议** · **终止会议** · **立即合成**（研讨）· **强制共识**（裁决）\n自由问答阶段另可：**提问 …**"
		case InputPhaseMeetingPaused:
			return "请发送 **恢复会议** 或 **终止会议**。"
		case InputPhaseMeetingFreeDialogue:
			return "可发送：**提问 [参与者] …** · 运行期干预指令（见上）\n发送 **会议状态** 随时查看本提示。"
		case InputPhaseMeetingConfirmation:
			return "请发送：**批准** / **驳回 …** / 逐项 **1: …  2: …**\n触顶时 **1/2/3** 三选一。"
		case InputPhasePostMeeting:
			return "可发送：**获取纪要** · **获取草案** · **获取待决** · **获取结论**\n或 **新会议** 开始下一场。"
		case InputPhaseExpertSetup:
			return "按主持人上一条提示回复；**取消专家** 放弃向导。"
		case InputPhaseReceptionConfirm:
			return "请发送 **1** 确认或 **0** 取消上一条操作。"
		case InputPhaseReceptionClarify:
			return "请按上一条提示补充信息；**取消专家** 或 **0** 放弃。"
		default:
			return "可发送：**新会议** / **!rt 专家 列表** / **!rt principal bind** / **!rt help**"
		}
	}
	switch phase {
	case InputPhaseSetupTopic:
		return "Send the **meeting topic**, or **取消会议** to cancel."
	case InputPhaseSetupParticipants:
		return "Send **cast id** (e.g. C1), **participant index/name** (e.g. 1,2), **0** for all, or **取消会议**."
	case InputPhaseSetupBrief:
		return "Follow the brief prompt: **1** / **confirm** to accept template content, **0** / **skip** to leave empty, or send text to edit; **取消会议** to cancel."
	case InputPhaseSetupMenu:
		return "Reply **1–6** / **J1–J5**, **0** for custom, **取消会议** to cancel."
	case InputPhaseSetupCustom:
		return "Follow the Moderator prompt; **0** back, **取消会议** to cancel."
	case InputPhaseMeetingRunning:
		return "**Pause** · **abort** · **force synthesis** / **force consensus**; **ask …** during free dialogue."
	case InputPhaseMeetingPaused:
		return "Send **resume** or **abort**."
	case InputPhaseMeetingFreeDialogue:
		return "**ask [participant] …** · running controls · **status** for this hint."
	case InputPhaseMeetingConfirmation:
		return "**approve** / **reject …** / item notes **1: …**; limit fallback **1/2/3**."
	case InputPhasePostMeeting:
		return "**get minutes/draft/open/conclusion** · **新会议** for next meeting."
	case InputPhaseExpertSetup:
		return "Follow the Moderator prompt; **cancel expert** to abort."
	case InputPhaseReceptionConfirm:
		return "Reply **1** to confirm or **0** to cancel."
	case InputPhaseReceptionClarify:
		return "Reply with the requested detail; **0** or **cancel expert** to abort."
	default:
		return "**新会议** · **!rt expert list** · **!rt principal bind** · **!rt help**"
	}
}

func webIdleInputHint(loc Locale) string {
	if loc == LocaleZH {
		return "可发送：**新会议** / **开个会** / **!rt 专家 列表** / **!rt help**"
	}
	return "**新会议** / **!rt meet** · **!rt expert list** · **!rt help**"
}

func phaseExpectsPrincipalInput(phase ChannelInputPhase) bool {
	switch phase {
	case InputPhaseSetupTopic, InputPhaseSetupParticipants, InputPhaseSetupBrief, InputPhaseSetupMenu, InputPhaseSetupCustom,
		InputPhaseMeetingConfirmation, InputPhaseMeetingPaused, InputPhasePostMeeting, InputPhaseExpertSetup,
		InputPhaseReceptionConfirm, InputPhaseReceptionClarify:
		return true
	default:
		return false
	}
}

// MisplacedInputHint returns a short hint when Principal sends unrecognized text in a waiting phase.
func (r *MeetRunner) MisplacedInputHint(msg transport.Inbound) (string, bool) {
	if strings.TrimSpace(msg.Content) == "" {
		return "", false
	}
	phase := r.InputPhase(msg.ChannelID)
	if !phaseExpectsPrincipalInput(phase) {
		return "", false
	}
	if !r.isScopePrincipal(msg) {
		return "", false
	}
	// Setup and confirmation already return parse/validation errors from their handlers.
	if phase == InputPhaseSetupTopic || phase == InputPhaseSetupParticipants || phase == InputPhaseSetupBrief || phase == InputPhaseSetupMenu || phase == InputPhaseSetupCustom ||
		phase == InputPhaseMeetingConfirmation {
		return "", false
	}
	loc := r.locale()
	return formatMisplacedInputHint(loc, phase), true
}

func (r *MeetRunner) isScopePrincipal(msg transport.Inbound) bool {
	binding, ok := r.bindingFor(msg)
	return ok && binding.ExternalID == msg.AuthorID
}

// bindingFor resolves Principal identity: web sessions are implicit; Discord uses Registry.
func (r *MeetRunner) bindingFor(msg transport.Inbound) (principalbind.Binding, bool) {
	if msg.Platform == "web" {
		return principalbind.Binding{
			PrincipalID: principalbind.PrincipalIDForExternal("web", msg.ChannelID),
			Platform:    "web",
			ExternalID:  msg.AuthorID,
			DisplayName: "浏览器",
			Scope:       principalbind.ScopeKey("web", "", msg.ChannelID),
		}, true
	}
	if r.Registry == nil {
		return principalbind.Binding{}, false
	}
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	b, ok := r.Registry.Get(scope)
	return b, ok
}

func formatMisplacedInputHint(loc Locale, phase ChannelInputPhase) string {
	hint := inputPhaseHint(loc, phase)
	if loc == LocaleZH {
		return fmt.Sprintf("❓ 当前不在等待该输入。\n\n%s\n\n发送 **会议状态** 可随时查看。", hint)
	}
	return fmt.Sprintf("❓ That input is not expected right now.\n\n%s\n\nSend **status** anytime.", hint)
}
