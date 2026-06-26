package discord

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// BotPool routes outbound messages to participant-specific Discord bots.
type BotPool struct {
	Default ChannelSender
	byID    map[string]ChannelSender
	closer  []func() error
}

// PoolOptions configures participant bot connections.
type PoolOptions struct {
	Default   ChannelSender
	BotOpts   Options
	Mapping   map[string]string // participant_id -> env var name for token
}

// ParticipantBotEnvKey returns the default env var for a participant bot token.
func ParticipantBotEnvKey(participantID string) string {
	id := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(participantID), "-", "_"))
	return "DISCORD_BOT_TOKEN_" + id
}

// ParseParticipantBotMapping parses `id` or `id:ENV_VAR` comma-separated entries.
func ParseParticipantBotMapping(raw string) map[string]string {
	out := make(map[string]string)
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		id, envKey, ok := strings.Cut(item, ":")
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if !ok {
			out[id] = ParticipantBotEnvKey(id)
			continue
		}
		envKey = strings.TrimSpace(envKey)
		if envKey == "" {
			envKey = ParticipantBotEnvKey(id)
		}
		out[id] = envKey
	}
	return out
}

// OpenBotPool connects send-only participant bots. Missing tokens are skipped (fallback to Default).
func OpenBotPool(opts PoolOptions) (*BotPool, error) {
	if opts.Default == nil {
		return nil, fmt.Errorf("discord: bot pool default sender required")
	}
	pool := &BotPool{
		Default: opts.Default,
		byID:    make(map[string]ChannelSender),
	}
	for id, envKey := range opts.Mapping {
		token := strings.TrimSpace(os.Getenv(envKey))
		if token == "" {
			log.Printf("discord: participant bot %q skipped — set %s in .env", id, envKey)
			continue
		}
		bot, err := New(Options{
			Token:      token,
			AllowDM:    opts.BotOpts.AllowDM,
			AllowGuild: opts.BotOpts.AllowGuild,
			GuildID:    opts.BotOpts.GuildID,
		})
		if err != nil {
			return nil, fmt.Errorf("discord: participant bot %q: %w", id, err)
		}
		pool.byID[id] = bot
		name := bot.DisplayName()
		pool.closer = append(pool.closer, bot.Close)
		log.Printf("discord participant bot connected id=%q name=%q env=%s", id, name, envKey)
	}
	return pool, nil
}

// SenderFor returns the bot for a participant, or Default when unmapped.
func (p *BotPool) SenderFor(participantID string) ChannelSender {
	if p == nil {
		return nil
	}
	id := strings.TrimSpace(participantID)
	if id != "" {
		if s, ok := p.byID[id]; ok {
			return s
		}
	}
	return p.Default
}

// HasBot reports whether a dedicated sender exists for participantID.
func (p *BotPool) HasBot(participantID string) bool {
	if p == nil {
		return false
	}
	id := strings.TrimSpace(participantID)
	if id == "" {
		return false
	}
	_, ok := p.byID[id]
	return ok
}

// Count returns the number of connected participant bots (excluding Default).
func (p *BotPool) Count() int {
	if p == nil {
		return 0
	}
	return len(p.byID)
}

// Close disconnects participant bot sessions.
func (p *BotPool) Close() {
	if p == nil {
		return
	}
	for _, closeFn := range p.closer {
		if closeFn != nil {
			_ = closeFn()
		}
	}
}
