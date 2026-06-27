package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type settingField struct {
	Key              string
	Label            string
	Group            string
	Subsection       string // optional second level, e.g. transport → discord
	Section          string // optional third level, e.g. meeting → 默认会议模式
	Secret           bool
	Editable         bool
	RestartRequired  bool
	Placeholder      string
	Description      string
	InputType        string // "", "number", "select", "switch", "radio"
	Options          []SettingOption
	Min              *int
	Max              *int
	apply            func(*Config, string) error
	read             func(Config) string
	secretConfigured func() bool
}

// SettingOption is one choice for select-style settings fields.
type SettingOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

const (
	groupService   = "服务"
	groupStorage   = "存储"
	groupLLM       = "LLM"
	groupMeeting   = "会议"
	groupTransport = "IM"

	meetingSectionLimits   = "设定的上限"

	subDiscord   = "discord"
	subTelegram  = "telegram"
	subSlack     = "slack"
	subDeepSeek  = "deepseek"
	subOpenAI    = "openai"
	subAnthropic = "anthropic"
	subQwen      = "qwen"
	subGemini    = "gemini"
	subGrok      = "grok"
	subOllama    = "ollama"
	subFeishu    = "feishu"
	subWhatsApp  = "whatsapp"
	subLark      = "lark"
	subIMessage  = "imessage"
)

func intSettingBound(n int) *int {
	return &n
}

var localeOptions = []SettingOption{
	{Value: "zh", Label: "中文"},
	{Value: "en", Label: "English"},
}

func parseConfirmationModeSetting(v string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "required", "true", "1", "yes", "on":
		return "required", nil
	case "skip", "false", "0", "no", "off":
		return "skip", nil
	default:
		return "", fmt.Errorf("确认关：请填写 required 或 skip")
	}
}

func parseIntSettingRange(v string, min, max int, label string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0, fmt.Errorf("%s：请输入整数", label)
	}
	if n < min || n > max {
		return 0, fmt.Errorf("%s：需在 %d–%d 之间", label, min, max)
	}
	return n, nil
}

func parseOneOfSetting(v string, allowed []string, label string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	for _, a := range allowed {
		if v == a {
			return v, nil
		}
	}
	return "", fmt.Errorf("%s：请填写 %s", label, strings.Join(allowed, " 或 "))
}

// SettingsSubsectionMeta describes a second-level tab within a group.
type SettingsSubsectionMeta struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Available bool   `json:"available"`
}

var groupSubsections = map[string][]SettingsSubsectionMeta{
	groupLLM: {
		{ID: subDeepSeek, Label: "DeepSeek", Available: true},
		{ID: subOpenAI, Label: "OpenAI", Available: true},
		{ID: subAnthropic, Label: "Anthropic", Available: true},
		{ID: subQwen, Label: "Qwen", Available: false},
		{ID: subGemini, Label: "Gemini", Available: false},
		{ID: subGrok, Label: "Grok", Available: false},
		{ID: subOllama, Label: "Ollama", Available: false},
	},
	groupTransport: {
		{ID: subDiscord, Label: "Discord", Available: true},
		{ID: subTelegram, Label: "Telegram", Available: false},
		{ID: subSlack, Label: "Slack", Available: false},
		{ID: subFeishu, Label: "飞书", Available: false},
		{ID: subWhatsApp, Label: "WhatsApp", Available: false},
		{ID: subLark, Label: "Lark", Available: false},
		{ID: subIMessage, Label: "iMessage", Available: false},
	},
}

