package discord

import (
	"fmt"
	"strings"
)

// WebInputStatusReply is deprecated; web handler uses FormatInputPhaseStatus with live phase.
func WebInputStatusReply(loc Locale) string {
	return FormatInputPhaseStatus(loc, InputPhaseIdle, "")
}

// MatchWebExpertListTrigger reports natural-language expert roster queries.
func MatchWebExpertListTrigger(content string) bool {
	s := strings.TrimSpace(content)
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	for _, phrase := range []string{
		"专家列表", "专家名录", "有哪些专家", "列出专家", "查看专家",
		"expert list", "list experts", "show experts",
	} {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// WebChatNoMatchReply is shown when no handler matched (avoid silent failures).
func WebChatNoMatchReply(loc Locale) string {
	if loc == LocaleZH {
		return "暂未理解这条消息。可试：**新会议**、**会议状态**、**!rt 专家 列表**、**!rt help**。"
	}
	return "I didn't understand that. Try **新会议**, **status**, **!rt expert list**, or **!rt help**."
}

// MatchWebStatusTrigger reports 会议状态 / status shortcuts for browser chat.
func MatchWebStatusTrigger(content string) bool {
	return isInputStatusTrigger(content)
}

// WebHelpText returns browser chat help.
func WebHelpText(loc Locale) string {
	if loc == LocaleZH {
		return `📖 **RoundTable 浏览器聊天**

- **新会议** / **开个会** / **!rt meet** — 发起会议（可选简报模板；无需 Principal 绑定）
- **!rt meet -template 模板ID 主题** — 用简报模板预填开会
- **会议状态** — 查看当前输入态
- **!rt 专家 列表** — 专家名录
- **获取纪要** / **获取草案** — 会议结束后拉取交付物
- 运行中：**暂停会议** · **终止会议** · **提问 …**（自由问答阶段）
- 自然语言提问 — 查询专家、会议状态（Reception Agent）`
	}
	return `📖 **RoundTable Web Chat**

- **新会议** / **!rt meet** — start a meeting (optional brief template; no Principal bind)
- **!rt meet -template id topic** — start with a brief template
- **status** — input phase
- **!rt expert list** — roster
- **fetch minutes** / artifacts after meeting
- Running: pause/abort/ask questions in free dialogue
- Natural language queries (Reception Agent)`
}

// FormatInputPhaseStatus formats channel input phase for web chat.
func FormatInputPhaseStatus(loc Locale, phase ChannelInputPhase, meetingID string) string {
	return formatInputPhaseStatusForPlatform(loc, phase, meetingID, "web")
}

// MeetingIDForPhase returns meeting id for status display.
func (r *MeetRunner) MeetingIDForPhase(channelID string, phase ChannelInputPhase) string {
	return r.meetingIDForPhase(channelID, phase)
}

func IsExpertCancelTrigger(content string) bool  { return isExpertCancelTrigger(content) }
func IsMeetCancelTrigger(content string) bool    { return isMeetCancelTrigger(content) }
func IsMeetStartTrigger(content string) bool     { return isMeetStartTrigger(content) }
func MeetSetupNothingToCancelText(loc Locale) string { return meetSetupNothingToCancelText(loc) }
func MeetDisabledText(loc Locale) string         { return meetDisabledText(loc) }
func MeetUsageText(loc Locale, prefix string) string { return meetUsageText(loc, prefix) }
func ParseMeetArgs(args []string, defaultMode string) (meetParseResult, error) {
	return parseMeetArgs(args, defaultMode)
}

func webChatPrincipalUnavailableText(loc Locale) string {
	if loc == LocaleZH {
		return "浏览器聊天不使用 Principal 绑定；请直接发送消息与主持人对话。"
	}
	return "Web chat does not use Principal binding; send messages to talk to the moderator."
}

func webChatMeetUnavailableText(loc Locale) string {
	if loc == LocaleZH {
		return "浏览器聊天暂不支持发起或进行会议；请使用 Discord Transport，或在后续版本等待此能力开放。"
	}
	return "Web chat cannot start or run meetings yet; use Discord Transport or wait for a future release."
}

func webChatMutatingUnavailableText(loc Locale) string {
	if loc == LocaleZH {
		return "浏览器聊天暂不支持写操作（新建专家等）；请在 Discord 或 Web 设置页操作。"
	}
	return "Web chat does not support write actions yet (create expert, …); use Discord or Web settings."
}

// WebChatMutatingUnavailableText explains write actions are unavailable in browser chat.
func WebChatMutatingUnavailableText(loc Locale) string {
	return webChatMutatingUnavailableText(loc)
}

// IsWebPlatformHint reports copy that should render as system (not 主持人) in browser chat.
func IsWebPlatformHint(reply string) bool {
	s := strings.TrimSpace(reply)
	if s == "" {
		return false
	}
	for _, hint := range []string{
		webChatMutatingUnavailableText(LocaleZH),
		webChatMutatingUnavailableText(LocaleEN),
		webChatMeetUnavailableText(LocaleZH),
		webChatMeetUnavailableText(LocaleEN),
		webChatPrincipalUnavailableText(LocaleZH),
		webChatPrincipalUnavailableText(LocaleEN),
	} {
		if s == hint {
			return true
		}
	}
	return strings.Contains(s, "浏览器聊天暂不支持") ||
		strings.Contains(s, "Web chat does not support") ||
		strings.Contains(s, "Web chat cannot start")
}

// MatchWebDiscordOnlyAction reports triggers that remain Discord-only on web (none for meetings).
func MatchWebDiscordOnlyAction(content string) (kind string, blocked bool) {
	return "", false
}

// WebChatBlockedActionHint returns a hint when Discord-only triggers are sent on web.
func WebChatBlockedActionHint(loc Locale, kind string) string {
	switch kind {
	case "meet":
		return webChatMeetUnavailableText(loc)
	case "principal":
		return webChatPrincipalUnavailableText(loc)
	default:
		if loc == LocaleZH {
			return fmt.Sprintf("浏览器聊天暂不支持该操作（%s）。", kind)
		}
		return fmt.Sprintf("That action is not available in web chat (%s).", kind)
	}
}
