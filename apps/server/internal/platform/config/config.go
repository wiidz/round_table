package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds runtime settings.
// Load order (later wins): defaults → server.yaml → .env → process environment.
type Config struct {
	Server  Server  `yaml:"server"`
	Meeting Meeting `yaml:"meeting"`
	Storage Storage `yaml:"storage"`
	Secrets Secrets
}

type Server struct {
	Addr            string `yaml:"addr"`
	ReadTimeoutSec  int    `yaml:"read_timeout_sec"`
	WriteTimeoutSec int    `yaml:"write_timeout_sec"`
}

type Meeting struct {
	MaxRoundsPerSegment   int    `yaml:"max_rounds_per_segment"`
	MaxConfirmationCycles int    `yaml:"max_confirmation_cycles"`
	ConfirmationMode      string `yaml:"confirmation_mode"`
}

type Storage struct {
	Driver     string `yaml:"driver"`
	SQLitePath string `yaml:"sqlite_path"`
}

// Secrets are loaded only from .env / environment — never from YAML.
type Secrets struct {
	DatabaseDSN     string
	OpenAIAPIKey    string
	AnthropicAPIKey string
	DeepSeekAPIKey  string
}

// Addr returns the HTTP listen address.
func (c Config) Addr() string {
	return c.Server.Addr
}

// Load reads configuration from yaml, .env, and environment variables.
func Load() Config {
	cfg := defaults()
	root := configRoot()

	if data, err := os.ReadFile(root + "/configs/server.yaml"); err == nil {
		_ = yaml.Unmarshal(data, &cfg)
	}

	_ = loadEnvFile(root + "/.env")

	applyEnv(&cfg)
	cfg.Secrets = loadSecrets()
	return cfg
}

// SaveEnv writes non-secret overrides and secret keys to .env for local development.
func SaveEnv(cfg Config) error {
	return saveEnvFile(configRoot()+"/.env", cfg)
}

func defaults() Config {
	return Config{
		Server: Server{
			Addr:            ":7777",
			ReadTimeoutSec:  30,
			WriteTimeoutSec: 30,
		},
		Meeting: Meeting{
			MaxRoundsPerSegment:   5,
			MaxConfirmationCycles: 3,
			ConfirmationMode:      "required",
		},
		Storage: Storage{
			Driver:     "memory",
			SQLitePath: "./data/roundtable.db",
		},
	}
}

func applyEnv(cfg *Config) {
	overrideString(&cfg.Server.Addr, "ROUND_TABLE_ADDR")
	overrideInt(&cfg.Server.ReadTimeoutSec, "ROUND_TABLE_READ_TIMEOUT_SEC")
	overrideInt(&cfg.Server.WriteTimeoutSec, "ROUND_TABLE_WRITE_TIMEOUT_SEC")

	overrideInt(&cfg.Meeting.MaxRoundsPerSegment, "ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT")
	overrideInt(&cfg.Meeting.MaxConfirmationCycles, "ROUND_TABLE_MAX_CONFIRMATION_CYCLES")
	overrideString(&cfg.Meeting.ConfirmationMode, "ROUND_TABLE_CONFIRMATION_MODE")

	overrideString(&cfg.Storage.Driver, "ROUND_TABLE_STORAGE_DRIVER")
	overrideString(&cfg.Storage.SQLitePath, "ROUND_TABLE_STORAGE_SQLITE_PATH")
}

func loadSecrets() Secrets {
	return Secrets{
		DatabaseDSN:     os.Getenv("ROUND_TABLE_DATABASE_DSN"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		DeepSeekAPIKey:  os.Getenv("DEEPSEEK_API_KEY"),
	}
}

func overrideString(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func overrideInt(dst *int, key string) {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			*dst = n
		}
	}
}
