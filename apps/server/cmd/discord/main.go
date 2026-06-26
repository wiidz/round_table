package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"round_table/apps/server/internal/adapter/transport"
	discordtransport "round_table/apps/server/internal/adapter/transport/discord"
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

	bot, err := discordtransport.New(discordtransport.Options{
		Token:      cfg.Secrets.DiscordBotToken,
		AllowDM:    dc.AllowDM,
		AllowGuild: dc.AllowGuild,
		GuildID:    dc.GuildID,
	})
	if err != nil {
		log.Fatalf("discord: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("discord bot connected — allow_dm=%v allow_guild=%v guild_id=%q", dc.AllowDM, dc.AllowGuild, dc.GuildID)
	log.Printf("send a message in an allowed channel; bot replies with echo")

	handler := func(_ context.Context, msg transport.Inbound) (string, error) {
		log.Printf("inbound channel=%s author=%s: %s", msg.ChannelID, msg.AuthorID, msg.Content)
		return fmt.Sprintf("RoundTable 收到: %s", msg.Content), nil
	}

	if err := bot.Run(ctx, handler); err != nil {
		log.Fatalf("discord: %v", err)
	}
}
