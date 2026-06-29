package discord

import "strings"

// Locale selects Discord-facing copy (engine logs stay English for CLI).
type Locale string

const (
	LocaleEN Locale = "en"
	LocaleZH Locale = "zh"
)

// ParseLocale normalizes config values; default is English.
func ParseLocale(raw string) Locale {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "zh", "zh-cn", "zh_cn", "zh-hans", "chinese", "cn":
		return LocaleZH
	default:
		return LocaleEN
	}
}

func participantLabel(id string, loc Locale) string {
	if loc != LocaleZH {
		return id
	}
	switch id {
	case "designer":
		return "方案"
	case "player":
		return "代表"
	case "dev":
		return "开发"
	case "ops":
		return "运维"
	case "moderator":
		return "主持人"
	default:
		return id
	}
}

func phaseLabel(phase string, loc Locale) string {
	if loc != LocaleZH {
		return phase
	}
	switch phase {
	case "pre-meeting":
		return "会前准备"
	case "deliberation":
		return "研讨"
	case "debate":
		return "辩论"
	case "free-dialogue-ask":
		return "自由问答·提问"
	case "free-dialogue-answer":
		return "自由问答·回答"
	case "deliberation-synthesis":
		return "草案合成"
	case "deliberation-readiness":
		return "合成就绪"
	default:
		return phase
	}
}

func meetingModeLabel(mode string, loc Locale) string {
	if loc != LocaleZH {
		return mode
	}
	switch mode {
	case "deliberation":
		return "研讨"
	case "decision":
		return "裁决"
	default:
		return mode
	}
}

func resolvedByLabel(v string, loc Locale) string {
	if loc != LocaleZH {
		return v
	}
	switch v {
	case "synthesis":
		return "研讨合成"
	case "max_rounds":
		return "达最大轮次"
	case "moderator":
		return "主持人"
	case "deliberation":
		return "研讨模式"
	default:
		return v
	}
}

func outcomeLabel(v string, loc Locale) string {
	if loc != LocaleZH {
		return v
	}
	switch strings.ToLower(v) {
	case "completed":
		return "已完成"
	case "aborted":
		return "已中止"
	case "failed":
		return "失败"
	default:
		return v
	}
}

func statusLabel(v string, loc Locale) string {
	if loc != LocaleZH {
		return v
	}
	switch v {
	case "Preparing":
		return "准备中"
	case "Running":
		return "进行中"
	case "Completed":
		return "已完成"
	case "Paused":
		return "已暂停"
	default:
		return v
	}
}
