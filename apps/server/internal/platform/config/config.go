package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds runtime settings.
// Load order (later wins): defaults → server.yaml → .env → process environment.
type Config struct {
	Server    Server    `yaml:"server"`
	Meeting   Meeting   `yaml:"meeting"`
	Model     Model     `yaml:"model"`
	Storage   Storage   `yaml:"storage"`
	Workspace Workspace `yaml:"workspace"`
	Profile   Profile   `yaml:"profile"`
	Knowledge Knowledge `yaml:"knowledge"`
	Secrets   Secrets
}

type Server struct {
	Addr            string `yaml:"addr"`
	ReadTimeoutSec  int    `yaml:"read_timeout_sec"`
	WriteTimeoutSec int    `yaml:"write_timeout_sec"`
}

type Meeting struct {
	MaxRoundsPerSegment      int    `yaml:"max_rounds_per_segment"`
	MaxConfirmationCycles    int    `yaml:"max_confirmation_cycles"`
	ConfirmationMode         string `yaml:"confirmation_mode"`
	FreeDialogueMaxQuestions int    `yaml:"free_dialogue_max_questions"`
}

type Model struct {
	Provider     string `yaml:"provider"`
	BaseURL      string `yaml:"base_url"`
	DefaultModel string `yaml:"default_model"`
	TimeoutSec   int    `yaml:"timeout_sec"`
}

type Storage struct {
	Driver     string `yaml:"driver"`
	SQLitePath string `yaml:"sqlite_path"`
}

type Workspace struct {
	Root string `yaml:"root"`
}

type Profile struct {
	Root      string `yaml:"root"`
	Templates string `yaml:"templates"`
}

type Knowledge struct {
	Root          string `yaml:"root"`
	Templates     string `yaml:"templates"`
	SharedEnabled bool   `yaml:"shared_enabled"`
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
			MaxRoundsPerSegment:      5,
			MaxConfirmationCycles:    3,
			ConfirmationMode:         "required",
			FreeDialogueMaxQuestions: 1,
		},
		Model: Model{
			Provider:     "deepseek",
			BaseURL:      "https://api.deepseek.com/v1",
			DefaultModel: "deepseek-chat",
			TimeoutSec:   120,
		},
		Storage: Storage{
			Driver:     "memory",
			SQLitePath: "./data/roundtable.db",
		},
		Workspace: Workspace{
			Root: "./data/workspaces",
		},
		Profile: Profile{
			Root:      "./data/profiles",
			Templates: "./data/_templates/profiles",
		},
		Knowledge: Knowledge{
			Root:          "./data/knowledge",
			Templates:     "./data/_templates/knowledge",
			SharedEnabled: true,
		},
	}
}

func applyEnv(cfg *Config) {
	overrideString(&cfg.Server.Addr, "ROUND_TABLE_ADDR")
	overrideInt(&cfg.Server.ReadTimeoutSec, "ROUND_TABLE_READ_TIMEOUT_SEC")
	overrideInt(&cfg.Server.WriteTimeoutSec, "ROUND_TABLE_WRITE_TIMEOUT_SEC")

	overrideInt(&cfg.Meeting.MaxRoundsPerSegment, "ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT")
	overrideInt(&cfg.Meeting.MaxConfirmationCycles, "ROUND_TABLE_MAX_CONFIRMATION_CYCLES")
	overrideInt(&cfg.Meeting.FreeDialogueMaxQuestions, "ROUND_TABLE_FREE_DIALOGUE_MAX_QUESTIONS")
	overrideString(&cfg.Meeting.ConfirmationMode, "ROUND_TABLE_CONFIRMATION_MODE")

	overrideString(&cfg.Model.Provider, "ROUND_TABLE_MODEL_PROVIDER")
	overrideString(&cfg.Model.BaseURL, "ROUND_TABLE_MODEL_BASE_URL")
	overrideString(&cfg.Model.DefaultModel, "DEEPSEEK_MODEL_NAME")
	overrideString(&cfg.Model.DefaultModel, "ROUND_TABLE_MODEL_DEFAULT_MODEL")
	overrideInt(&cfg.Model.TimeoutSec, "ROUND_TABLE_MODEL_TIMEOUT_SEC")

	overrideString(&cfg.Storage.Driver, "ROUND_TABLE_STORAGE_DRIVER")
	overrideString(&cfg.Storage.SQLitePath, "ROUND_TABLE_STORAGE_SQLITE_PATH")

	overrideString(&cfg.Workspace.Root, "ROUND_TABLE_WORKSPACE_ROOT")

	overrideString(&cfg.Profile.Root, "ROUND_TABLE_PROFILE_ROOT")
	overrideString(&cfg.Profile.Templates, "ROUND_TABLE_PROFILE_TEMPLATES")

	overrideString(&cfg.Knowledge.Root, "ROUND_TABLE_KNOWLEDGE_ROOT")
	overrideString(&cfg.Knowledge.Templates, "ROUND_TABLE_KNOWLEDGE_TEMPLATES")
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
