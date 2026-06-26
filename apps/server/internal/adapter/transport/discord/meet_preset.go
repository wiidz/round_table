package discord

import (
	"fmt"
	"strings"

	"round_table/apps/server/internal/domain/meeting"
)

type meetPreset struct {
	ID   string
	Icon string
	Name string
	Hint string
	Make func(topic string) meetLaunchConfig
}

func deliberationPresets(defaultCfg meetLaunchConfig, loc Locale) []meetPreset {
	return []meetPreset{
		{
			ID: "1", Icon: "⚡", Name: presetName(loc, "直接开始（默认）", "Start now (default)"),
			Hint: presetHint(loc, defaultCfg.MaxRounds, defaultCfg.Confirmation, 0, defaultCfg.Mode == meeting.MeetingModeDeliberation),
			Make: func(t string) meetLaunchConfig {
				cfg := defaultCfg
				cfg.Topic = t
				cfg.FreeDialogueQuestions = 0
				return cfg
			},
		},
		{
			ID: "2", Icon: "🌩️", Name: presetName(loc, "闪电研讨", "Flash"),
			Hint: presetHint(loc, 1, meeting.ConfirmationModeSkip, 0, true),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDeliberation, 1, meeting.ConfirmationModeSkip, 0)
			},
		},
		{
			ID: "3", Icon: "📐", Name: presetName(loc, "标准研讨", "Standard"),
			Hint: presetHint(loc, 3, meeting.ConfirmationModeSkip, 0, true),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDeliberation, 3, meeting.ConfirmationModeSkip, 0)
			},
		},
		{
			ID: "4", Icon: "💬", Name: presetName(loc, "研讨 + 自由对话", "Deliberation + Q&A"),
			Hint: presetHintFree(loc, 3, meeting.ConfirmationModeSkip, 1),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDeliberation, 3, meeting.ConfirmationModeSkip, 1)
			},
		},
		{
			ID: "5", Icon: "✅", Name: presetName(loc, "研讨 + 需确认", "Deliberation + review"),
			Hint: presetHint(loc, 3, meeting.ConfirmationModeRequired, 1, true),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDeliberation, 3, meeting.ConfirmationModeRequired, 1)
			},
		},
		{
			ID: "6", Icon: "🔬", Name: presetName(loc, "深度研讨", "Deep"),
			Hint: presetHint(loc, 5, meeting.ConfirmationModeRequired, 1, true),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDeliberation, 5, meeting.ConfirmationModeRequired, 1)
			},
		},
	}
}

func decisionPresets(loc Locale) []meetPreset {
	return []meetPreset{
		{
			ID: "J1", Icon: "🌩️", Name: presetName(loc, "闪电裁决", "Flash"),
			Hint: presetHint(loc, 1, meeting.ConfirmationModeSkip, 0, false),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDecision, 1, meeting.ConfirmationModeSkip, 0)
			},
		},
		{
			ID: "J2", Icon: "⚡", Name: presetName(loc, "快速裁决", "Quick"),
			Hint: presetHint(loc, 2, meeting.ConfirmationModeSkip, 0, false),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDecision, 2, meeting.ConfirmationModeSkip, 0)
			},
		},
		{
			ID: "J3", Icon: "📋", Name: presetName(loc, "标准裁决", "Standard"),
			Hint: presetHint(loc, 3, meeting.ConfirmationModeSkip, 0, false),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDecision, 3, meeting.ConfirmationModeSkip, 0)
			},
		},
		{
			ID: "J4", Icon: "✅", Name: presetName(loc, "裁决 + 需确认", "Decision + review"),
			Hint: presetHint(loc, 3, meeting.ConfirmationModeRequired, 1, false),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDecision, 3, meeting.ConfirmationModeRequired, 1)
			},
		},
		{
			ID: "J5", Icon: "🔬", Name: presetName(loc, "深度裁决", "Deep"),
			Hint: presetHint(loc, 5, meeting.ConfirmationModeRequired, 1, false),
			Make: func(t string) meetLaunchConfig {
				return presetLaunchConfig(t, meeting.MeetingModeDecision, 5, meeting.ConfirmationModeRequired, 1)
			},
		},
	}
}

func presetName(loc Locale, zh, en string) string {
	if loc == LocaleZH {
		return zh
	}
	return en
}

func presetHint(loc Locale, rounds int, confirmation string, free int, deliberation bool) string {
	if loc == LocaleZH {
		mode := "裁决"
		if deliberation {
			mode = "研讨"
		}
		freePart := " · 无自由对话"
		if free > 0 {
			freePart = fmt.Sprintf(" · 自由对话 %d 轮/人", free)
		}
		return fmt.Sprintf("%s · %d 轮 · 确认%s%s", mode, rounds, confirmationModeLabel(confirmation, loc), freePart)
	}
	mode := "decision"
	if deliberation {
		mode = "deliberation"
	}
	freePart := ""
	if free > 0 {
		freePart = fmt.Sprintf(" · free %d Q", free)
	}
	return fmt.Sprintf("%s · %d rounds · confirm %s%s", mode, rounds, confirmation, freePart)
}

