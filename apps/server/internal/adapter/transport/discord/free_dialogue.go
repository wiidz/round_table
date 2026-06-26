package discord

import (
	"errors"
	"fmt"
	"strings"
)

var errFreeDialogueQuestionEmpty = errors.New("请在「提问」后输入问题内容")

func isFreeDialogueQuestionTrigger(content string) bool {
	_, ok := parseFreeDialogueQuestion(content)
	return ok
}

func parseFreeDialogueQuestion(content string) (question string, ok bool) {
	s := strings.TrimSpace(content)
	if s == "" {
		return "", false
	}
	norm := normalizeASCIIForms(s)
	lower := strings.ToLower(norm)

	for _, prefix := range []string{"提问", "ask", "question"} {
		pl := strings.ToLower(prefix)
		if lower == pl {
			return "", true // trigger without body — caller validates
		}
		if strings.HasPrefix(lower, pl) {
			rest := strings.TrimSpace(s[len(prefix):])
			rest = strings.TrimPrefix(rest, "：")
			rest = strings.TrimPrefix(rest, ":")
			rest = strings.TrimSpace(rest)
			if rest != "" {
				return rest, true
			}
			return "", true
		}
	}
	return "", false
}

func freeDialogueQuestionAckText(loc Locale, question string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("💬 已收到提问，将在当前发言结束后发出：\n> %s", question)
	}
	return fmt.Sprintf("💬 Question queued — will ask after the current turn:\n> %s", question)
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
		return "❌ " + err.Error() + "\n用法：**提问** 你的问题"
	}
	return "❌ " + err.Error() + "\nUsage: **ask** your question"
}
