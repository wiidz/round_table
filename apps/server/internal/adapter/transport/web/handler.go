package web

import (
	"context"
	"strings"

	profFS "round_table/apps/server/internal/adapter/profile/fs"
	brieffs "round_table/apps/server/internal/adapter/brief/fs"
	discordtransport "round_table/apps/server/internal/adapter/transport/discord"
	"round_table/apps/server/internal/adapter/transport"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/platform/bootstrap"
	"round_table/apps/server/internal/platform/config"
)

// Handler serves browser chat — implicit session operator, no Principal bind step.
type Handler struct {
	meet         *discordtransport.MeetRunner
	reception    *discordtransport.Reception
	participants *discordtransport.ParticipantAdmin
	locale       discordtransport.Locale
	prefix       string
}

// NewHandler builds web chat with MeetRunner wired to the session Hub sender.
func NewHandler(cfg config.Config, configSvc *config.Service, hub *Hub) *Handler {
	dc := cfg.Transport.Discord
	loc := cfg.Locale()
	locale := discordtransport.ParseLocale(loc)

	prefix := strings.TrimSpace(dc.CommandPrefix)
	if prefix == "" {
		prefix = "!rt"
	}

	pool := NewBotPool(hub, dc.MeetParticipants)

	meet := &discordtransport.MeetRunner{
		Cfg:       cfg,
		ConfigSvc: configSvc,
		Discord:   dc,
		Briefs:    brieffs.NewStore(cfg.Brief.Root, cfg.Brief.Templates),
		Bots:      pool,
		Principal: discordtransport.NewChannelPrincipal(pool, loc),
	}

	var participantAdmin *discordtransport.ParticipantAdmin
	if configSvc != nil {
		participantAdmin = &discordtransport.ParticipantAdmin{
			ConfigSvc: configSvc,
			Profile:   profFS.NewStore(cfg.Profile.Root, cfg.Profile.Templates),
			Locale:    func() discordtransport.Locale { return locale },
			Prefix:    prefix,
		}
	}

	var reception *discordtransport.Reception
	modelPort, modelName := bootstrap.NewModelPortOptional(cfg)
	if modelPort != nil && dc.ReceptionAgentEnabled {
		reception = &discordtransport.Reception{
			Model:        modelPort,
			ModelName:    modelName,
			Enabled:      true,
			Meet:         meet,
			Participants: participantAdmin,
			Phase:        func(channelID string) discordtransport.ChannelInputPhase { return inputPhase(meet, reception, participantAdmin, channelID) },
			Locale:       func() discordtransport.Locale { return locale },
		}
	}

	return &Handler{
		meet:         meet,
		reception:    reception,
		participants: participantAdmin,
		locale:       locale,
		prefix:       prefix + " ",
	}
}

// Handle processes one inbound browser message.
func (h *Handler) Handle(ctx context.Context, msg transport.Inbound) (Reply, error) {
	if h == nil || msg.Platform != "web" {
		return Reply{}, nil
	}
	body := strings.TrimSpace(msg.Content)
	if body == "" {
		return Reply{}, nil
	}

	if discordtransport.MatchWebStatusTrigger(body) {
		phase := inputPhase(h.meet, h.reception, h.participants, msg.ChannelID)
		meetingID := ""
		if h.meet != nil {
			meetingID = h.meet.MeetingIDForPhase(msg.ChannelID, phase)
		}
		return moderatorReply(discordtransport.FormatInputPhaseStatus(h.locale, phase, meetingID)), nil
	}

	if discordtransport.MatchWebExpertListTrigger(body) {
		if h.participants != nil {
			return moderatorReply(h.participants.FormatList(h.locale)), nil
		}
		return systemReply(discordtransport.WebChatMutatingUnavailableText(h.locale)), nil
	}

	if h.reception != nil {
		if reply, err := h.reception.HandleConfirmReply(ctx, msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return classifyReply(reply), nil
		}
	}

	if h.reception != nil {
		if reply, err := h.reception.HandleClarifyFollowUp(ctx, msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return classifyReply(reply), nil
		}
	}

	if h.participants != nil {
		if reply, err := h.participants.HandleSetupReply(msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return moderatorReply(reply), nil
		}
	}

	if discordtransport.IsExpertCancelTrigger(body) && h.participants != nil {
		if reply, ok := h.participants.CancelSetup(msg.ChannelID, msg.AuthorID); ok {
			return moderatorReply(reply), nil
		}
	}

	if h.meet != nil && discordtransport.IsMeetCancelTrigger(body) {
		if reply, ok := h.meet.CancelSetup(msg.ChannelID, msg.AuthorID); ok {
			return moderatorReply(reply), nil
		}
		return moderatorReply(discordtransport.MeetSetupNothingToCancelText(h.locale)), nil
	}

	if h.meet != nil {
		if reply, err := h.meet.HandleConfirmationReply(msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return moderatorReply(reply), nil
		}
	}

	if h.meet != nil {
		if reply, err := h.meet.HandleArtifactFetch(msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return moderatorReply(reply), nil
		}
	}

	if h.meet != nil {
		if reply, err := h.meet.HandleFreeDialogueQuestion(msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return moderatorReply(reply), nil
		}
	}

	if h.meet != nil {
		if reply, err := h.meet.HandleRunningIntervention(msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return moderatorReply(reply), nil
		}
	}

	if strings.HasPrefix(body, h.prefix) {
		if reply, ok := h.handlePrefixCommand(msg, body); ok {
			return reply, nil
		}
	}

	if h.meet != nil {
		if reply, err := h.meet.HandleSetupReply(msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return moderatorReply(reply), nil
		}
	}

	if h.meet != nil {
		if reply, err := h.meet.TryBeginNaturalMeet(msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return moderatorReply(reply), nil
		}
	}

	if h.meet != nil && discordtransport.IsMeetStartTrigger(body) {
		reply, err := h.meet.BeginSetupFromTrigger(msg)
		if err != nil {
			return Reply{}, err
		}
		return moderatorReply(reply), nil
	}

	if h.reception != nil {
		if reply, err := h.reception.TryHandle(ctx, msg); err != nil {
			return Reply{}, err
		} else if reply != "" {
			return classifyReply(reply), nil
		}
	}

	if h.meet != nil {
		if hint, ok := h.meet.MisplacedInputHint(msg); ok {
			return moderatorReply(hint), nil
		}
	}

	return systemReply(discordtransport.WebChatNoMatchReply(h.locale)), nil
}

