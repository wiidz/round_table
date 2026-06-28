package discord

import (
	"fmt"
	"strings"
	"time"
)

// ReadyLogMarker is written when the Discord gateway is up and the bot accepts commands.
// The supervisor scans logs for this marker to distinguish starting vs ready.
const ReadyLogMarker = "[discord] ready ·"

// StartupInfo holds metadata for the readable startup banner in transport logs.
type StartupInfo struct {
	StartedAt            time.Time
	Prefix               string
	BindingsFile         string
	Locale               Locale
	ModeratorUsername    string
	ParticipantConnected int
	ParticipantTotal     int
}

// StartupReadyLogLines returns plain-text log lines for a successful Discord transport boot.
func StartupReadyLogLines(h *CommandHandler, info StartupInfo) []string {
	p := strings.TrimSpace(info.Prefix)
	if p != "" && !strings.HasSuffix(p, " ") {
		p += " "
	}
	loc := info.Locale
	if h != nil {
		loc = h.locale()
	}
	if loc == "" {
		loc = LocaleZH
	}

	elapsed := time.Since(info.StartedAt).Round(time.Millisecond)
	readyAt := info.StartedAt.Add(elapsed)

	lines := []string{
		"[discord] ── 服务就绪 ──",
		fmt.Sprintf("[discord] 启动完成：%s（耗时 %s）", formatLogTime(readyAt), formatLogDuration(elapsed)),
		fmt.Sprintf("[discord] 主持人 Bot：%s", orDash(info.ModeratorUsername)),
		fmt.Sprintf("[discord] 参与 Bot：%d/%d 已连接", info.ParticipantConnected, info.ParticipantTotal),
		fmt.Sprintf("[discord] 指令前缀：%q · Principal 绑定：%s", p, info.BindingsFile),
		"[discord]",
		"[discord] 可用指令：",
	}
	lines = append(lines, commandReferenceLogLines(loc, p)...)
	lines = append(lines,
		"[discord]",
		fmt.Sprintf("%s %s", ReadyLogMarker, readyAt.UTC().Format(time.RFC3339)),
	)
	return lines
}

func commandReferenceLogLines(loc Locale, prefix string) []string {
	if loc == LocaleZH {
		return []string{
			"  · 新会议 / 开始会议 / 会议开始 — 发起会议（无需前缀，主持人逐步引导）",
			"  · 取消会议 — 取消待确认的会议配置",
			"  · " + prefix + "help — 显示完整帮助",
			"  · " + prefix + "status — 查看频道输入态与可接受指令",
			"  · " + prefix + "principal bind — 绑定本范围 Principal（每服务器/私信一位）",
			"  · " + prefix + "principal whoami — 查看 Principal 绑定",
			"  · " + prefix + "principal unbind — 解除 Principal 绑定",
			"  · " + prefix + "meet [-mode decision|deliberation] 主题 — 带主题发起会议",
			"  · " + prefix + "meet cancel — 取消待确认的会议配置",
			"  · " + prefix + "expert list / 专家 列表 — 查看专家名录",
			"  · " + prefix + "expert new / 专家 新建 — 新建专家（逐步引导）",
			"  · " + prefix + "expert edit|delete <代号> — 编辑或删除专家",
			"  · 会议进行中：暂停会议 · 恢复会议 · 终止会议 · 立即合成（研讨）· 强制共识（裁决）",
			"  · 自由问答：提问 … / 提问 designer … — 指定参与者提问",
			"  · 会议结束后：获取纪要 · 获取草案 · 获取待决 · 获取结论",
		}
	}
	return []string{
		"  · 开始会议 / 新会议 / start meeting — start a meeting (no prefix; Moderator guides you)",
		"  · cancel meeting — cancel pending meet setup",
		"  · " + prefix + "help — show full help",
		"  · " + prefix + "status — show input phase and accepted commands",
		"  · " + prefix + "principal bind — bind Principal (one per server/DM)",
		"  · " + prefix + "principal whoami — show Principal binding",
		"  · " + prefix + "principal unbind — remove Principal binding",
		"  · " + prefix + "meet [-mode decision|deliberation] topic — start with inline topic",
		"  · " + prefix + "meet cancel — cancel pending meet setup",
		"  · " + prefix + "expert list — list expert roster",
		"  · " + prefix + "expert new — create expert (guided)",
		"  · " + prefix + "expert edit|delete <id> — update or remove expert",
		"  · While meeting runs: pause · resume · stop · synthesize now · force consensus",
		"  · Free dialogue: ask … / 提问 designer … — ask a participant",
		"  · After meeting: get minutes · get draft · get open · get conclusion",
	}
}

func formatLogTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04:05 MST")
}

func formatLogDuration(d time.Duration) string {
	if d < time.Second {
		return d.String()
	}
	return d.Round(100 * time.Millisecond).String()
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
