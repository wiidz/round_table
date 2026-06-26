package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"round_table/apps/server/internal/platform/config"
	"round_table/apps/server/internal/platform/server"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx, cfg.Addr, server.HealthHandler()); err != nil {
		log.Fatalf("server: %v", err)
	}
	os.Exit(0)
}
