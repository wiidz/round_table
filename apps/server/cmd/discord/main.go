package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	discordtransport "round_table/apps/server/internal/adapter/transport/discord"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/bootstrap"
	"round_table/apps/server/internal/platform/config"
)

func main() {
	startedAt := time.Now()
	base := config.Load()
	dc := base.Transport.Discord

	log.Printf("[discord] 正在启动 RoundTable Discord Transport…")

	if base.Secrets.DiscordBotToken == "" {
		fmt.Fprintln(os.Stderr, "discord: configure Moderator bot token in Web 设置 → IM → Discord")
		fmt.Fprintln(os.Stderr, "       optional: transport.discord.* in apps/server/configs/server.yaml")
		os.Exit(2)
	}

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
	log.Printf("[discord] 配置已加载 · locale=%s · prefix=%q", loc, dc.CommandPrefix)

	botOpts := discordtransport.Options{
		AllowDM:    dc.AllowDM,
		AllowGuild: dc.AllowGuild,
		GuildID:    dc.GuildID,
		Locale:     loc,
	}

	participantTotal := len(config.DiscordBotApplicationIDs(base))
	log.Printf("[discord] 正在连接 Discord Gateway…")

	bot, err := discordtransport.New(discordtransport.Options{
		Token:      base.Secrets.DiscordBotToken,
		AllowDM:    dc.AllowDM,
		AllowGuild: dc.AllowGuild,
		GuildID:    dc.GuildID,
		Locale:     loc,
	})
	if err != nil {
		log.Fatalf("discord: %v", err)
	}

	pool, err := discordtransport.OpenBotPool(discordtransport.PoolOptions{
		Default: bot,
		BotOpts: botOpts,
		BotIDs:  config.DiscordBotApplicationIDs(base),
		Mapping: nil,
		ResolveToken: func(botID string) string {
			if base.Secrets.DiscordParticipantTokens == nil {
				return ""
			}
			return base.Secrets.DiscordParticipantTokens[botID]
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
		Bots:      pool,
		Principal: discordtransport.NewChannelPrincipal(pool, loc),
	}

	locale := discordtransport.ParseLocale(loc)
	bot.SetOnGatewayResumed(func() {
		for _, chID := range meet.ActiveMeetingChannelIDs() {
			_ = bot.Send(context.Background(), chID, discordtransport.GatewayResumedText(locale))
		}
	})

	cmd := discordtransport.NewCommandHandler(dc.CommandPrefix, reg, meet)

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
