package discord

import "fmt"

func (h *CommandHandler) locale() Locale {
	if h.Meet != nil {
		return h.Meet.locale()
	}
	if h.Participants != nil && h.Participants.Locale != nil {
		return h.Participants.Locale()
	}
	return LocaleEN
}

func (h *CommandHandler) helpText() string {
	p := h.Prefix
	if h.locale() == LocaleZH {
		return fmt.Sprintf(`📖 **RoundTable Discord 指令**

前缀：`+"`%s`"+`

- `+"`会议状态`"+` / `+"`状态`"+` — 查看当前频道输入态与可接受指令
- `+"`%sstatus`"+` — 同上
- `+"`%shelp`"+` — 显示帮助
- `+"`%sprincipal bind`"+` — 绑定本范围 Principal（每服务器/私信一位）
- `+"`%sprincipal whoami`"+` — 查看 Principal 绑定
- `+"`%sprincipal unbind`"+` — 解除 Principal 绑定
- `+"`开始会议`"+` / `+"`新会议`"+` / `+"`会议开始`"+` — 发起会议（无需前缀，主持人逐步引导）
- `+"`取消会议`"+` — 取消待确认的会议配置
- **会议进行中**：`+"`暂停会议`"+` · `+"`恢复会议`"+` · `+"`终止会议`"+` · `+"`立即合成`"+`（研讨）· `+"`强制共识`"+`（裁决）
- **自由问答阶段**：`+"`提问 …`"+` / `+"`提问 designer …`"+` — 指定参与者提问
- **会议结束后**：`+"`获取纪要`"+` · `+"`获取草案`"+` · `+"`获取待决`"+` · `+"`获取结论`"+`
- `+"`%smeet [-mode decision|deliberation] 主题`"+` — 同上（带主题时可跳过主题输入）
- `+"`%smeet cancel`"+` — 取消待确认的会议配置
- `+"`%sexpert list`"+` / `+"`%s专家 列表`"+` — 查看专家名录
- `+"`%sexpert new`"+` / `+"`%s专家 新建`"+` — 新建专家（逐步引导）
- `+"`%sexpert edit <代号>`"+` / `+"`%sexpert delete <代号>`"+` — 编辑或删除专家`, p, p, p, p, p, p, p, p, p, p, p, p, p, p)
	}
	return fmt.Sprintf(`📖 **RoundTable Discord commands**

Prefix: `+"`%s`"+`

- `+"`会议状态`"+` / `+"`status`"+` — Show current input phase and accepted commands
- `+"`%sstatus`"+` — Same
- `+"`%shelp`"+` — Show this help
- `+"`%sprincipal bind`"+` — Bind yourself as Principal (one per server/DM)
- `+"`%sprincipal whoami`"+` — Show Principal binding
- `+"`%sprincipal unbind`"+` — Remove Principal binding
- `+"`开始会议`"+` / `+"`新会议`"+` / `+"`会议开始`"+` — Start a meeting (no prefix; Moderator guides you)
- `+"`取消会议`"+` — Cancel pending meet setup
- **While meeting runs:** `+"`暂停会议`"+` · `+"`恢复会议`"+` · `+"`终止会议`"+` · `+"`立即合成`"+` (deliberation) · `+"`强制共识`"+` (decision)
- **During free dialogue:** `+"`ask …`"+` / `+"`提问 designer …`"+` — ask a participant
- **After meeting:** `+"`get minutes`"+` · `+"`get draft`"+` · `+"`get open`"+` · `+"`get conclusion`"+`
- `+"`%smeet [-mode decision|deliberation] topic`"+` — Same (topic inline skips topic prompt)
- `+"`%smeet cancel`"+` — Cancel pending meet setup
- `+"`%sexpert list`"+` — List expert roster
- `+"`%sexpert new`"+` — Create expert (guided wizard)
- `+"`%sexpert edit <id>`"+` / `+"`%sexpert delete <id>`"+` — Update or remove expert`, p, p, p, p, p, p, p, p, p, p, p, p)
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
		return fmt.Sprintf("用法：`%smeet [-mode decision|deliberation] 会议主题`\n主持人会展示研讨 **1–6**、裁决 **J1–J5** 选项；**0** 进入自定义。", prefix)
	}
	return fmt.Sprintf("Usage: `%smeet [-mode decision|deliberation] topic`\nModerator shows deliberation **1–6**, decision **J1–J5**; **0** for custom.", prefix)
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
		return "⏳ 本频道已有待确认的会议配置。研讨 **1–6**，裁决 **J1–J5**，或 **0** 自定义；**取消会议** 取消。"
	}
	return "⏳ Setup pending. Deliberation **1–6**, decision **J1–J5**, **0** custom; **取消会议** to cancel."
}

func meetTopicRequiredText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 请输入会议主题（直接发送文字即可）。"
	}
	return "⚠️ Please send the meeting topic as plain text."
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
		return "❌ " + err.Error() + "\n研讨 **1–6** · 裁决 **J1–J5** · **0** 自定义 · 自定义步骤中 **0** 返回"
	}
	return "❌ " + err.Error() + "\nDeliberation **1–6** · decision **J1–J5** · **0** custom · **0** back in wizard"
}

func meetParticipantsPickErrorText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ " + err.Error() + "\n请发 **阵容编号**、**专家编号/名字**（逗号分隔），或 **0** 全员。"
	}
	return "❌ " + err.Error() + "\nSend **cast id**, **participant index/name** (comma-separated), or **0** for all."
}
