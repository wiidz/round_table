package discord

import (
	"fmt"
	"strings"

	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/platform/config"
)

type meetPreset struct {
	ID      string
	Name    string
	Hint    string
	Command string
	Make    func(topic string) meetLaunchConfig
}

func launchConfigFromPreset(p config.MeetPresetConfig, topic string) meetLaunchConfig {
	minRounds := 2
	if p.MaxRounds < minRounds {
		minRounds = p.MaxRounds
	}
	return meetLaunchConfig{
		Topic:                    topic,
		Mode:                     p.Mode,
		MaxRounds:                p.MaxRounds,
		MinRoundsBeforeSynthesis: minRounds,
		Confirmation:             p.Confirmation,
		FreeDialogueQuestions:    p.FreeDialogueQuestions,
	}
}

func buildMeetPresets(stored []config.MeetPresetConfig, loc Locale) []meetPreset {
	if len(stored) == 0 {
		stored = config.DefaultMeetPresets(config.Config{})
	}
	out := make([]meetPreset, 0, len(stored))
	for _, p := range stored {
		p := p
		out = append(out, meetPreset{
			ID:      p.ID,
			Name:    presetName(loc, config.PresetMenuNameZH(p), p.NameEN),
			Hint:    presetHint(loc, p.MaxRounds, p.Confirmation, p.FreeDialogueQuestions, p.Mode == meeting.MeetingModeDeliberation),
			Command: p.Command,
			Make: func(topic string) meetLaunchConfig {
				return launchConfigFromPreset(p, topic)
			},
		})
	}
	return out
}

func filterMeetPresets(all []meetPreset, group string) []meetPreset {
	out := make([]meetPreset, 0, len(all))
	for _, p := range all {
		if group == "deliberation" && !strings.HasPrefix(strings.ToUpper(p.ID), "J") {
			out = append(out, p)
		}
		if group == "decision" && strings.HasPrefix(strings.ToUpper(p.ID), "J") {
			out = append(out, p)
		}
	}
	return out
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

func formatModeratorSetupPrompt(loc Locale, prefix string, all []meetPreset) string {
	delib := filterMeetPresets(all, "deliberation")
	decide := filterMeetPresets(all, "decision")
	var b strings.Builder

	if loc == LocaleZH {
		b.WriteString("🎙️ **主持人** — 请选择会议方案\n\n")
		if len(delib) > 0 {
			b.WriteString("━━━━━━━━━━━━━━━━\n")
			b.WriteString("📋 **研讨** · 出方案草案\n")
			b.WriteString("回复 **1–6** 数字\n\n")
		}
	} else {
		b.WriteString("🎙️ **Moderator** — pick a preset\n\n")
		if len(delib) > 0 {
			b.WriteString("━━━━━━━━━━━━━━━━\n")
			b.WriteString("📋 **Deliberation** · design draft\n")
			b.WriteString("Reply **1–6**\n\n")
		}
	}

	for _, p := range delib {
		fmt.Fprintf(&b, "%s", formatPresetMenuLine(p))
	}

	if len(decide) > 0 {
		if loc == LocaleZH {
			b.WriteString("━━━━━━━━━━━━━━━━\n")
			b.WriteString("⚖️ **裁决** · 投票拍板\n")
			b.WriteString("回复 **J1–J5**（字母 J + 数字）\n\n")
		} else {
			b.WriteString("━━━━━━━━━━━━━━━━\n")
			b.WriteString("⚖️ **Decision** · vote & conclude\n")
			b.WriteString("Reply **J1–J5** (letter J + number)\n\n")
		}
	}

	for _, p := range decide {
		fmt.Fprintf(&b, "%s", formatPresetMenuLine(p))
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

func formatPresetMenuLine(p meetPreset) string {
	primary := p.ID
	if cmd := strings.TrimSpace(p.Command); cmd != "" {
		primary = cmd
	}
	return fmt.Sprintf("**%s** **%s**\n    └ %s\n\n", primary, p.Name, p.Hint)
}

func normalizePresetChoice(content string) string {
	s := strings.TrimSpace(content)
	if s == "" {
		return ""
	}
	return config.NormalizePresetCommandKey(s)
}

func lookupPreset(choice string, all []meetPreset) (meetPreset, bool) {
	key := normalizePresetChoice(choice)
	if key == "" {
		return meetPreset{}, false
	}
	for _, p := range all {
		if config.NormalizePresetCommandKey(p.Command) == key {
			return p, true
		}
	}
	return meetPreset{}, false
}

func normalizeASCIIForms(s string) string {
	return config.NormalizePresetASCIIForms(s)
}

func presetName(loc Locale, zh, en string) string {
	if loc == LocaleZH {
		return zh
	}
	return en
}
