package discord

import "fmt"

func (h *CommandHandler) locale() Locale {
	if h.Meet != nil {
		return ParseLocale(h.Meet.Discord.Locale)
	}
	return LocaleEN
}

func (h *CommandHandler) helpText() string {
	p := h.Prefix
	if h.locale() == LocaleZH {
		return fmt.Sprintf(`📖 **RoundTable Discord 指令**

前缀：`+"`%s`"+`

- `+"`%shelp`"+` — 显示帮助
- `+"`%sprincipal bind`"+` — 绑定本范围 Principal（每服务器/私信一位）
- `+"`%sprincipal whoami`"+` — 查看 Principal 绑定
- `+"`%sprincipal unbind`"+` — 解除 Principal 绑定
- `+"`%smeet [-mode decision|deliberation] 主题`"+` — 发起会议；主持人列出数字选项，回复 **1–5** 即可
- `+"`%smeet cancel`"+` — 取消待确认的会议配置`, p, p, p, p, p, p, p)
	}
	return fmt.Sprintf(`📖 **RoundTable Discord commands**

Prefix: `+"`%s`"+`

- `+"`%shelp`"+` — Show this help
- `+"`%sprincipal bind`"+` — Bind yourself as Principal (one per server/DM)
- `+"`%sprincipal whoami`"+` — Show Principal binding
- `+"`%sprincipal unbind`"+` — Remove Principal binding
- `+"`%smeet [-mode decision|deliberation] topic`"+` — Start a meeting; Moderator shows numbered options — reply **1–5**
- `+"`%smeet cancel`"+` — Cancel pending meet setup`, p, p, p, p, p, p, p)
}

func unknownCommandText(loc Locale, prefix, cmd string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("❓ 未知指令 `%s`。发送 `%shelp` 查看用法。", cmd, prefix)
	}
	return fmt.Sprintf("❓ Unknown command `%s`. Send `%shelp` for usage.", cmd, prefix)
}

func principalUsageText(loc Locale, prefix string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("用法：`%sprincipal bind|whoami|unbind`", prefix)
	}
	return fmt.Sprintf("Usage: `%sprincipal bind|whoami|unbind`", prefix)
}

func principalBindFailedText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ 绑定失败：" + err.Error()
	}
	return "❌ Bind failed: " + err.Error()
}

func principalBindOKText(loc Locale, principalID, displayName, scopeLabel string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("✅ 已绑定 Principal\n- 🆔 `%s`\n- 👤 显示名：%s\n- 📍 范围：%s", principalID, displayName, scopeLabel)
	}
	return fmt.Sprintf("✅ Principal bound\n- 🆔 `%s`\n- 👤 Display: %s\n- 📍 Scope: %s", principalID, displayName, scopeLabel)
}

func scopeLabel(loc Locale, isDM bool) string {
	if loc == LocaleZH {
		if isDM {
			return "你的私信会话"
		}
		return "本服务器"
	}
	if isDM {
		return "this DM thread"
	}
	return "this server"
}

func principalNotBoundText(loc Locale, prefix string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("ℹ️ 当前范围尚未绑定 Principal。发送 `%sprincipal bind` 绑定。", prefix)
	}
	return fmt.Sprintf("ℹ️ No Principal bound in this scope. Send `%sprincipal bind` to bind.", prefix)
}

func principalWhoamiSelfText(loc Locale, principalID, displayName, boundAt string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("👤 你是本范围的 Principal\n- 🆔 `%s`\n- 显示名：%s\n- 绑定于：%s", principalID, displayName, boundAt)
	}
	return fmt.Sprintf("👤 You are the Principal here\n- 🆔 `%s`\n- Display: %s\n- Bound at: %s", principalID, displayName, boundAt)
}

func principalWhoamiOtherText(loc Locale, displayName, principalID string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("👤 本范围 Principal：**%s** (`%s`)", displayName, principalID)
	}
	return fmt.Sprintf("👤 Principal here: **%s** (`%s`)", displayName, principalID)
}

func principalUnbindFailedText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ 解绑失败：" + err.Error()
	}
	return "❌ Unbind failed: " + err.Error()
}

func principalUnbindOKText(loc Locale) string {
	if loc == LocaleZH {
		return "✅ 已解除 Principal 绑定。"
	}
	return "✅ Principal binding removed."
}

func meetDisabledText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 会议功能未启用。"
	}
	return "⚠️ Meeting feature is not enabled."
}

func meetUsageText(loc Locale, prefix string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("用法：`%smeet [-mode decision|deliberation] 会议主题`\n主持人会给出 **1–5** 数字选项；回复一个数字即可开始或进入自定义。", prefix)
	}
	return fmt.Sprintf("Usage: `%smeet [-mode decision|deliberation] topic`\nModerator shows **1–5** options; reply with one number to start or customize.", prefix)
}

func meetNeedBindText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 请先绑定 Principal，再发起会议。"
	}
	return "⚠️ Bind as Principal first, then start a meeting."
}

func meetNotScopePrincipalText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 只有本范围的 Principal 可以发起会议。"
	}
	return "⚠️ Only the bound Principal in this scope can start a meeting."
}

func meetChannelBusyText(loc Locale, meetingID string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("⏳ 本频道已有会议进行中（`%s`）。", meetingID)
	}
	return fmt.Sprintf("⏳ A meeting is already running in this channel (`%s`).", meetingID)
}

func meetEngineFailedText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ 会议启动失败：" + err.Error()
	}
	return "❌ Failed to start meeting: " + err.Error()
}

func meetConfigErrorText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ 会议配置错误：" + err.Error()
	}
	return "❌ Meeting config error: " + err.Error()
}

func meetCreateFailedText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ 创建会议失败：" + err.Error()
	}
	return "❌ Failed to create meeting: " + err.Error()
}

func meetRunFailedText(loc Locale, meetingID string, err error) string {
	if loc == LocaleZH {
		return fmt.Sprintf("❌ 会议 `%s` 失败：%v", meetingID, err)
	}
	return fmt.Sprintf("❌ Meeting `%s` failed: %v", meetingID, err)
}

func meetSetupPendingText(loc Locale) string {
	if loc == LocaleZH {
		return "⏳ 本频道已有待确认的会议配置。回复 **1–5** 选择，或 `!rt meet cancel` 取消。"
	}
	return "⏳ Meet setup is pending. Reply **1–5**, or `!rt meet cancel` to cancel."
}

func meetSetupCancelledText(loc Locale) string {
	if loc == LocaleZH {
		return "🛑 已取消会议配置。"
	}
	return "🛑 Meet setup cancelled."
}

func meetSetupNothingToCancelText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 当前没有待确认的会议配置。"
	}
	return "ℹ️ No pending meet setup."
}

func meetSetupNotOwnerText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 只有发起会议的 Principal 可以确认或调整配置。"
	}
	return "⚠️ Only the Principal who started setup can confirm or adjust."
}

func meetSetupParseErrorText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ " + err.Error() + "\n请回复菜单上的 **数字**（如 `1`），自定义步骤中 `0` 返回上一级。"
	}
	return "❌ " + err.Error() + "\nReply with a menu **number** (e.g. `1`); use `0` to go back during custom setup."
}
