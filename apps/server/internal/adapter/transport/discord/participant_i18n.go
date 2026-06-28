package discord

import (
	"fmt"
	"strings"

	"round_table/apps/server/internal/platform/config"
)

func expertUsageText(loc Locale, prefix string) string {
	if loc == LocaleZH {
		return fmt.Sprintf(`📖 **专家管理**

- `+"`%sexpert list`"+` / `+"`%s专家 列表`"+` — 查看名录
- `+"`%sexpert show <代号|名称>`"+` — 查看详情
- `+"`%sexpert new`"+` / `+"`%s专家 新建`"+` — 新建专家（逐步引导）
- `+"`%sexpert edit <代号|名称>`"+` — 编辑名称/专长/代号
- `+"`%sexpert delete <代号|名称>`"+` — 删除专家
- `+"`%sexpert cancel`"+` / `+"`取消专家`"+` — 取消进行中的向导`, prefix, prefix, prefix, prefix, prefix, prefix, prefix, prefix)
	}
	return fmt.Sprintf(`📖 **Expert management**

- `+"`%sexpert list`"+` — list roster
- `+"`%sexpert show <id|name>`"+` — show details
- `+"`%sexpert new`"+` — create expert (guided)
- `+"`%sexpert edit <id|name>`"+` — edit display/expertise/id
- `+"`%sexpert delete <id|name>`"+` — delete expert
- `+"`%sexpert cancel`"+` — cancel pending wizard`, prefix, prefix, prefix, prefix, prefix, prefix)
}

func expertStorageRequiredText(loc Locale) string {
	if loc == LocaleZH {
		return "❌ 专家管理需要 SQLite 设置存储（与 Web 控制台相同）。"
	}
	return "❌ Expert management requires SQLite settings storage (same as Web console)."
}

func expertRefRequiredText(loc Locale) string {
	if loc == LocaleZH {
		return "❌ 请指定专家代号或显示名。"
	}
	return "❌ Expert id or display name required."
}

func expertErrorText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ " + err.Error()
	}
	return "❌ " + err.Error()
}

func expertProfileErrorText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ 档案操作失败：" + err.Error()
	}
	return "❌ Profile operation failed: " + err.Error()
}

func expertListEmptyText(loc Locale) string {
	if loc == LocaleZH {
		return "📋 当前没有配置专家。发送 `!rt 专家 新建` 创建。"
	}
	return "📋 No experts configured. Send `!rt expert new` to create one."
}

func expertListFooterText(loc Locale, prefix string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("新建：`%sexpert new` · 详情：`%sexpert show <代号>`", prefix, prefix)
	}
	return fmt.Sprintf("Create: `%sexpert new` · Details: `%sexpert show <id>`", prefix, prefix)
}

func formatExpertAskDisplayName(loc Locale) string {
	if loc == LocaleZH {
		return `🧑‍💼 **新建专家 · 1/4 名称**

请发送 **显示名**（如「LOL 玩家代表」）。
取消：发送 **取消专家**`
	}
	return `🧑‍💼 **New expert · 1/4 display name**

Send the **display name** (e.g. LOL player advocate).
Cancel: send **cancel expert**`
}

func formatExpertAskID(loc Locale, suggested string) string {
	if loc == LocaleZH {
		return fmt.Sprintf(`🧑‍💼 **新建专家 · 2/4 代号**

建议代号：`+"`%s`"+`

请发送代号（小写英文，如 `+"`player_lol`"+`），或发 **自动** / **-%s** 使用建议。
取消：发送 **取消专家**`, suggested, suggested)
	}
	return fmt.Sprintf(`🧑‍💼 **New expert · 2/4 id**

Suggested: `+"`%s`"+`

Send codename (lowercase, e.g. `+"`player_lol`"+`) or **auto** / send suggested id as-is.
Cancel: send **cancel expert**`, suggested)
}

func formatExpertAskExpertise(loc Locale) string {
	if loc == LocaleZH {
		return `🧑‍💼 **新建专家 · 3/4 专长**

请发送 **专长标签**（如 gameplay、玩家体验），或 **跳过** 使用默认。
取消：发送 **取消专家**`
	}
	return `🧑‍💼 **New expert · 3/4 expertise**

Send an **expertise tag** (e.g. gameplay) or **skip** for default.
Cancel: send **cancel expert**`
}

func formatExpertConfirmCreate(loc Locale, item config.ParticipantRosterItem) string {
	exp := strings.TrimSpace(item.Expertise)
	if exp == "" {
		exp = defaultExpertiseTag
	}
	if loc == LocaleZH {
		return fmt.Sprintf(`🧑‍💼 **新建专家 · 确认**

- 代号：`+"`%s`"+`
- 名称：%s
- 专长：%s

将创建 roster 条目，并从模板生成 SOUL/AGENTS 档案草稿。

**1** — 确认创建 · **0** — 取消`, item.ID, item.DisplayName, exp)
	}
	return fmt.Sprintf(`🧑‍💼 **New expert · confirm**

- Id: `+"`%s`"+`
- Display: %s
- Expertise: %s

Creates roster entry and template SOUL/AGENTS profile.

**1** — create · **0** — cancel`, item.ID, item.DisplayName, exp)
}