func presetHintFree(loc Locale, rounds int, confirmation string, free int) string {
	if loc == LocaleZH {
		return fmt.Sprintf("研讨 · %d 轮 · 自由对话 %d 轮/人 · 确认%s", rounds, free, confirmationModeLabel(confirmation, loc))
	}
	return fmt.Sprintf("deliberation · %d rounds · free %d Q · confirm %s", rounds, free, confirmation)
}

func formatModeratorSetupPrompt(loc Locale, prefix string, defaultCfg meetLaunchConfig) string {
	delib := deliberationPresets(defaultCfg, loc)
	decide := decisionPresets(loc)
	var b strings.Builder

	if loc == LocaleZH {
		fmt.Fprintf(&b, "🎙️ **主持人** — 请选择会议方案\n\n📌 **主题：** %s\n", defaultCfg.Topic)
		b.WriteString("\n━━━━━━━━━━━━━━━━\n")
		b.WriteString("📋 **研讨** · 出方案草案\n")
		b.WriteString("回复 **1–6** 数字\n\n")
	} else {
		fmt.Fprintf(&b, "🎙️ **Moderator** — pick a preset\n\n📌 **Topic:** %s\n", defaultCfg.Topic)
		b.WriteString("\n━━━━━━━━━━━━━━━━\n")
		b.WriteString("📋 **Deliberation** · design draft\n")
		b.WriteString("Reply **1–6**\n\n")
	}

	for _, p := range delib {
		fmt.Fprintf(&b, "**%s** %s **%s**\n    └ %s\n\n", p.ID, p.Icon, p.Name, p.Hint)
	}

	if loc == LocaleZH {
		b.WriteString("━━━━━━━━━━━━━━━━\n")
		b.WriteString("⚖️ **裁决** · 投票拍板\n")
		b.WriteString("回复 **J1–J5**（字母 J + 数字）\n\n")
	} else {
		b.WriteString("━━━━━━━━━━━━━━━━\n")
		b.WriteString("⚖️ **Decision** · vote & conclude\n")
		b.WriteString("Reply **J1–J5** (letter J + number)\n\n")
	}

	for _, p := range decide {
		fmt.Fprintf(&b, "**%s** %s **%s**\n    └ %s\n\n", p.ID, p.Icon, p.Name, p.Hint)
	}

	if loc == LocaleZH {
		b.WriteString("━━━━━━━━━━━━━━━━\n")
		b.WriteString("**0** — 自定义（逐步配置）\n\n")
		b.WriteString("取消：发送 **取消会议**")
	} else {
		b.WriteString("━━━━━━━━━━━━━━━━\n")
		b.WriteString("**0** — Custom (step-by-step)\n\n")
		b.WriteString("Cancel: send **取消会议**")
	}

	_ = prefix
	return strings.TrimRight(b.String(), "\n")
}

func normalizePresetChoice(content string) string {
	s := strings.TrimSpace(content)
	if s == "" {
		return ""
	}
	s = normalizeASCIIForms(s)
	lower := strings.ToLower(s)
	if lower == "开始" || lower == "默认" || lower == "start" || lower == "default" || lower == "ok" {
		return "1"
	}
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ToUpper(s)
	if len(s) >= 2 && s[0] == 'J' {
		end := 1
		for end < len(s) && s[end] >= '0' && s[end] <= '9' {
			end++
		}
		if end > 1 {
			return s[:end]
		}
	}
	return s
}

func normalizeASCIIForms(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '０', '0':
			b.WriteByte('0')
		case '１', '1':
			b.WriteByte('1')
		case '２', '2':
			b.WriteByte('2')
		case '３', '3':
			b.WriteByte('3')
		case '４', '4':
			b.WriteByte('4')
		case '５', '5':
			b.WriteByte('5')
		case '６', '6':
			b.WriteByte('6')
		case '７', '7':
			b.WriteByte('7')
		case '８', '8':
			b.WriteByte('8')
		case '９', '9':
			b.WriteByte('9')
		case 'Ｊ', 'J', 'j':
			b.WriteByte('J')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func lookupPreset(choice string, defaultCfg meetLaunchConfig, loc Locale) (meetPreset, bool) {
	for _, p := range deliberationPresets(defaultCfg, loc) {
		if p.ID == choice {
			return p, true
		}
	}
	for _, p := range decisionPresets(loc) {
		if p.ID == choice {
			return p, true
		}
	}
	return meetPreset{}, false
}
