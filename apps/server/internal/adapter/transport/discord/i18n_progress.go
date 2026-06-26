package discord

import (
	"fmt"
	"strconv"
	"strings"
)

func localizeProgressLine(line string, loc Locale) string {
	if loc == LocaleEN {
		return localizeProgressEN(line)
	}
	return localizeProgressZH(line)
}

func localizeProgressEN(line string) string {
	// Light emoji polish for English users.
	switch {
	case strings.HasPrefix(line, "▶ engine run started"):
		return "🚀 " + line
	case strings.HasPrefix(line, "▶ pre-meeting started"):
		return "📋 " + line
	case strings.HasPrefix(line, "▶ debate round"):
		return "🔄 " + line
	case strings.HasPrefix(line, "■ pre-meeting completed"):
		return "✅ " + line
	case strings.HasPrefix(line, "■ debate round"):
		return "✅ " + line
	case strings.HasPrefix(line, "★ synthesis completed"):
		return "🎉 " + line
	case strings.HasPrefix(line, "★ consensus reached"):
		return "🤝 " + line
	case strings.HasPrefix(line, "■ meeting finished"):
		return "🏁 " + line
	case strings.HasPrefix(line, "◇ deliberation ready"):
		return "✨ " + line
	case strings.HasPrefix(line, "◇ max deliberation rounds"):
		return "⏱️ " + line
	case strings.HasPrefix(line, "◇ synthesis readiness: ready"):
		return "✅ " + line
	case strings.HasPrefix(line, "◇ synthesis readiness: not ready"):
		return "⏳ " + line
	default:
		return line
	}
}

func localizeProgressZH(line string) string {
	switch {
	case strings.HasPrefix(line, "▶ engine run started"):
		rest := strings.TrimPrefix(line, "▶ engine run started ")
		meetingID := extractKV(rest, "meeting")
		status := extractKV(rest, "status")
		return fmt.Sprintf("🚀 会议运行中 · %s · `%s`", statusLabel(status, LocaleZH), meetingID)

	case strings.HasPrefix(line, "▶ pre-meeting started"):
		order := strings.TrimPrefix(line, "▶ pre-meeting started order=")
		return fmt.Sprintf("📋 会前准备开始 · 顺序 %s", localizeOrder(order, LocaleZH))

	case strings.HasPrefix(line, "▶ debate round"):
		rest := strings.TrimPrefix(line, "▶ debate round ")
		before, order, ok := strings.Cut(rest, " started order=")
		if !ok {
			return "🔄 " + line
		}
		var round int
		fmt.Sscanf(strings.TrimSpace(before), "%d", &round)
		return fmt.Sprintf("🔄 第 %d 轮研讨开始 · 顺序 %s", round, localizeOrder(order, LocaleZH))

	case line == "■ pre-meeting completed → starting debate round 1":
		return "✅ 会前准备结束 · 进入第 1 轮研讨"

	case strings.HasPrefix(line, "■ debate round"):
		var round int
		if _, err := fmt.Sscanf(line, "■ debate round %d completed", &round); err == nil {
			return fmt.Sprintf("✅ 第 %d 轮研讨结束", round)
		}
		return "✅ " + line

	case strings.HasPrefix(line, "▶ free dialogue after round"):
		var round, total, maxQ int
		if _, err := fmt.Sscanf(line, "▶ free dialogue after round %d (%d Q&A pairs, max_questions=%d/person)", &round, &total, &maxQ); err == nil {
			return fmt.Sprintf("💬 第 %d 轮后自由问答 · %d 组问答 · 每人最多 %d 问", round, total, maxQ)
		}

	case strings.HasPrefix(line, "■ free dialogue completed"):
		var n int
		if _, err := fmt.Sscanf(line, "■ free dialogue completed (%d exchanges)", &n); err == nil {
			return fmt.Sprintf("✅ 自由问答结束 · 共 %d 次交流", n)
		}

	case strings.HasPrefix(line, "★ synthesis completed"):
		rest := strings.TrimPrefix(line, "★ synthesis completed ")
		resolved := extractKV(rest, "resolved_by")
		openQ := extractKV(rest, "open_questions")
		n, _ := strconv.Atoi(openQ)
		return fmt.Sprintf("🎉 设计草案合成完成 · %s · 开放问题 %d 条",
			resolvedByLabel(resolved, LocaleZH), n)

	case strings.HasPrefix(line, "★ consensus reached"):
		rest := strings.TrimPrefix(line, "★ consensus reached ")
		strategy := extractKV(rest, "strategy")
		resolved := extractKV(rest, "resolved_by")
		return fmt.Sprintf("🤝 达成共识 · 策略 %s · %s",
			meetingModeLabel(strategy, LocaleZH), resolvedByLabel(resolved, LocaleZH))

	case strings.HasPrefix(line, "■ meeting finished outcome="):
		outcome := strings.TrimPrefix(line, "■ meeting finished outcome=")
		return fmt.Sprintf("🏁 会议结束 · 结果：%s", outcomeLabel(outcome, LocaleZH))

	case strings.HasPrefix(line, "◇ synthesis readiness: ready"):
		rationale := strings.TrimPrefix(line, "◇ synthesis readiness: ready (")
		rationale = strings.TrimSuffix(rationale, ")")
		return fmt.Sprintf("✅ 研讨就绪 · %s", rationale)

	case strings.HasPrefix(line, "◇ synthesis readiness: not ready — "):
		return "⏳ 继续研讨 · " + strings.TrimPrefix(line, "◇ synthesis readiness: not ready — ")

	case strings.HasPrefix(line, "◇ synthesis readiness: not ready ("):
		rationale := strings.TrimPrefix(line, "◇ synthesis readiness: not ready (")
		rationale = strings.TrimSuffix(rationale, ")")
		return fmt.Sprintf("⏳ 继续研讨 · %s", rationale)

	case strings.HasPrefix(line, "◇ deliberation ready at round"):
		var round int
		if _, err := fmt.Sscanf(line, "◇ deliberation ready at round %d — synthesizing design draft", &round); err == nil {
			return fmt.Sprintf("✨ 第 %d 轮研讨就绪 · 开始合成设计草案", round)
		}

	case strings.HasPrefix(line, "◇ max deliberation rounds reached"):
		var max int
		if _, err := fmt.Sscanf(line, "◇ max deliberation rounds reached (%d) — synthesizing design draft", &max); err == nil {
			return fmt.Sprintf("⏱️ 已达最大轮次 (%d) · 开始合成设计草案", max)
		}

	case strings.HasPrefix(line, "▶ confirmation prepared"):
		var cycle int
		if _, err := fmt.Sscanf(line, "▶ confirmation prepared cycle=%d", &cycle); err == nil {
			return fmt.Sprintf("📎 确认清单已准备 · 第 %d 轮", cycle)
		}

	case strings.HasPrefix(line, "★ confirmation approved"):
		var cycle int
		if _, err := fmt.Sscanf(line, "★ confirmation approved cycle=%d", &cycle); err == nil {
			return fmt.Sprintf("✅ Principal 已确认 · 第 %d 轮", cycle)
		}

	case strings.HasPrefix(line, "↩ confirmation rejected"):
		var cycle int
		if _, err := fmt.Sscanf(line, "↩ confirmation rejected cycle=%d — resuming debate", &cycle); err == nil {
			return fmt.Sprintf("↩ Principal 驳回 · 第 %d 轮 · 继续研讨", cycle)
		}

	case strings.HasPrefix(line, "⏸ meeting paused"):
		reason := strings.TrimPrefix(line, "⏸ meeting paused (")
		reason = strings.TrimSuffix(reason, ")")
		return fmt.Sprintf("⏸ 会议暂停 · %s", reason)

	case line == "▶ meeting resumed":
		return "▶ 会议已恢复"

	case strings.HasPrefix(line, "■ principal abort"):
		reason := strings.TrimPrefix(line, "■ principal abort (")
		reason = strings.TrimSuffix(reason, ")")
		return fmt.Sprintf("🛑 Principal 中止会议 · %s", reason)
	}

	return line
}

