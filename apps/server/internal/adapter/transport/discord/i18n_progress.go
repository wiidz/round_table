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
	case strings.HasPrefix(line, "… waiting for principal confirmation"):
		var cycle int
		if _, err := fmt.Sscanf(line, "… waiting for principal confirmation cycle=%d", &cycle); err == nil {
			return fmt.Sprintf("⏳ Waiting for Principal confirmation · cycle %d", cycle)
		}
		return line
	case strings.HasPrefix(line, "▶ confirmation prepared"):
		var cycle int
		if _, err := fmt.Sscanf(line, "▶ confirmation prepared cycle=%d", &cycle); err == nil {
			return fmt.Sprintf("📎 Confirmation brief prepared · cycle %d", cycle)
		}
		return line
	case strings.HasPrefix(line, "★ confirmation approved"):
		var cycle int
		if _, err := fmt.Sscanf(line, "★ confirmation approved cycle=%d", &cycle); err == nil {
			return fmt.Sprintf("✅ Principal approved · cycle %d", cycle)
		}
		return line
	case strings.HasPrefix(line, "↩ confirmation rejected"):
		var cycle, round int
		if _, err := fmt.Sscanf(line, "↩ confirmation rejected cycle=%d — starting round %d", &cycle, &round); err == nil {
			return fmt.Sprintf("↩ Principal rejected · cycle %d · starting round %d", cycle, round)
		}
		var cycleOnly int
		if _, err := fmt.Sscanf(line, "↩ confirmation rejected cycle=%d — adding one round", &cycleOnly); err == nil {
			return fmt.Sprintf("↩ Principal rejected · cycle %d · adding one round", cycleOnly)
		}
		if _, err := fmt.Sscanf(line, "↩ confirmation limit continue cycle=%d — reset cycles", &cycleOnly); err == nil {
			return fmt.Sprintf("↩ Limit continue · cycle %d · cycles reset", cycleOnly)
		}
		return line
	case strings.HasPrefix(line, "… confirmation limit reached"):
		return "⚠️ " + line
	case strings.HasPrefix(line, "↩ confirmation limit continue"):
		return "↩ " + line
	case strings.HasPrefix(line, "⏸ principal pause"):
		return "⏸ " + line
	case strings.HasPrefix(line, "◇ principal force synthesis"):
		return "⚡ " + line
	case strings.HasPrefix(line, "◇ principal force consensus"):
		return "🤝 " + line
	case strings.HasPrefix(line, "▶ free dialogue turn"):
		var idx, total int
		var answerer string
		if _, err := fmt.Sscanf(line, "▶ free dialogue turn %d/%d answerer=%s", &idx, &total, &answerer); err == nil {
			return fmt.Sprintf("💬 Free dialogue %d/%d · **%s** answering", idx, total, answerer)
		}
		return line
	case strings.HasPrefix(line, "▶ free dialogue after round"):
		var round, total, maxQ int
		if _, err := fmt.Sscanf(line, "▶ free dialogue after round %d (%d Q&A pairs, max_questions=%d/person)", &round, &total, &maxQ); err == nil {
			return fmt.Sprintf("💬 Free dialogue after round %d · %d Q&A pairs · max %d/person\n\nPrincipal: send **ask …**", round, total, maxQ)
		}
		return line
	case strings.HasPrefix(line, "◆ free dialogue question "):
		return localizeFreeDialogueQuestionLine(line, LocaleEN)
	case strings.HasPrefix(line, "◆ free dialogue answer "):
		return localizeFreeDialogueAnswerLine(line, LocaleEN)
	case strings.HasPrefix(line, "■ free dialogue completed"):
		var n int
		if _, err := fmt.Sscanf(line, "■ free dialogue completed (%d exchanges)", &n); err == nil {
			return fmt.Sprintf("✅ Free dialogue completed · %d exchanges", n)
		}
		return line
	default:
		return line
	}
}

