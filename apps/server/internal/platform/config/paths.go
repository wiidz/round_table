package config

import "os"

// configRoot returns apps/server whether make is run from repo root or go run from apps/server.
func configRoot() string {
	if v := os.Getenv("ROUND_TABLE_ROOT"); v != "" {
		return v
	}
	if _, err := os.Stat("apps/server/configs/server.yaml"); err == nil {
		return "apps/server"
	}
	return "."
}
