package discord

import (
	"fmt"
	"strings"

	prin "round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/domain/event"
)

func formatConfirmationBrief(loc Locale, meetingID string, cycle int, brief event.ConfirmationBrief) string {
	var b strings.Builder
	if loc == LocaleZH {
		fmt.Fprintf(&b, "📋 **Principal 确认关** · 第 %d 轮\n", cycle)
		fmt.Fprintf(&b, "🆔 `%s`\n\n", meetingID)
		if brief.ExecutiveSummary != "" {
			fmt.Fprintf(&b, "%s\n\n", brief.ExecutiveSummary)
		}
		for _, item := range brief.Items {
			fmt.Fprintf(&b, "**%d. %s**\n", item.Index, item.Title)
			if item.Description != "" {
				desc := clipMessageRunes(item.Description)
				fmt.Fprintf(&b, "%s\n", desc)
			}
			if item.Source != "" {
				fmt.Fprintf(&b, "_来源：%s_\n", item.Source)
			}
			b.WriteByte('\n')
		}
		b.WriteString("请审阅方案草案，回复：\n")
		b.WriteString("**批准** — 通过并归档\n")
		b.WriteString("**驳回** — 追加 1 轮研讨（可附修改意见，如：`驳回 技能数值需重算`）\n\n")
		b.WriteString("也可回复 **1** 批准 · **2** 驳回")
		return strings.TrimRight(b.String(), "\n")
	}

	fmt.Fprintf(&b, "📋 **Principal confirmation** · cycle %d\n", cycle)
	fmt.Fprintf(&b, "🆔 `%s`\n\n", meetingID)
	if brief.ExecutiveSummary != "" {
		fmt.Fprintf(&b, "%s\n\n", brief.ExecutiveSummary)
	}
	for _, item := range brief.Items {
		fmt.Fprintf(&b, "**%d. %s**\n", item.Index, item.Title)
		if item.Description != "" {
			fmt.Fprintf(&b, "%s\n", clipMessageRunes(item.Description))
		}
		if item.Source != "" {
			fmt.Fprintf(&b, "_Source: %s_\n", item.Source)
		}
		b.WriteByte('\n')
	}
	b.WriteString("Reply:\n**approve** / **1** — accept and finish\n")
	b.WriteString("**reject** / **2** — resume debate (optional feedback)\n")
	b.WriteString("Example: `reject need more detail on cooldowns`")
	return strings.TrimRight(b.String(), "\n")
}

var confirmApproveExact = map[string]bool{
	"1": true, "批准": true, "通过": true, "同意": true, "确认": true,
	"approve": true, "approved": true, "ok": true, "yes": true, "y": true,
}

var confirmRejectExact = map[string]bool{
	"2": true, "驳回": true, "拒绝": true, "退回": true,
	"reject": true, "rejected": true, "no": true, "n": true,
}

func parseConfirmationReply(content string) (prin.Response, error) {
	s := strings.TrimSpace(content)
	if s == "" {
		return prin.Response{}, errConfirmReplyEmpty
	}
	norm := normalizeASCIIForms(s)
	lower := strings.ToLower(norm)

	if confirmApproveExact[norm] || confirmApproveExact[lower] {
		return prin.Response{Decision: prin.DecisionApproved}, nil
	}
	if confirmRejectExact[norm] || confirmRejectExact[lower] {
		return prin.Response{Decision: prin.DecisionRejected}, nil
	}

	for _, prefix := range []string{"驳回", "拒绝", "退回", "reject", "rejected"} {
		if strings.HasPrefix(lower, prefix) {
			fb := strings.TrimSpace(s[len(prefix):])
			fb = strings.TrimPrefix(fb, "：")
			fb = strings.TrimPrefix(fb, ":")
			fb = strings.TrimSpace(fb)
			return prin.Response{Decision: prin.DecisionRejected, Feedback: fb}, nil
		}
	}
	for _, prefix := range []string{"批准", "通过", "同意", "approve", "approved"} {
		if strings.HasPrefix(lower, prefix) {
			return prin.Response{Decision: prin.DecisionApproved}, nil
		}
	}

	return prin.Response{}, errConfirmReplyUnrecognized
}

func confirmNotOwnerText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 只有本会议的 Principal 可以确认或驳回。"
	}
	return "⚠️ Only the meeting Principal can confirm or reject."
}

func confirmParseErrorText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ " + err.Error() + "\n请回复 **批准** / **驳回**，或 **1** / **2**。"
	}
	return "❌ " + err.Error() + "\nReply **approve** / **reject**, or **1** / **2**."
}

func confirmAlreadyAnsweredText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 本轮确认已收到回复，主持人正在处理。"
	}
	return "ℹ️ Confirmation reply already received."
}

func confirmReceivedApproveText(loc Locale) string {
	if loc == LocaleZH {
		return "✅ 已收到 **批准**，正在归档会议…"
	}
	return "✅ **Approved** — finishing meeting…"
}

func confirmReceivedRejectText(loc Locale) string {
	if loc == LocaleZH {
		return "↩ 已收到 **驳回**，将追加 **1 轮** 研讨…"
	}
	return "↩ **Rejected** — adding **one** debate round with your feedback…"
}
