package discord

import (
	"context"
	"strings"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/domain/meeting"
)

// CommandHandler routes RoundTable Discord text commands (Principal binding, help).
type CommandHandler struct {
	Prefix   string
	Registry *principalbind.Registry
	Meet     *MeetRunner
}

// NewCommandHandler returns a handler with normalized prefix (trailing space).
func NewCommandHandler(prefix string, reg *principalbind.Registry, meet *MeetRunner) *CommandHandler {
	p := strings.TrimSpace(prefix)
	if p == "" {
		p = "!rt"
	}
	return &CommandHandler{Prefix: p + " ", Registry: reg, Meet: meet}
}

// Handle implements transport.MessageHandler.
func (h *CommandHandler) Handle(_ context.Context, msg transport.Inbound) (string, error) {
	body := strings.TrimSpace(msg.Content)

	if h.Meet != nil && isMeetCancelTrigger(body) {
		if reply, ok := h.Meet.CancelSetup(msg.ChannelID, msg.AuthorID); ok {
			return reply, nil
		}
		return meetSetupNothingToCancelText(h.locale()), nil
	}

	if h.Meet != nil {
		if reply, err := h.Meet.HandleConfirmationReply(msg); err != nil {
			return "", err
		} else if reply != "" {
			return reply, nil
		}
	}

	if h.Meet != nil {
		if reply, err := h.Meet.HandleSetupReply(msg); err != nil {
			return "", err
		} else if reply != "" {
			return reply, nil
		}
	}

	if h.Meet != nil && isMeetStartTrigger(body) {
		return h.Meet.BeginSetupFromTrigger(msg)
	}

	if !strings.HasPrefix(body, h.Prefix) {
		return "", nil
	}
	args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(body, h.Prefix)))
	if len(args) == 0 {
		return h.helpText(), nil
	}

	loc := h.locale()
	switch strings.ToLower(args[0]) {
	case "help", "h", "?":
		return h.helpText(), nil
	case "principal", "p":
		return h.handlePrincipal(msg, args[1:])
	case "meet", "m":
		return h.handleMeet(msg, args[1:])
	default:
		return unknownCommandText(loc, h.Prefix, args[0]), nil
	}
}

func (h *CommandHandler) handlePrincipal(msg transport.Inbound, args []string) (string, error) {
	loc := h.locale()
	if len(args) == 0 {
		return principalUsageText(loc, h.Prefix), nil
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
			return principalBindFailedText(loc, err), nil
		}
		return principalBindOKText(loc, b.PrincipalID, b.DisplayName, scopeLabel(loc, msg.GuildID == "")), nil

	case "whoami", "me", "status":
		b, ok := h.Registry.Get(scope)
		if !ok {
			return principalNotBoundText(loc, h.Prefix), nil
		}
		if b.ExternalID == msg.AuthorID {
			return principalWhoamiSelfText(loc, b.PrincipalID, b.DisplayName, b.BoundAt.Format("2006-01-02 15:04 UTC")), nil
		}
		return principalWhoamiOtherText(loc, b.DisplayName, b.PrincipalID), nil

	case "unbind", "unregister":
		if err := h.Registry.Unbind(scope, msg.AuthorID); err != nil {
			return principalUnbindFailedText(loc, err), nil
		}
		return principalUnbindOKText(loc), nil

	default:
		return principalUsageText(loc, h.Prefix), nil
	}
}

func (h *CommandHandler) handleMeet(msg transport.Inbound, args []string) (string, error) {
	loc := h.locale()
	if h.Meet == nil {
		return meetDisabledText(loc), nil
	}
	if len(args) > 0 && strings.EqualFold(args[0], "cancel") {
		if reply, ok := h.Meet.CancelSetup(msg.ChannelID, msg.AuthorID); ok {
			return reply, nil
		}
		return meetSetupNothingToCancelText(loc), nil
	}
	defaultMode := h.Meet.Discord.MeetMode
	if defaultMode == "" {
		defaultMode = meeting.MeetingModeDecision
	}
	parsed, err := parseMeetArgs(args, defaultMode)
	if err != nil {
		return meetUsageText(loc, h.Prefix), nil
	}
	return h.Meet.BeginSetup(msg, parsed)
}
