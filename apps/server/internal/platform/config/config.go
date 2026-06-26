package config

import (
	"os"
)

// Config holds runtime settings loaded from configs/server.yaml (skeleton).
type Config struct {
	Addr string
}

// Load reads minimal config; full YAML parsing in a later phase.
func Load() Config {
	addr := os.Getenv("ROUND_TABLE_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return Config{Addr: addr}
}
