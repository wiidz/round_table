package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	profFS "round_table/apps/server/internal/adapter/profile/fs"
	brieffs "round_table/apps/server/internal/adapter/brief/fs"
	discordtransport "round_table/apps/server/internal/adapter/transport/discord"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/bootstrap"
	"round_table/apps/server/internal/platform/config"
	"round_table/apps/server/internal/platform/discordsvc"
)

func main() {
	startedAt := time.Now()
	base := config.Load()
	dc := base.Transport.Discord

	log.Printf("[discord] 正在启动 RoundTable Discord Transport…")

	store, err := bootstrap.OpenStorage(base.Storage)
	if err != nil {
		log.Fatalf("discord: storage: %v", err)
	}
	var configSvc *config.Service
	if s, ok := store.(config.SettingsStore); ok {
		configSvc, err = config.NewService(s)
		if err != nil {
			log.Fatalf("discord: config: %v", err)
		}
		base = configSvc.Current()
		dc = base.Transport.Discord
	}

	reg, err := principalbind.NewRegistry(dc.BindingsFile)
	if err != nil {
		log.Fatalf("discord: principal registry: %v", err)
	}

	loc := base.Locale()
	var overrides map[string]string
	if configSvc != nil {
		overrides = configSvc.SettingsOverrides()
	}
	primaryID := config.EffectivePrimaryBotID(overrides)
	tokens := config.EffectiveDiscordBotTokens(base, overrides)
	poolBotIDs := config.FilterDiscordParticipantBotIDs(base, overrides)
	hostToken := config.HostBotToken(base, overrides)
	gatewayToken := strings.TrimSpace(hostToken)
	if gatewayToken == "" {
		gatewayToken = strings.TrimSpace(base.Secrets.DiscordBotToken)
	}
	if gatewayToken == "" {
		fmt.Fprintln(os.Stderr, "discord: configure host bot token in Web 设置 → IM → Discord")
		os.Exit(2)
	}

	log.Printf("[discord] 配置已加载 · locale=%s · prefix=%q · host_bot=%q", loc, dc.CommandPrefix, primaryID)

	pidPath := config.DiscordTransportPIDPath(base)
	lockFile, err := discordsvc.AcquireTransportLock(discordsvc.TransportLockPath(base))
	if err != nil {
		log.Fatalf("discord: %v", err)
	}
	defer discordsvc.ReleaseTransportLock(lockFile)

	if err := writeTransportPID(pidPath, os.Getpid()); err != nil {
		log.Printf("[discord] warning: write pid file: %v", err)
	}
	defer os.Remove(pidPath)

	botOpts := discordtransport.Options{
		AllowDM:    dc.AllowDM,
		AllowGuild: dc.AllowGuild,
		GuildID:    dc.GuildID,
		Locale:     loc,
	}

	participantTotal := len(config.DiscordBotApplicationIDs(base))
	log.Printf("[discord] 正在连接 Discord Gateway…")

	bot, err := discordtransport.New(discordtransport.Options{
		Token:      gatewayToken,
		AllowDM:    dc.AllowDM,
		AllowGuild: dc.AllowGuild,
		GuildID:    dc.GuildID,
		Locale:     loc,
	})
	if err != nil {
		log.Fatalf("discord: %v", err)
	}
	bot.SetHostGuard(func() bool {
		if configSvc == nil {
			return strings.TrimSpace(hostToken) != "" && strings.TrimSpace(hostToken) == bot.Token()
		}
		overrides := configSvc.SettingsOverrides()
		return config.BotShouldHandleCommandsForToken(bot.Token(), configSvc.Current(), overrides)
	})
	dedupDir := config.DiscordInboundDedupDir(base)
	pruneInboundDedup := func() {
		discordtransport.PruneInboundMessageClaims(dedupDir)
	}
	pruneInboundDedup()
	bot.SetMessageClaim(func(messageID string) bool {
		return discordtransport.ClaimInboundMessage(dedupDir, messageID)
	})

	pool, err := discordtransport.OpenBotPool(discordtransport.PoolOptions{
		Default:   bot,
		BotOpts:   botOpts,
		BotIDs:    poolBotIDs,
		HostToken: hostToken,
		Mapping:   nil,
		ResolveToken: func(botID string) string {
			return tokens.TokenForBot(botID, primaryID)
		},
		ParticipantBotID: func(participantID string) string {
			return config.DiscordBotForParticipant(base.Transport.Discord.ParticipantIMBindings, participantID)
		},
	})
	if err != nil {
		log.Fatalf("discord: participant bots: %v", err)
	}
	defer pool.Close()

	meet := &discordtransport.MeetRunner{
		Cfg:       base,
		ConfigSvc: configSvc,
		Discord:   dc,
		Registry:  reg,
		Briefs:    brieffs.NewStore(base.Brief.Root, base.Brief.Templates),
		Bots:      pool,
		Principal: discordtransport.NewChannelPrincipal(pool, loc),
	}

	var participantAdmin *discordtransport.ParticipantAdmin
	profileStore := profFS.NewStore(base.Profile.Root, base.Profile.Templates)
	if configSvc != nil {
		participantAdmin = &discordtransport.ParticipantAdmin{
			ConfigSvc: configSvc,
			Profile:   profileStore,
			Locale:    func() discordtransport.Locale { return discordtransport.ParseLocale(loc) },
			Prefix:    strings.TrimSpace(dc.CommandPrefix),
		}
		if participantAdmin.Prefix == "" {
			participantAdmin.Prefix = "!rt"
		}
	}

	locale := discordtransport.ParseLocale(loc)
	bot.SetOnGatewayResumed(func() {
		for _, chID := range meet.ActiveMeetingChannelIDs() {
			_ = bot.Send(context.Background(), chID, discordtransport.GatewayResumedText(locale))
		}
	})

	cmd := discordtransport.NewCommandHandler(dc.CommandPrefix, reg, meet)
	cmd.Profiles = profileStore
	cmd.Participants = participantAdmin

	modelPort, modelName := bootstrap.NewModelPortOptional(base)
	if modelPort != nil && dc.ReceptionAgentEnabled {
		cmd.Reception = &discordtransport.Reception{
			Model:        modelPort,
			ModelName:    modelName,
			Enabled:      true,
			Registry:     reg,
			Profiles:     profileStore,
			Meet:         meet,
			Participants: participantAdmin,
			Phase: func(channelID string) discordtransport.ChannelInputPhase {
				if participantAdmin != nil {
					if phase := participantAdmin.InputPhase(channelID); phase != discordtransport.InputPhaseIdle {
						return phase
					}
				}
				return meet.InputPhase(channelID)
			},
			Locale: func() discordtransport.Locale { return discordtransport.ParseLocale(loc) },
		}
		log.Printf("[discord] Reception Agent 已启用 · model=%s", modelName)
	} else if dc.ReceptionAgentEnabled && modelPort == nil {
		log.Printf("[discord] Reception Agent 已配置但未挂载（缺少 model API key）")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for _, line := range discordtransport.StartupReadyLogLines(cmd, discordtransport.StartupInfo{
		StartedAt:            startedAt,
		Prefix:               cmd.Prefix,
		BindingsFile:         dc.BindingsFile,
		Locale:               discordtransport.ParseLocale(loc),
		ModeratorUsername:    bot.DisplayName(),
		ParticipantConnected:   pool.Count(),
		ParticipantTotal:     participantTotal,
	}) {
		log.Print(line)
	}

	if err := bot.Run(ctx, cmd.Handle); err != nil {
		log.Fatalf("discord: %v", err)
	}
}

func writeTransportPID(path string, pid int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0o644)
}