func mergeMeetingStartProgress(engineLine, preMeetingLine string, loc Locale) string {
	run := localizeProgressLine(engineLine, loc)
	prep := localizeProgressLine(preMeetingLine, loc)
	return run + "\n" + prep
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

	case strings.HasPrefix(line, "▶ free dialogue turn"):
		var idx, total int
		var answerer string
		if _, err := fmt.Sscanf(line, "▶ free dialogue turn %d/%d answerer=%s", &idx, &total, &answerer); err == nil {
			return fmt.Sprintf("💬 自由问答 %d/%d · 轮到 **%s** 回答", idx, total, participantLabel(answerer, LocaleZH))
		}

	case strings.HasPrefix(line, "▶ free dialogue after round"):
		var round, total, maxQ int
		if _, err := fmt.Sscanf(line, "▶ free dialogue after round %d (%d Q&A pairs, max_questions=%d/person)", &round, &total, &maxQ); err == nil {
			return fmt.Sprintf("💬 第 %d 轮后自由问答 · %d 组问答 · 每人最多 %d 问\n\nPrincipal 可发送 **提问 …**", round, total, maxQ)
		}

	case strings.HasPrefix(line, "◆ free dialogue question "):
		return localizeFreeDialogueQuestionLine(line, LocaleZH)

	case strings.HasPrefix(line, "◆ free dialogue answer "):
		return localizeFreeDialogueAnswerLine(line, LocaleZH)

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

	case strings.HasPrefix(line, "… waiting for principal confirmation"):
		var cycle int
		if _, err := fmt.Sscanf(line, "… waiting for principal confirmation cycle=%d", &cycle); err == nil {
			return fmt.Sprintf("⏳ 等待 Principal 确认 · 第 %d 轮", cycle)
		}

	case strings.HasPrefix(line, "… confirmation limit reached"):
		var cycle int
		if _, err := fmt.Sscanf(line, "… confirmation limit reached cycle=%d — awaiting principal fallback", &cycle); err == nil {
			return fmt.Sprintf("⚠️ 确认关已达上限 · 第 %d 轮 · 请选择兜底方案", cycle)
		}

	case strings.HasPrefix(line, "↩ confirmation limit continue"):
		return "↩ 已选择继续研讨 · 确认轮次已重置"

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
		var cycle, round int
		if _, err := fmt.Sscanf(line, "↩ confirmation rejected cycle=%d — starting round %d", &cycle, &round); err == nil {
			return fmt.Sprintf("↩ Principal 驳回 · 第 %d 轮 · 追加第 %d 轮研讨", cycle, round)
		}
		var cycleOnly int
		if _, err := fmt.Sscanf(line, "↩ confirmation rejected cycle=%d — adding one round", &cycleOnly); err == nil {
			return fmt.Sprintf("↩ Principal 驳回 · 第 %d 轮 · 追加 1 轮研讨", cycleOnly)
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

	case strings.HasPrefix(line, "⏸ principal pause"):
		reason := strings.TrimPrefix(line, "⏸ principal pause (")
		reason = strings.TrimSuffix(reason, ")")
		return fmt.Sprintf("⏸ Principal 暂停会议 · %s", reason)

	case strings.HasPrefix(line, "◇ principal force synthesis"):
		reason := strings.TrimPrefix(line, "◇ principal force synthesis (")
		reason = strings.TrimSuffix(reason, ")")
		return fmt.Sprintf("⚡ Principal 要求立即合成 · %s", reason)

	case strings.HasPrefix(line, "◇ principal force consensus"):
		reason := strings.TrimPrefix(line, "◇ principal force consensus (")
		reason = strings.TrimSuffix(reason, ")")
		return fmt.Sprintf("🤝 Principal 强制共识 · %s", reason)
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

func localizeFreeDialogueQuestionLine(line string, loc Locale) string {
	rest := strings.TrimPrefix(line, "◆ free dialogue question ")
	parts := strings.SplitN(rest, "\n", 2)
	header := parts[0]
	body := ""
	if len(parts) > 1 {
		body = strings.TrimSpace(parts[1])
	}
	var idx, total int
	var asker, answerer string
	if strings.Contains(header, "principal →") {
		if _, err := fmt.Sscanf(header, "%d/%d principal → %s", &idx, &total, &answerer); err != nil {
			return line
		}
		if loc == LocaleZH {
			msg := fmt.Sprintf("❓ 自由问答 %d/%d · **Principal**（主持人转达）→ **%s**",
				idx, total, participantLabel(answerer, loc))
			if body != "" {
				msg += "\n\n" + body
			}
			return msg
		}
		msg := fmt.Sprintf("❓ Free dialogue %d/%d · **Principal** (via Moderator) → **%s**",
			idx, total, answerer)
		if body != "" {
			msg += "\n\n" + body
		}
		return msg
	}
	if _, err := fmt.Sscanf(header, "%d/%d %s → %s", &idx, &total, &asker, &answerer); err != nil {
		return line
	}
	if loc == LocaleZH {
		msg := fmt.Sprintf("❓ 自由问答 %d/%d · **%s** → **%s**",
			idx, total, participantLabel(asker, loc), participantLabel(answerer, loc))
		if body != "" {
			msg += "\n\n" + body
		}
		return msg
	}
	msg := fmt.Sprintf("❓ Free dialogue %d/%d · **%s** → **%s**", idx, total, asker, answerer)
	if body != "" {
		msg += "\n\n" + body
	}
	return msg
}

func localizeFreeDialogueAnswerLine(line string, loc Locale) string {
	rest := strings.TrimPrefix(line, "◆ free dialogue answer ")
	parts := strings.SplitN(rest, "\n", 2)
	header := parts[0]
	body := ""
	if len(parts) > 1 {
		body = strings.TrimSpace(parts[1])
	}
	var idx, total int
	var answerer, asker string
	if _, err := fmt.Sscanf(header, "%d/%d %s → %s", &idx, &total, &answerer, &asker); err != nil {
		return line
	}
	if loc == LocaleZH {
		msg := fmt.Sprintf("💡 自由问答 %d/%d · **%s** 回答 **%s**",
			idx, total, participantLabel(answerer, loc), participantLabel(asker, loc))
		if body != "" {
			msg += "\n\n" + body
		}
		return msg
	}
	msg := fmt.Sprintf("💡 Free dialogue %d/%d · **%s** answers **%s**", idx, total, answerer, asker)
	if body != "" {
		msg += "\n\n" + body
	}
	return msg
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