func formatExpertAskEditFields(loc Locale, item config.ParticipantRosterItem) string {
	exp := strings.TrimSpace(item.Expertise)
	if exp == "" {
		exp = defaultExpertiseTag
	}
	if loc == LocaleZH {
		return fmt.Sprintf(`✏️ **编辑专家** `+"`%s`"+`

当前：名称=%s · 专长=%s · 代号=%s

请一行发送修改（示例：`+"`名称=新名称, 专长=pvp`"+`）。
可改字段：**名称** / **专长** / **代号**（无需改的字段可省略）。
取消：发送 **取消专家**`, item.ID, item.DisplayName, exp, item.ID)
	}
	return fmt.Sprintf(`✏️ **Edit expert** `+"`%s`"+`

Current: name=%s · expertise=%s · id=%s

Send updates in one line (e.g. `+"`name=New Name, expertise=pvp`"+`).
Fields: **name** / **expertise** / **id**.
Cancel: send **cancel expert**`, item.ID, item.DisplayName, exp, item.ID)
}

func formatExpertConfirmEdit(loc Locale, oldID string, item config.ParticipantRosterItem) string {
	exp := strings.TrimSpace(item.Expertise)
	if exp == "" {
		exp = defaultExpertiseTag
	}
	if loc == LocaleZH {
		return fmt.Sprintf(`✏️ **编辑专家 · 确认**

- 原代号：`+"`%s`"+`
- 新代号：`+"`%s`"+`
- 名称：%s
- 专长：%s

**1** — 确认保存 · **0** — 取消`, oldID, item.ID, item.DisplayName, exp)
	}
	return fmt.Sprintf(`✏️ **Edit expert · confirm**

- Old id: `+"`%s`"+`
- New id: `+"`%s`"+`
- Display: %s
- Expertise: %s

**1** — save · **0** — cancel`, oldID, item.ID, item.DisplayName, exp)
}

func formatExpertConfirmDelete(loc Locale, item config.ParticipantRosterItem) string {
	if loc == LocaleZH {
		return fmt.Sprintf(`⚠️ **删除专家**

确认删除 `+"`%s`"+` · %s ？

将从会议名录移除，并删除本地档案目录（不可恢复）。

**1** — 确认删除 · **0** — 取消`, item.ID, item.DisplayName)
	}
	return fmt.Sprintf(`⚠️ **Delete expert**

Delete `+"`%s`"+` · %s ?

Removes roster entry and local profile directory (irreversible).

**1** — delete · **0** — cancel`, item.ID, item.DisplayName)
}

func formatExpertCreated(loc Locale, item config.ParticipantRosterItem) string {
	if loc == LocaleZH {
		return fmt.Sprintf("✅ 已创建专家 `%s` · %s\n\n可在 Web 控制台绑定 Discord Bot、编辑 SOUL/AGENTS。", item.ID, item.DisplayName)
	}
		return fmt.Sprintf("✅ Created expert `%s` · %s\n\nBind Discord bot and edit SOUL/AGENTS in Web console.", item.ID, item.DisplayName)
}

func formatExpertUpdated(loc Locale, oldID string, item config.ParticipantRosterItem) string {
	if oldID != item.ID {
		if loc == LocaleZH {
			return fmt.Sprintf("✅ 已更新专家：`%s` → `%s` · %s", oldID, item.ID, item.DisplayName)
		}
		return fmt.Sprintf("✅ Expert updated: `%s` → `%s` · %s", oldID, item.ID, item.DisplayName)
	}
	if loc == LocaleZH {
		return fmt.Sprintf("✅ 已更新专家 `%s` · %s", item.ID, item.DisplayName)
	}
	return fmt.Sprintf("✅ Expert updated `%s` · %s", item.ID, item.DisplayName)
}

func formatExpertDeleted(loc Locale, item config.ParticipantRosterItem) string {
	if loc == LocaleZH {
		return fmt.Sprintf("✅ 已删除专家 `%s` · %s", item.ID, item.DisplayName)
	}
	return fmt.Sprintf("✅ Deleted expert `%s` · %s", item.ID, item.DisplayName)
}

func expertSetupBusyText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 本频道已有进行中的专家向导。发送 **取消专家** 可放弃。"
	}
	return "ℹ️ An expert wizard is already pending in this channel. Send **cancel expert** to abort."
}

func expertSetupCancelledText(loc Locale) string {
	if loc == LocaleZH {
		return "✅ 已取消专家操作。"
	}
	return "✅ Expert operation cancelled."
}

func expertSetupNotOwnerText(loc Locale) string {
	if loc == LocaleZH {
		return "❌ 只有发起该向导的用户可以继续或取消。"
	}
	return "❌ Only the user who started this wizard may continue or cancel."
}

func expertNothingToCancelText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 当前没有进行中的专家向导。"
	}
	return "ℹ️ No expert wizard in progress."
}

func expertDisplayNameRequiredText(loc Locale) string {
	if loc == LocaleZH {
		return "❌ 名称不能为空。"
	}
	return "❌ Display name required."
}

func expertIDTakenText(loc Locale, id string) string {
	if loc == LocaleZH {
		return fmt.Sprintf("❌ 代号 %q 已被占用，请换一个。", id)
	}
	return fmt.Sprintf("❌ Id %q is already taken.", id)
}

func expertConfirmChoiceText(loc Locale) string {
	if loc == LocaleZH {
		return "请回复 **1** 确认，或 **0** 取消。"
	}
	return "Reply **1** to confirm or **0** to cancel."
}
