package discord

import "strings"

func isArtifactFetchTrigger(content string) (kind string, ok bool) {
	s := strings.TrimSpace(content)
	if s == "" {
		return "", false
	}
	norm := normalizeASCIIForms(s)
	lower := strings.ToLower(norm)
	switch {
	case matchExact(lower, "获取纪要", "获取 minutes", "get minutes", "minutes"):
		return "minutes", true
	case matchExact(lower, "获取草案", "获取 draft", "get draft", "draft"):
		return "draft", true
	case matchExact(lower, "获取待决", "获取开放问题", "get open", "open questions", "open"):
		return "open", true
	case matchExact(lower, "获取结论", "获取结论纪要", "get conclusion", "conclusion"):
		return "conclusion", true
	}
	return "", false
}

func artifactFetchUsageText(loc Locale) string {
	if loc == LocaleZH {
		return "用法：**获取纪要** · **获取草案** · **获取待决** · **获取结论**"
	}
	return "Usage: **get minutes** · **get draft** · **get open** · **get conclusion**"
}

func artifactFetchMissingText(loc Locale, kind, meetingID string) string {
	if loc == LocaleZH {
		return "ℹ️ 会议 `" + meetingID + "` 没有可用的 **" + kind + "** 文件。"
	}
	return "ℹ️ Meeting `" + meetingID + "` has no **" + kind + "** artifact."
}

func artifactFetchSentText(loc Locale, kind string) string {
	title := artifactTitle(loc, kind)
	if loc == LocaleZH {
		return "✅ 已推送 " + title
	}
	return "✅ Sent " + title
}

func artifactFetchNoMeetingText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 本频道暂无已结束的会议记录。请先完成一场会议。"
	}
	return "ℹ️ No completed meeting in this channel yet."
}
