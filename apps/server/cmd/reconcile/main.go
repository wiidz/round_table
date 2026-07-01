package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"round_table/apps/server/internal/platform/bootstrap"
	"round_table/apps/server/internal/platform/config"
)

func main() {
	reason := "手动 reconcile：清理僵尸会议"
	if len(os.Args) > 1 && os.Args[1] != "" {
		reason = os.Args[1]
	}

	cfg := config.Load()
	store, err := bootstrap.OpenStorage(cfg.Storage)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	result, err := bootstrap.ReconcileMeetings(context.Background(), cfg, store, reason)
	if err != nil {
		log.Fatalf("reconcile: %v", err)
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
}
