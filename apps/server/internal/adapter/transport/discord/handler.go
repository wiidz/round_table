package discord

import (
	"context"
	"fmt"
	"strings"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
)

// CommandHandler routes RoundTable Discord text commands (Principal binding, help).
type CommandHandler struct {
	Prefix   string
	Registry *principalbind.Registry
}

// NewCommandHandler returns a handler with normalized prefix (trailing space).
func NewCommandHandler(prefix string, reg *principalbind.Registry) *CommandHandler {
	p := strings.TrimSpace(prefix)
	if p == "" {
		p = "!rt"
	}
	return &CommandHandler{Prefix: p + " ", Registry: reg}
}

// Handle implements transport.MessageHandler.
func (h *CommandHandler) Handle(_ context.Context, msg transport.Inbound) (string, error) {
	body := strings.TrimSpace(msg.Content)
	if !strings.HasPrefix(body, h.Prefix) {
		return "", nil
	}
	args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(body, h.Prefix)))
	if len(args) == 0 {
		return h.helpText(), nil
	}

	switch strings.ToLower(args[0]) {
	case "help", "h", "?":
		return h.helpText(), nil
	case "principal", "p":
		return h.handlePrincipal(msg, args[1:])
	default:
		return fmt.Sprintf("未知指令 `%s`。发送 `%shelp` 查看用法。", args[0], h.Prefix), nil
	}
}

func (h *CommandHandler) handlePrincipal(msg transport.Inbound, args []string) (string, error) {
	if len(args) == 0 {
		return fmt.Sprintf("用法：`%sprincipal bind|whoami|unbind`", h.Prefix), nil
	}
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	display := msg.AuthorName
	if display == "" {
		display = msg.AuthorID
	}

	switch strings.ToLower(args[0]) {
	case "bind", "register":
		b, err := h.Registry.Bind(scope, msg.Platform, msg.AuthorID, display)
		if err != nil {
			return "绑定失败：" + err.Error(), nil
		}
		scopeLabel := "本服务器"
		if msg.GuildID == "" {
			scopeLabel = "你的私信会话"
		}
		return fmt.Sprintf("已绑定 Principal\n- ID: `%s`\n- 显示名: %s\n- 范围: %s",
			b.PrincipalID, b.DisplayName, scopeLabel), nil

	case "whoami", "me", "status":
		b, ok := h.Registry.Get(scope)
		if !ok {
			return fmt.Sprintf("当前范围尚未绑定 Principal。发送 `%sprincipal bind` 绑定。", h.Prefix), nil
		}
		if b.ExternalID == msg.AuthorID {
			return fmt.Sprintf("你是本范围的 Principal\n- ID: `%s`\n- 显示名: %s\n- 绑定于: %s",
				b.PrincipalID, b.DisplayName, b.BoundAt.Format("2006-01-02 15:04 UTC")), nil
		}
		return fmt.Sprintf("本范围 Principal 是 **%s** (`%s`)", b.DisplayName, b.PrincipalID), nil

	case "unbind", "unregister":
		if err := h.Registry.Unbind(scope, msg.AuthorID); err != nil {
			return "解绑失败：" + err.Error(), nil
		}
		return "已解除 Principal 绑定。", nil

	default:
		return fmt.Sprintf("用法：`%sprincipal bind|whoami|unbind`", h.Prefix), nil
	}
}

func (h *CommandHandler) helpText() string {
	return fmt.Sprintf(`**RoundTable Discord 指令**

前缀：`+"`%s`"+`

- `+"`%shelp`"+` — 显示帮助
- `+"`%sprincipal bind`"+` — 将你自己绑定为本范围 Principal（每个服务器/私信会话仅一位）
- `+"`%sprincipal whoami`"+` — 查看当前 Principal 绑定
- `+"`%sprincipal unbind`"+` — 解除你的 Principal 绑定`, h.Prefix, h.Prefix, h.Prefix, h.Prefix, h.Prefix)
}
