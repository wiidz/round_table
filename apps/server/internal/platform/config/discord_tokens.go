package config

import (
	"encoding/json"
	"os"
	"strings"
)

const DiscordBotTokensSetting = "ROUND_TABLE_DISCORD_BOT_TOKENS"

// DiscordBotTokens holds moderator and participant bot tokens (SQLite app_settings).
type DiscordBotTokens struct {
	Moderator    string            `json:"moderator,omitempty"`
	Participants map[string]string `json:"participants,omitempty"`
}

func parseDiscordBotTokensJSON(raw string) DiscordBotTokens {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DiscordBotTokens{}
	}
	var out DiscordBotTokens
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return DiscordBotTokens{}
	}
	if out.Participants == nil {
		out.Participants = make(map[string]string)
	}
	return out
}

func formatDiscordBotTokensJSON(tokens DiscordBotTokens) string {
	if tokens.Moderator == "" && len(tokens.Participants) == 0 {
		return "{}"
	}
	if tokens.Participants == nil {
		tokens.Participants = make(map[string]string)
	}
	b, _ := json.Marshal(tokens)
	return string(b)
}

func discordBotTokensFromOverrides(overrides map[string]string) DiscordBotTokens {
	if overrides == nil {
		return DiscordBotTokens{Participants: make(map[string]string)}
	}
	return parseDiscordBotTokensJSON(overrides[DiscordBotTokensSetting])
}

func (t DiscordBotTokens) TokenForBot(botID, primaryBotID string) string {
	botID = strings.TrimSpace(botID)
	primaryBotID = strings.TrimSpace(primaryBotID)
	if primaryBotID == "" {
		primaryBotID = ModeratorBotID
	}
	if botID == primaryBotID {
		return strings.TrimSpace(t.Moderator)
	}
	if t.Participants != nil {
		return strings.TrimSpace(t.Participants[botID])
	}
	return ""
}

func (t DiscordBotTokens) TokenFor(botID string) string {
	return t.TokenForBot(botID, ModeratorBotID)
}

func (t DiscordBotTokens) IsConfiguredFor(botID, primaryBotID string) bool {
	return t.TokenForBot(botID, primaryBotID) != ""
}

func (t DiscordBotTokens) IsConfigured(botID string) bool {
	return t.IsConfiguredFor(botID, ModeratorBotID)
}

// MaskSecretToken returns a length-proportional asterisk mask for API responses.
func MaskSecretToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	n := len(token)
	if n < 12 {
		n = 12
	}
	if n > 32 {
		n = 32
	}
	return strings.Repeat("*", n)
}

func applyDiscordBotTokens(cfg *Config, overrides map[string]string) {
	tokens := discordBotTokensFromOverrides(overrides)
	primaryID := effectivePrimaryBotID(overrides)
	if tok := tokens.TokenForBot(primaryID, primaryID); tok != "" {
		cfg.Secrets.DiscordBotToken = tok
	}
	if len(tokens.Participants) > 0 {
		cfg.Secrets.DiscordParticipantTokens = tokens.Participants
	} else {
		cfg.Secrets.DiscordParticipantTokens = nil
	}
}

func effectiveDiscordBotTokens(cfg Config, overrides map[string]string) DiscordBotTokens {
	tokens := discordBotTokensFromOverrides(overrides)
	if tokens.Moderator == "" && cfg.Secrets.DiscordBotToken != "" {
		tokens.Moderator = cfg.Secrets.DiscordBotToken
	}
	if tokens.Participants == nil {
		tokens.Participants = make(map[string]string)
	}
	if len(tokens.Participants) == 0 && len(cfg.Secrets.DiscordParticipantTokens) > 0 {
		for id, tok := range cfg.Secrets.DiscordParticipantTokens {
			tokens.Participants[id] = tok
		}
	}
	return tokens
}

// HostBotToken returns the token of the configured host bot from SQLite overrides.
func HostBotToken(cfg Config, overrides map[string]string) string {
	tokens := effectiveDiscordBotTokens(cfg, overrides)
	primary := effectivePrimaryBotID(overrides)
	return strings.TrimSpace(tokens.TokenForBot(primary, primary))
}

// BotShouldHandleCommandsForToken reports whether the connected bot token is the current host.
func BotShouldHandleCommandsForToken(botToken string, cfg Config, overrides map[string]string) bool {
	botToken = strings.TrimSpace(botToken)
	hostToken := HostBotToken(cfg, overrides)
	return botToken != "" && hostToken != "" && botToken == hostToken
}

// EffectiveDiscordBotTokens merges SQLite token overrides with loaded config secrets.
func EffectiveDiscordBotTokens(cfg Config, overrides map[string]string) DiscordBotTokens {
	return effectiveDiscordBotTokens(cfg, overrides)
}

func mergeDiscordBotTokenUpdates(
	current DiscordBotTokens,
	primaryBotID string,
	moderatorToken string,
	moderatorRoleToken string,
	participantTokens map[string]string,
	activeParticipantIDs []string,
) DiscordBotTokens {
	if current.Participants == nil {
		current.Participants = make(map[string]string)
	}
	primaryBotID = strings.TrimSpace(primaryBotID)
	if primaryBotID == "" {
		primaryBotID = ModeratorBotID
	}
	if strings.TrimSpace(moderatorToken) != "" {
		current.Moderator = strings.TrimSpace(moderatorToken)
	}
	if strings.TrimSpace(moderatorRoleToken) != "" {
		current.Participants[ModeratorBotID] = strings.TrimSpace(moderatorRoleToken)
	}
	for id, tok := range participantTokens {
		id = strings.TrimSpace(id)
		if id == "" || id == primaryBotID {
			if id == primaryBotID && strings.TrimSpace(tok) != "" {
				current.Moderator = strings.TrimSpace(tok)
			}
			continue
		}
		if strings.TrimSpace(tok) != "" {
			current.Participants[id] = strings.TrimSpace(tok)
		}
	}
	active := make(map[string]bool, len(activeParticipantIDs))
	for _, id := range activeParticipantIDs {
		active[id] = true
	}
	for id := range current.Participants {
		if id == ModeratorBotID {
			continue
		}
		if !active[id] {
			delete(current.Participants, id)
		}
	}
	delete(current.Participants, primaryBotID)
	return current
}

// migrateDiscordTokensFromEnv imports deploy/.env bot tokens into SQLite once.
func migrateDiscordTokensFromEnv(overrides map[string]string, cfg Config) map[string]string {
	raw, exists := overrides[DiscordBotTokensSetting]
	if exists && strings.TrimSpace(raw) != "" && raw != "{}" {
		return nil
	}
	tokens := DiscordBotTokens{Participants: make(map[string]string)}
	hasAny := false

	if v := strings.TrimSpace(os.Getenv(ModeratorEnvKey)); v != "" {
		tokens.Moderator = v
		hasAny = true
	} else if v := strings.TrimSpace(cfg.Secrets.DiscordBotToken); v != "" {
		tokens.Moderator = v
		hasAny = true
	}

	for _, id := range splitCSV(cfg.Transport.Discord.ParticipantBots) {
		envKey := ParticipantBotEnvKey(id)
		if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
			tokens.Participants[id] = v
			hasAny = true
		}
	}
	if !hasAny {
		return nil
	}
	return map[string]string{DiscordBotTokensSetting: formatDiscordBotTokensJSON(tokens)}
}