func (h *Handler) handlePrefixCommand(msg transport.Inbound, body string) (Reply, bool) {
	args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(body, h.prefix)))
	if len(args) == 0 {
		return systemReply(discordtransport.WebHelpText(h.locale)), true
	}
	switch strings.ToLower(args[0]) {
	case "help", "h", "?":
		return systemReply(discordtransport.WebHelpText(h.locale)), true
	case "status", "st":
		phase := inputPhase(h.meet, h.reception, h.participants, msg.ChannelID)
		meetingID := ""
		if h.meet != nil {
			meetingID = h.meet.MeetingIDForPhase(msg.ChannelID, phase)
		}
		return moderatorReply(discordtransport.FormatInputPhaseStatus(h.locale, phase, meetingID)), true
	case "principal", "p":
		return systemReply(discordtransport.WebChatBlockedActionHint(h.locale, "principal")), true
	case "meet", "m":
		reply, err := h.handleMeet(msg, args[1:])
		if err != nil {
			return Reply{}, false
		}
		return moderatorReply(reply), true
	case "expert", "experts", "专家", "e":
		return h.handleExpertCommand(args[1:])
	default:
		return Reply{}, false
	}
}

func (h *Handler) handleMeet(msg transport.Inbound, args []string) (string, error) {
	loc := h.locale
	if h.meet == nil {
		return discordtransport.MeetDisabledText(loc), nil
	}
	if len(args) > 0 && strings.EqualFold(args[0], "cancel") {
		if reply, ok := h.meet.CancelSetup(msg.ChannelID, msg.AuthorID); ok {
			return reply, nil
		}
		return discordtransport.MeetSetupNothingToCancelText(loc), nil
	}
	defaultMode := h.meet.ActiveCfg().Meeting.DefaultMode
	if defaultMode == "" {
		defaultMode = meeting.MeetingModeDecision
	}
	parsed, err := discordtransport.ParseMeetArgs(args, defaultMode)
	if err != nil {
		return discordtransport.MeetUsageText(loc, h.prefix), nil
	}
	return h.meet.BeginSetup(msg, parsed)
}

func (h *Handler) handleExpertCommand(args []string) (Reply, bool) {
	if len(args) == 0 {
		return systemReply(discordtransport.WebHelpText(h.locale)), true
	}
	switch strings.ToLower(args[0]) {
	case "list", "列表", "ls":
		if h.participants != nil {
			return moderatorReply(h.participants.FormatList(h.locale)), true
		}
		return systemReply(discordtransport.WebChatMutatingUnavailableText(h.locale)), true
	default:
		return systemReply(discordtransport.WebChatMutatingUnavailableText(h.locale)), true
	}
}

func inputPhase(
	meet *discordtransport.MeetRunner,
	reception *discordtransport.Reception,
	participants *discordtransport.ParticipantAdmin,
	channelID string,
) discordtransport.ChannelInputPhase {
	if reception != nil {
		if phase := reception.InputPhase(channelID); phase != discordtransport.InputPhaseIdle {
			return phase
		}
	}
	if participants != nil {
		if phase := participants.InputPhase(channelID); phase != discordtransport.InputPhaseIdle {
			return phase
		}
	}
	if meet != nil {
		return meet.InputPhase(channelID)
	}
	return discordtransport.InputPhaseIdle
}

func classifyReply(reply string) Reply {
	if discordtransport.IsWebPlatformHint(reply) {
		return systemReply(reply)
	}
	return moderatorReply(reply)
}
