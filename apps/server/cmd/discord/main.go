package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	discordtransport "round_table/apps/server/internal/adapter/transport/discord"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

func main() {
	cfg := config.Load()
	dc := cfg.Transport.Discord

	if cfg.Secrets.DiscordBotToken == "" {
		fmt.Fprintln(os.Stderr, "discord: set DISCORD_BOT_TOKEN in apps/server/.env")
		fmt.Fprintln(os.Stderr, "       optional: transport.discord.* in apps/server/configs/server.yaml")
		os.Exit(2)
	}

	reg, err := principalbind.NewRegistry(dc.BindingsFile)
	if err != nil {
		log.Fatalf("discord: principal registry: %v", err)
	}

	bot, err := discordtransport.New(discordtransport.Options{
		Token:      cfg.Secrets.DiscordBotToken,
		AllowDM:    dc.AllowDM,
		AllowGuild: dc.AllowGuild,
		GuildID:    dc.GuildID,
	})
	if err != nil {
		log.Fatalf("discord: %v", err)
	}

	botOpts := discordtransport.Options{
		AllowDM:    dc.AllowDM,
		AllowGuild: dc.AllowGuild,
		GuildID:    dc.GuildID,
	}
	pool, err := discordtransport.OpenBotPool(discordtransport.PoolOptions{
		Default: bot,
		BotOpts: botOpts,
		Mapping: discordtransport.ParseParticipantBotMapping(dc.ParticipantBots),
	})
	if err != nil {
		log.Fatalf("discord: participant bots: %v", err)
	}
	defer pool.Close()

	meet := &discordtransport.MeetRunner{
		Cfg:      cfg,
		Discord:  dc,
		Registry: reg,
		Bots:     pool,
	}

	cmd := discordtransport.NewCommandHandler(dc.CommandPrefix, reg, meet)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("discord bot connected — prefix=%q bindings=%s", cmd.Prefix, dc.BindingsFile)
	log.Printf("discord participant bots: %d/%d connected",
		pool.Count(), len(discordtransport.ParseParticipantBotMapping(dc.ParticipantBots)))
	log.Printf("try: %sprincipal bind | %smeet 会议主题 | %shelp", cmd.Prefix, cmd.Prefix, cmd.Prefix)

	if err := bot.Run(ctx, cmd.Handle); err != nil {
		log.Fatalf("discord: %v", err)
	}
}
