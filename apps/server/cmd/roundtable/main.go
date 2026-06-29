package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"round_table/apps/server/internal/adapter/storage"
	httptransport "round_table/apps/server/internal/adapter/transport/http"
	"round_table/apps/server/internal/platform/bootstrap"
	"round_table/apps/server/internal/platform/config"
	"round_table/apps/server/internal/platform/discordsvc"
	"round_table/apps/server/internal/platform/server"
)

func main() {
	base := config.Load()

	store, err := bootstrap.OpenStorage(base.Storage)
	if err != nil {
		log.Fatalf("server: storage: %v", err)
	}
	var catalog storage.MeetingCatalog
	if c, ok := store.(storage.MeetingCatalog); ok {
		catalog = c
	}

	var settingsStore config.SettingsStore
	if s, ok := store.(config.SettingsStore); ok {
		settingsStore = s
	}
	configSvc, err := config.NewService(settingsStore)
	if err != nil {
		log.Fatalf("server: config: %v", err)
	}
	cfg := configSvc.Current()

	discordSvc := &discordsvc.Supervisor{}

	if cfg.Transport.Discord.AutoStart {
		if err := discordSvc.Start(context.Background(), cfg); err != nil {
			log.Printf("discord transport auto-start: %v", err)
		} else {
			log.Printf("discord transport auto-started")
		}
	}

	mux := http.NewServeMux()
	h, err := httptransport.NewHandler(cfg, catalog, configSvc, discordSvc)
	if err != nil {
		log.Fatalf("server: %v", err)
	}
	h.Register(mux)

	if root := strings.TrimSpace(os.Getenv("ROUND_TABLE_WEB_ROOT")); root != "" {
		if err := httptransport.RegisterWebUI(mux, root); err != nil {
			log.Fatalf("server: web ui: %v", err)
		}
		log.Printf("web ui: serving %s", root)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	defer discordSvc.Shutdown()

	if err := server.Run(ctx, cfg.Addr(), httptransport.WithCORS(mux)); err != nil {
		log.Fatalf("server: %v", err)
	}
	os.Exit(0)
}
