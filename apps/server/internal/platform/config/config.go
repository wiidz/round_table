package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds runtime settings.
// Load order (later wins): defaults → server.yaml → deploy/.env → process environment.
type Config struct {
	Server    Server    `yaml:"server"`
	Meeting   Meeting   `yaml:"meeting"`
	Model     Model     `yaml:"model"`
	Storage   Storage   `yaml:"storage"`
	Workspace Workspace `yaml:"workspace"`
	Profile   Profile   `yaml:"profile"`
	Knowledge Knowledge `yaml:"knowledge"`
	Transport Transport `yaml:"transport"`
	Secrets   Secrets
}

type Server struct {
	Addr            string `yaml:"addr"`
	ReadTimeoutSec  int    `yaml:"read_timeout_sec"`
	WriteTimeoutSec int    `yaml:"write_timeout_sec"`
	Locale          string `yaml:"locale"` // en | zh — user-facing copy (Transport, Web, …)
}

type Meeting struct {
	DefaultMode              string `yaml:"default_mode"`
	MaxRoundsPerSegment        int    `yaml:"max_rounds_per_segment"`
	MinRoundsBeforeSynthesis   int    `yaml:"min_rounds_before_synthesis"`
	MaxConfirmationCycles      int    `yaml:"max_confirmation_cycles"`
	ConfirmationMode         string `yaml:"confirmation_mode"`
	FreeDialogueMaxQuestions int    `yaml:"free_dialogue_max_questions"`
	LLMModeratorRoundSummary bool   `yaml:"llm_moderator_round_summary"`
	LLMModeratorExecutiveRecap bool `yaml:"llm_moderator_executive_recap"`
	MeetPresets              []MeetPresetConfig `yaml:"-"`
	MeetCasts                []MeetCastConfig   `yaml:"-"`
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

// Transport holds external chat platform settings (non-secret).
type Transport struct {
	Discord DiscordTransport `yaml:"discord"`
}

// DiscordTransport configures the Discord bot adapter.
type DiscordTransport struct {
	AutoStart                 bool   `yaml:"auto_start"`
	Enabled                   bool   `yaml:"enabled"`
	AllowDM                 bool   `yaml:"allow_dm"`
	AllowGuild              bool   `yaml:"allow_guild"`
	GuildID                 string `yaml:"guild_id"`
	CommandPrefix           string `yaml:"command_prefix"`
	BindingsFile            string `yaml:"bindings_file"`
	MeetParticipants        string `yaml:"meet_participants"`
	MeetMode                string `yaml:"meet_mode"`
	MeetConfirmation        string `yaml:"meet_confirmation"`
	MeetMaxRounds           int    `yaml:"meet_max_rounds"`
	MeetFreeDialogueQuestions int  `yaml:"meet_free_dialogue_questions"`
	ParticipantBots         string `yaml:"participant_bots"`
	Locale                  string `yaml:"locale"`
	ReceptionAgentEnabled   bool   `yaml:"reception_agent_enabled"`
	ParticipantIMBindings   ParticipantIMBindings `yaml:"-"`
}

// Secrets are loaded from .env / environment and SQLite app_settings (Discord bot tokens).
type Secrets struct {
	DatabaseDSN              string
	OpenAIAPIKey             string
	AnthropicAPIKey          string
	DeepSeekAPIKey           string
	DiscordBotToken          string
	DiscordParticipantTokens map[string]string // participant id -> token
}

// Addr returns the HTTP listen address.
func (c Config) Addr() string {
	return c.Server.Addr
}

// Locale returns user-facing locale (en | zh).
func (c Config) Locale() string {
	if loc := strings.TrimSpace(c.Server.Locale); loc != "" {
		return loc
	}
	return c.Transport.Discord.Locale
}

// Load reads configuration from yaml, deploy/.env, and environment variables.
// For runtime overrides from SQLite app_settings, use NewService.
func Load() Config {
	return loadBase()
}

func loadBase() Config {
	cfg := defaults()
	root := configRoot()

	if data, err := os.ReadFile(root + "/configs/server.yaml"); err == nil {
		_ = yaml.Unmarshal(data, &cfg)
	}

	_ = loadEnvFile(deployEnvPath())

	applyEnv(&cfg)
	cfg.Secrets = loadSecrets()
	normalizeLoadedConfig(&cfg)
	return cfg
}

// SaveEnv writes non-secret overrides and secret keys to deploy/.env for local development.
func SaveEnv(cfg Config) error {
	return saveEnvFile(deployEnvPath(), cfg)
}

func defaults() Config {
	return Config{
		Server: Server{
			Addr:            ":7777",
			ReadTimeoutSec:  30,
			WriteTimeoutSec: 30,
			Locale:          "en",
		},
		Meeting: Meeting{
			DefaultMode:              "deliberation",
			MaxRoundsPerSegment:      5,
			MinRoundsBeforeSynthesis: 2,
			MaxConfirmationCycles:    3,
			ConfirmationMode:         "required",
			FreeDialogueMaxQuestions: 1,
			LLMModeratorRoundSummary: true,
			LLMModeratorExecutiveRecap: true,
		},
		Model: Model{
			Provider:     "deepseek",
			BaseURL:      "https://api.deepseek.com/v1",
			DefaultModel: "deepseek-chat",
			TimeoutSec:   120,
		},
		Storage: Storage{
			Driver:     "sqlite",
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
		Transport: Transport{
			Discord: DiscordTransport{
				AutoStart:                 false,
				Enabled:                   false,
				AllowDM:                   true,
				AllowGuild:                true,
				CommandPrefix:             "!rt",
				BindingsFile:              "./data/transport/discord-principal.json",
				MeetParticipants:          "designer:游戏策划:gameplay,player:玩家代表:experience,dev:开发:backend,ops:运维:infra",
				MeetMode:                  "deliberation",
				MeetConfirmation:          "skip",
				MeetMaxRounds:             2,
				MeetFreeDialogueQuestions: 0,
				ParticipantBots:           "designer,player,dev,ops",
				Locale:                    "en",
				ReceptionAgentEnabled:     true,
			},
		},
	}
}

func applyEnv(cfg *Config) {
	overrideString(&cfg.Server.Addr, "ROUND_TABLE_ADDR")
	overrideInt(&cfg.Server.ReadTimeoutSec, "ROUND_TABLE_READ_TIMEOUT_SEC")
	overrideInt(&cfg.Server.WriteTimeoutSec, "ROUND_TABLE_WRITE_TIMEOUT_SEC")
	overrideString(&cfg.Server.Locale, "ROUND_TABLE_LOCALE")
	if cfg.Server.Locale == "" {
		overrideString(&cfg.Server.Locale, "ROUND_TABLE_DISCORD_LOCALE")
	}

	overrideString(&cfg.Meeting.DefaultMode, "ROUND_TABLE_DEFAULT_MEET_MODE")
	if cfg.Meeting.DefaultMode == "" {
		overrideString(&cfg.Meeting.DefaultMode, "ROUND_TABLE_DISCORD_MEET_MODE")
	}
	overrideInt(&cfg.Meeting.MaxRoundsPerSegment, "ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT")
	overrideInt(&cfg.Meeting.MinRoundsBeforeSynthesis, "ROUND_TABLE_MIN_ROUNDS_BEFORE_SYNTHESIS")
	overrideInt(&cfg.Meeting.MaxConfirmationCycles, "ROUND_TABLE_MAX_CONFIRMATION_CYCLES")
	overrideInt(&cfg.Meeting.FreeDialogueMaxQuestions, "ROUND_TABLE_FREE_DIALOGUE_MAX_QUESTIONS")
	overrideString(&cfg.Meeting.ConfirmationMode, "ROUND_TABLE_CONFIRMATION_MODE")
	overrideBool(&cfg.Meeting.LLMModeratorRoundSummary, "ROUND_TABLE_LLM_MODERATOR_ROUND_SUMMARY")
	overrideBool(&cfg.Meeting.LLMModeratorExecutiveRecap, "ROUND_TABLE_LLM_MODERATOR_EXECUTIVE_RECAP")

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

	overrideBool(&cfg.Transport.Discord.AutoStart, "ROUND_TABLE_DISCORD_AUTO_START")
	overrideBool(&cfg.Transport.Discord.Enabled, "ROUND_TABLE_DISCORD_ENABLED")
	overrideBool(&cfg.Transport.Discord.AllowDM, "ROUND_TABLE_DISCORD_ALLOW_DM")
	overrideBool(&cfg.Transport.Discord.AllowGuild, "ROUND_TABLE_DISCORD_ALLOW_GUILD")
	overrideString(&cfg.Transport.Discord.GuildID, "ROUND_TABLE_DISCORD_GUILD_ID")
	overrideString(&cfg.Transport.Discord.CommandPrefix, "ROUND_TABLE_DISCORD_COMMAND_PREFIX")
	overrideString(&cfg.Transport.Discord.BindingsFile, "ROUND_TABLE_DISCORD_BINDINGS_FILE")
	// Legacy discord-only keys — prefer ROUND_TABLE_LOCALE / meeting settings.
	overrideString(&cfg.Transport.Discord.Locale, "ROUND_TABLE_DISCORD_LOCALE")
	overrideString(&cfg.Transport.Discord.MeetMode, "ROUND_TABLE_DISCORD_MEET_MODE")
	overrideInt(&cfg.Transport.Discord.MeetMaxRounds, "ROUND_TABLE_DISCORD_MEET_MAX_ROUNDS")
}

func loadSecrets() Secrets {
	return Secrets{
		DatabaseDSN:     os.Getenv("ROUND_TABLE_DATABASE_DSN"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		DeepSeekAPIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		DiscordBotToken: os.Getenv("DISCORD_BOT_TOKEN"),
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

func overrideBool(dst *bool, key string) {
	if v := os.Getenv(key); v != "" {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			*dst = true
		case "0", "false", "no", "off":
			*dst = false
		}
	}
}