func localizeOrder(order string, loc Locale) string {
	parts := strings.Split(order, " → ")
	if loc != LocaleZH {
		return order
	}
	for i, p := range parts {
		parts[i] = participantLabel(strings.TrimSpace(p), loc)
	}
	return strings.Join(parts, " → ")
}

func extractKV(rest, key string) string {
	prefix := key + "="
	for _, field := range strings.Fields(rest) {
		if strings.HasPrefix(field, prefix) {
			return strings.TrimPrefix(field, prefix)
		}
	}
	return ""
}

func localizeStreamDetail(detail string, loc Locale) string {
	if loc != LocaleZH || detail == "" {
		return detail
	}
	// turn (1/4) · round 0
	if strings.HasPrefix(detail, "turn ") && strings.Contains(detail, "round ") {
		var cur, total, round int
		if _, err := fmt.Sscanf(detail, "turn (%d/%d) · round %d", &cur, &total, &round); err == nil {
			return fmt.Sprintf("第 %d/%d 发言 · 轮次 %d", cur, total, round)
		}
	}
	if strings.HasSuffix(detail, " readiness") {
		var round int
		if _, err := fmt.Sscanf(detail, "round %d readiness", &round); err == nil {
			return fmt.Sprintf("第 %d 轮就绪评估", round)
		}
	}
	if detail == "design-draft synthesis" {
		return "设计草案合成"
	}
	return detail
}

func formatStreamStart(meta streamMeta, loc Locale) string {
	who := participantLabel(meta.ParticipantID, loc)
	phase := phaseLabel(meta.Phase, loc)
	detail := localizeStreamDetail(meta.Detail, loc)
	if loc == LocaleZH {
		if detail != "" {
			return fmt.Sprintf("🎤 **%s** · %s · %s", who, phase, detail)
		}
		return fmt.Sprintf("🎤 **%s** · %s", who, phase)
	}
	if detail != "" {
		return fmt.Sprintf("🎤 **%s** · %s · %s", who, phase, detail)
	}
	return fmt.Sprintf("🎤 **%s** · %s", who, phase)
}

// streamMeta avoids importing stream package in tests that only need labels.
type streamMeta struct {
	ParticipantID string
	Phase         string
	Detail        string
}
