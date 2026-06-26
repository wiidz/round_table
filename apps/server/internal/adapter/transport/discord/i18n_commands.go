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
- `+"`%smeet [-mode decision|deliberation] 主题`"+` — 发起会议（仅 Principal，每频道一场）`, p, p, p, p, p, p)
	}
	return fmt.Sprintf(`📖 **RoundTable Discord commands**

Prefix: `+"`%s`"+`

- `+"`%shelp`"+` — Show this help
- `+"`%sprincipal bind`"+` — Bind yourself as Principal (one per server/DM)
- `+"`%sprincipal whoami`"+` — Show Principal binding
- `+"`%sprincipal unbind`"+` — Remove Principal binding
- `+"`%smeet [-mode decision|deliberation] topic`"+` — Start a meeting (Principal only, one per channel)`, p, p, p, p, p, p)
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
		return fmt.Sprintf("用法：`%smeet [-mode decision|deliberation] 会议主题`", prefix)
	}
	return fmt.Sprintf("Usage: `%smeet [-mode decision|deliberation] topic`", prefix)
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
