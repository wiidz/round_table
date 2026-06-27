package main

import (
	"log"
	"os"

	"round_table/apps/server/internal/adapter/storage/sqlite"
	"round_table/apps/server/internal/platform/config"
)

func main() {
	cfg := config.Load()
	path := cfg.Storage.SQLitePath
	if path == "" {
		path = "./data/roundtable.db"
	}

	st, err := sqlite.Open(path)
	if err != nil {
		log.Fatalf("migrate: open %s: %v", path, err)
	}
	defer st.Close()

	log.Printf("migrate: ok (%s)", path)
	os.Exit(0)
}
