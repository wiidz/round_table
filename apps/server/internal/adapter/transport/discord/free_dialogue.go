package discord

import (
	"errors"
	"fmt"
	"strings"
)

var errFreeDialogueQuestionEmpty = errors.New("请在「提问」后输入问题内容")

func isFreeDialogueQuestionTrigger(content string) bool {
	_, _, ok := parseFreeDialogueQuestion(content)
	return ok
}

// parseFreeDialogueQuestion extracts question text and optional target participant ID.
// Forms: `提问 analyst 风险如何评估` · `提问 风险如何评估` · `ask ops: timeline?`
func parseFreeDialogueQuestion(content string) (question, answererID string, ok bool) {
	s := strings.TrimSpace(content)
	if s == "" {
		return "", "", false
	}
	norm := normalizeASCIIForms(s)
	lower := strings.ToLower(norm)

	for _, prefix := range []string{"提问", "ask", "question"} {
		pl := strings.ToLower(prefix)
		if lower == pl {
			return "", "", true
		}
		if !strings.HasPrefix(lower, pl) {
			continue
		}
		rest := strings.TrimSpace(s[len(prefix):])
		rest = strings.TrimPrefix(rest, "：")
		rest = strings.TrimPrefix(rest, ":")
		rest = strings.TrimSpace(rest)
		if rest == "" {
			return "", "", true
		}
		if target, q, matched := splitFreeDialogueTarget(rest); matched {
			return q, target, true
		}
		return rest, "", true
	}
	return "", "", false
}

func splitFreeDialogueTarget(rest string) (target, question string, ok bool) {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", "", false
	}
	for _, sep := range []string{":", "："} {
		if before, after, found := strings.Cut(rest, sep); found {
			target = strings.TrimPrefix(strings.TrimSpace(before), "@")
			q := strings.TrimSpace(after)
			if looksLikeParticipantID(target) && q != "" {
				return target, q, true
			}
		}
	}
	first, remainder, found := strings.Cut(rest, " ")
	if !found {
		return "", "", false
	}
	target = strings.TrimPrefix(strings.TrimSpace(first), "@")
	remainder = strings.TrimSpace(remainder)
	remainder = strings.TrimPrefix(remainder, "：")
	remainder = strings.TrimPrefix(remainder, ":")
	remainder = strings.TrimSpace(remainder)
	if target == "" || remainder == "" {
		return "", "", false
	}
	if !looksLikeParticipantID(target) {
		return "", "", false
	}
	return target, remainder, true
}

func looksLikeParticipantID(s string) bool {
	if s == "" || strings.ContainsAny(s, " \t\n") {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func freeDialogueQuestionAckText(loc Locale, question, answererID string) string {
	target := ""
	if answererID != "" {
		if loc == LocaleZH {
			target = fmt.Sprintf(" → **%s**", participantLabel(answererID, loc))
		} else {
			target = fmt.Sprintf(" → **%s**", answererID)
		}
	}
	if loc == LocaleZH {
		return fmt.Sprintf("💬 已收到提问，将在当前发言结束后发出%s：\n> %s", target, question)
	}
	return fmt.Sprintf("💬 Question queued — will ask after the current turn%s:\n> %s", target, question)
}

func freeDialogueQuestionNotOwnerText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 只有本会议的 Principal 可以在自由问答中提问。"
	}
	return "⚠️ Only the meeting Principal can ask during free dialogue."
}

func freeDialogueQuestionWrongPhaseText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 当前不在自由问答阶段。Round 1 结束后可发送 **提问 …**。"
	}
	return "ℹ️ Free dialogue is not active. Send **ask …** after Round 1 completes."
}

func freeDialogueQuestionAlreadyQueuedText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 上一条 Principal 提问尚未转达，请稍候。"
	}
	return "ℹ️ Previous Principal question is still pending."
}

func freeDialogueQuestionConfirmBlocksText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 确认关进行中，请先 **批准** 或 **驳回**。"
	}
	return "ℹ️ Confirmation pending — reply **approve** or **reject** first."
}

func freeDialogueQuestionParseErrorText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ " + err.Error() + "\n用法：**提问** [参与者] 你的问题（例：`提问 analyst 主要风险是什么`）"
	}
	return "❌ " + err.Error() + "\nUsage: **ask** [participant] your question (e.g. `ask analyst what are the main risks`)"
}