var settingFields = []settingField{
	{
		Key: "ROUND_TABLE_ADDR", Label: "HTTP 监听地址", Group: groupService,
		Placeholder: ":7777", Description: "修改后需重启 roundtable-server",
		RestartRequired: true,
		read:            func(c Config) string { return c.Server.Addr },
	},
	{
		Key: "ROUND_TABLE_LOCALE", Label: "界面语言", Group: groupService,
		Editable: true, InputType: "radio", Options: localeOptions,
		Placeholder: "zh", Description: "影响 Web 与 Discord 等面向用户的文案",
		apply: func(c *Config, v string) error {
			loc, err := parseOneOfSetting(v, []string{"zh", "en"}, "界面语言")
			if err != nil {
				return err
			}
			c.Server.Locale = loc
			return nil
		},
		read: func(c Config) string {
			if loc := strings.TrimSpace(c.Server.Locale); loc != "" {
				return loc
			}
			return c.Transport.Discord.Locale
		},
	},
	{
		Key: "ROUND_TABLE_STORAGE_DRIVER", Label: "存储驱动", Group: groupStorage,
		Placeholder: "sqlite", Description: "修改后需重启服务",
		RestartRequired: true,
		read:            func(c Config) string { return c.Storage.Driver },
	},
	{
		Key: "ROUND_TABLE_STORAGE_SQLITE_PATH", Label: "SQLite 路径", Group: groupStorage,
		Placeholder: "./data/roundtable.db", RestartRequired: true,
		read: func(c Config) string { return c.Storage.SQLitePath },
	},
	{
		Key: "DEEPSEEK_MODEL_NAME", Label: "DeepSeek 模型", Group: groupLLM, Subsection: subDeepSeek,
		Editable: true, Placeholder: "deepseek-chat",
		apply: func(c *Config, v string) error {
			c.Model.DefaultModel = v
			return nil
		},
		read: func(c Config) string { return c.Model.DefaultModel },
	},
	{
		Key: "ROUND_TABLE_MODEL_BASE_URL", Label: "模型 API Base URL", Group: groupLLM, Subsection: subDeepSeek,
		Editable: true, Placeholder: "https://api.deepseek.com/v1",
		apply: func(c *Config, v string) error {
			c.Model.BaseURL = v
			return nil
		},
		read: func(c Config) string { return c.Model.BaseURL },
	},
	{
		Key: "ROUND_TABLE_MODEL_PROVIDER", Label: "模型 Provider", Group: groupLLM, Subsection: subDeepSeek,
		Editable: true, Placeholder: "deepseek",
		apply: func(c *Config, v string) error {
			c.Model.Provider = v
			return nil
		},
		read: func(c Config) string { return c.Model.Provider },
	},
	{
		Key: "ROUND_TABLE_MODEL_TIMEOUT_SEC", Label: "模型超时（秒）", Group: groupLLM, Subsection: subDeepSeek,
		Editable: true, Placeholder: "120",
		apply: func(c *Config, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				return fmt.Errorf("invalid timeout: %q", v)
			}
			c.Model.TimeoutSec = n
			return nil
		},
		read: func(c Config) string { return strconv.Itoa(c.Model.TimeoutSec) },
	},
	{
		Key: "DEEPSEEK_API_KEY", Label: "DeepSeek API Key", Group: groupLLM, Subsection: subDeepSeek,
		Secret: true, Description: "在 deploy/.env 配置，修改后需重启服务",
		secretConfigured: func() bool { return os.Getenv("DEEPSEEK_API_KEY") != "" },
	},
	{
		Key: "OPENAI_API_KEY", Label: "OpenAI API Key", Group: groupLLM, Subsection: subOpenAI, Secret: true,
		Description: "在 deploy/.env 配置，修改后需重启服务",
		secretConfigured: func() bool { return os.Getenv("OPENAI_API_KEY") != "" },
	},
	{
		Key: "ANTHROPIC_API_KEY", Label: "Anthropic API Key", Group: groupLLM, Subsection: subAnthropic, Secret: true,
		Description: "在 deploy/.env 配置，修改后需重启服务",
		secretConfigured: func() bool { return os.Getenv("ANTHROPIC_API_KEY") != "" },
	},
	{
		Key: "ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT", Label: "辩论轮次上限", Group: groupMeeting,
		Section: meetingSectionLimits,
		Editable: true, InputType: "number", Min: intSettingBound(1), Max: intSettingBound(20),
		Placeholder: "5",
		Description: "单个 Running Segment 内辩论轮（Round 1+）的硬上限，不含 Pre-meeting Round 0。\n\n预设菜单中单场轮次不可超过此值；达上限仍未 Consensus / 合成就绪时按 ADR-0005 / ADR-0011 兜底。",
		apply: func(c *Config, v string) error {
			n, err := parseIntSettingRange(v, 1, 20, "辩论轮次上限")
			if err != nil {
				return err
			}
			c.Meeting.MaxRoundsPerSegment = n
			return nil
		},
		read: func(c Config) string { return strconv.Itoa(c.Meeting.MaxRoundsPerSegment) },
	},
	{
		Key: "ROUND_TABLE_MIN_ROUNDS_BEFORE_SYNTHESIS", Label: "合成就绪最少轮数", Group: groupMeeting,
		Section: meetingSectionLimits,
		Editable: true, InputType: "number", Min: intSettingBound(1), Max: intSettingBound(20),
		Placeholder: "2",
		Description: "研讨型（deliberation）会议：至少进行几轮辩论后，Moderator 才会运行合成就绪检测（DeliberationReadinessChecked）。\n\n通常应 ≤ 辩论轮次上限。",
		apply: func(c *Config, v string) error {
			n, err := parseIntSettingRange(v, 1, 20, "合成就绪最少轮数")
			if err != nil {
				return err
			}
			c.Meeting.MinRoundsBeforeSynthesis = n
			return nil
		},
		read: func(c Config) string { return strconv.Itoa(c.Meeting.MinRoundsBeforeSynthesis) },
	},
	{
		Key: "ROUND_TABLE_MAX_CONFIRMATION_CYCLES", Label: "确认关轮次上限", Group: groupMeeting,
		Section: meetingSectionLimits,
		Editable: true, InputType: "number", Min: intSettingBound(1), Max: intSettingBound(10),
		Placeholder: "3",
		Description: "仅当预设或会议开启「需确认」时生效。\n\n每场会议 Principal 审阅合成方案的次数上限（含首次呈报，不是单独的「驳回计数」）。\n\nPrincipal 驳回后会继续讨论并再次进入确认；若在第 N 次确认仍驳回（N = 本上限），须在三项中选择：强制批准、继续研讨或中止会议。\n\n例：上限 3 → 最多 3 次确认呈报；第 1、2 次驳回可继续讨论，第 3 次仍驳回则触发兜底选择。",
		apply: func(c *Config, v string) error {
			n, err := parseIntSettingRange(v, 1, 10, "确认关轮次上限")
			if err != nil {
				return err
			}
			c.Meeting.MaxConfirmationCycles = n
			return nil
		},
		read: func(c Config) string { return strconv.Itoa(c.Meeting.MaxConfirmationCycles) },
	},
	{
		Key: "ROUND_TABLE_DISCORD_GUILD_ID", Label: "Guild ID", Group: groupTransport, Subsection: subDiscord,
		Editable: true, Placeholder: "留空表示不限制服务器",
		apply: func(c *Config, v string) error {
			c.Transport.Discord.GuildID = v
			return nil
		},
		read: func(c Config) string { return c.Transport.Discord.GuildID },
	},
	{
		Key: "ROUND_TABLE_DISCORD_AUTO_START", Label: "自动启动", Group: groupTransport, Subsection: subDiscord,
		Editable: true, RestartRequired: true,
		Description: "开启后，启动 roundtable 主服务时会自动拉起 Discord transport",
		apply: func(c *Config, v string) error {
			c.Transport.Discord.AutoStart = parseBoolSetting(v)
			return nil
		},
		read: func(c Config) string { return formatBoolSetting(c.Transport.Discord.AutoStart) },
	},
	{
		Key: DiscordBotsSetting, Label: "Discord 参与 Bot 列表", Group: groupTransport, Subsection: subDiscord,
		Editable: true,
		apply: func(c *Config, v string) error {
			entries, err := parseDiscordBotsJSON(v)
			if err != nil {
				return err
			}
			return applyDiscordBots(c, entries)
		},
		read: func(c Config) string { return formatDiscordBotsJSON(discordBotsFromTransport(c.Transport.Discord)) },
	},
}

func parseBoolSetting(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func formatBoolSetting(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func editableSettingFields() map[string]settingField {
	out := make(map[string]settingField)
	for _, f := range settingFields {
		if f.Editable {
			out[f.Key] = f
		}
	}
	return out
}

// IsEditableSettingKey reports whether key may be stored in app_settings.
func IsEditableSettingKey(key string) bool {
	_, ok := editableSettingFields()[key]
	return ok
}

// IsPersistableSettingKey reports whether key may be written to app_settings
// (editable fields plus internal cache keys managed by the server).
func IsPersistableSettingKey(key string) bool {
	if IsEditableSettingKey(key) {
		return true
	}
	return key == DiscordBotProfilesSetting || key == DiscordBotTokensSetting || key == MeetPresetsSetting || key == MeetParticipantsSetting || key == ParticipantIMBindingsSetting || key == DiscordModeratorRoleSetting || key == DiscordBotsMigrationV2Setting
}

func applySettingsMap(cfg *Config, m map[string]string) error {
	for key, val := range m {
		field, ok := editableSettingFields()[key]
		if !ok {
			continue
		}
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		if field.apply == nil {
			continue
		}
		if err := field.apply(cfg, val); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}
	return nil
}
