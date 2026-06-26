package discord

import (
	"errors"
	"strings"

	prin "round_table/apps/server/internal/adapter/principal"
)

var (
	errInterventionUnrecognized = errors.New("无法识别指令")
	errInterventionWrongPhase     = errors.New("当前处于确认关，请先批准或驳回")
)

func isInterventionTrigger(content string) bool {
	_, ok := parseIntervention(strings.TrimSpace(content))
	return ok
}

func parseIntervention(content string) (prin.RunningIntervention, bool) {
	s := strings.TrimSpace(content)
	if s == "" {
		return prin.RunningIntervention{}, false
	}
	norm := normalizeASCIIForms(s)
	lower := strings.ToLower(norm)

	switch {
	case matchExact(lower, "暂停会议", "暂停", "pause"):
		return prin.RunningIntervention{Kind: prin.RunningInterventionPause, Reason: "Principal 暂停会议"}, true
	case matchExact(lower, "恢复会议", "恢复", "resume"):
		return prin.RunningIntervention{Kind: prin.RunningInterventionResume, Reason: "Principal 恢复会议"}, true
	case matchExact(lower, "终止会议", "中止会议", "停止会议", "abort"):
		return prin.RunningIntervention{Kind: prin.RunningInterventionAbort, Reason: "Principal 终止会议"}, true
	case matchExact(lower, "立即合成", "强制合成", "force synthesis"):
		return prin.RunningIntervention{Kind: prin.RunningInterventionForceSynthesis, Reason: "Principal 要求立即合成"}, true
	case matchExact(lower, "强制共识", "force consensus"):
		return prin.RunningIntervention{Kind: prin.RunningInterventionForceConsensus, Reason: "Principal 要求强制共识"}, true
	}

	for _, prefix := range []string{"终止会议", "中止会议", "停止会议", "abort"} {
		if strings.HasPrefix(lower, prefix) {
			reason := strings.TrimSpace(s[len(prefix):])
			reason = strings.TrimPrefix(reason, "：")
			reason = strings.TrimPrefix(reason, ":")
			reason = strings.TrimSpace(reason)
			if reason == "" {
				reason = "Principal 终止会议"
			}
			return prin.RunningIntervention{Kind: prin.RunningInterventionAbort, Reason: reason}, true
		}
	}
	for _, prefix := range []string{"暂停会议", "暂停", "pause"} {
		if strings.HasPrefix(lower, prefix) && len(s) > len(prefix) {
			reason := strings.TrimSpace(s[len(prefix):])
			reason = strings.TrimPrefix(reason, "：")
			reason = strings.TrimPrefix(reason, ":")
			reason = strings.TrimSpace(reason)
			if reason != "" {
				return prin.RunningIntervention{Kind: prin.RunningInterventionPause, Reason: reason}, true
			}
		}
	}

	return prin.RunningIntervention{}, false
}

func matchExact(lower string, variants ...string) bool {
	for _, v := range variants {
		if lower == strings.ToLower(v) {
			return true
		}
	}
	return false
}

func interventionAckText(loc Locale, kind prin.RunningInterventionKind) string {
	if loc == LocaleZH {
		switch kind {
		case prin.RunningInterventionPause:
			return "⏸ 已收到 **暂停** 指令，将在当前发言结束后生效…"
		case prin.RunningInterventionAbort:
			return "🛑 已收到 **终止** 指令，将在当前发言结束后生效…"
		case prin.RunningInterventionForceSynthesis:
			return "⚡ 已收到 **立即合成** 指令，将在当前发言结束后生效…"
		case prin.RunningInterventionForceConsensus:
			return "🤝 已收到 **强制共识** 指令，将在当前发言结束后生效…"
		case prin.RunningInterventionResume:
			return "▶ 已收到 **恢复**，会议继续…"
		default:
			return "✅ 指令已收到"
		}
	}
	switch kind {
	case prin.RunningInterventionPause:
		return "⏸ **Pause** queued — takes effect after the current turn…"
	case prin.RunningInterventionAbort:
		return "🛑 **Abort** queued — takes effect after the current turn…"
	case prin.RunningInterventionForceSynthesis:
		return "⚡ **Force synthesis** queued…"
	case prin.RunningInterventionForceConsensus:
		return "🤝 **Force consensus** queued…"
	case prin.RunningInterventionResume:
		return "▶ **Resume** — meeting continues…"
	default:
		return "✅ Command received"
	}
}

func interventionNotOwnerText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 只有本会议的 Principal 可以控制会议进程。"
	}
	return "⚠️ Only the meeting Principal can control the run."
}

func interventionParseErrorText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ " + err.Error() + "\n可用：**暂停会议** · **恢复会议** · **终止会议** · **立即合成** · **强制共识**"
	}
	return "❌ " + err.Error() + "\nTry: **pause** · **resume** · **abort** · **force synthesis** · **force consensus**"
}

func interventionAlreadyQueuedText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 上一条控制指令尚未执行，请稍候。"
	}
	return "ℹ️ Previous control command is still pending."
}

func interventionConfirmBlocksText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 确认关进行中，请先 **批准** 或 **驳回**。"
	}
	return "ℹ️ Confirmation pending — reply **approve** or **reject** first."
}

func interventionNoMeetingText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 当前频道没有进行中的会议。"
	}
	return "ℹ️ No meeting is running in this channel."
}

func formatPausedWaitPrompt(loc Locale) string {
	if loc == LocaleZH {
		return "⏸ **会议已暂停**\n\n发送 **恢复会议** 继续，或 **终止会议** 中止。"
	}
	return "⏸ **Meeting paused**\n\nSend **resume** to continue or **abort** to stop."
}
