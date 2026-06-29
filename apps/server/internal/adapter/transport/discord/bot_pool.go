package discord

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// BotPool routes outbound messages to participant-specific Discord bots.
type BotPool struct {
	Default              ChannelSender
	byID                 map[string]ChannelSender
	participantBotID     func(participantID string) string
	closer               []func() error
}

// PoolOptions configures participant bot connections.
type PoolOptions struct {
	Default          ChannelSender
	BotOpts          Options
	BotIDs           []string
	Mapping          map[string]string // legacy: bot_id -> env key name
	ResolveToken     func(botID string) string
	ParticipantBotID func(participantID string) string // participant -> bound bot id
	HostToken        string                            // skip pool entry when token matches host
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

// NewMappedBotPool returns a pool with pre-wired senders (e.g. browser chat transport).
func NewMappedBotPool(defaultSender ChannelSender, byID map[string]ChannelSender) *BotPool {
	cp := make(map[string]ChannelSender, len(byID))
	for id, sender := range byID {
		if id = strings.TrimSpace(id); id != "" && sender != nil {
			cp[id] = sender
		}
	}
	return &BotPool{
		Default: defaultSender,
		byID:    cp,
	}
}

// OpenBotPool connects send-only participant bots. Missing tokens are skipped (fallback to Default).
func OpenBotPool(opts PoolOptions) (*BotPool, error) {
	if opts.Default == nil {
		return nil, fmt.Errorf("discord: bot pool default sender required")
	}
	pool := &BotPool{
		Default:          opts.Default,
		byID:             make(map[string]ChannelSender),
		participantBotID: opts.ParticipantBotID,
	}

	botIDs := opts.BotIDs
	if len(botIDs) == 0 && len(opts.Mapping) > 0 {
		for id := range opts.Mapping {
			botIDs = append(botIDs, id)
		}
	}

	for _, botID := range botIDs {
		botID = strings.TrimSpace(botID)
		if botID == "" {
			continue
		}
		token := ""
		if opts.ResolveToken != nil {
			token = strings.TrimSpace(opts.ResolveToken(botID))
		}
		if token == "" {
			envKey := ParticipantBotEnvKey(botID)
			if opts.Mapping != nil {
				if mapped, ok := opts.Mapping[botID]; ok && strings.TrimSpace(mapped) != "" {
					envKey = strings.TrimSpace(mapped)
				}
			}
			token = strings.TrimSpace(os.Getenv(envKey))
		}
		if token == "" {
			log.Printf("discord: participant bot %q skipped — no token configured", botID)
			continue
		}
		if host := strings.TrimSpace(opts.HostToken); host != "" && token == host {
			log.Printf("discord: participant bot %q skipped — same token as host bot", botID)
			continue
		}
		bot, err := New(Options{
			Token:      token,
			AllowDM:    opts.BotOpts.AllowDM,
			AllowGuild: opts.BotOpts.AllowGuild,
			GuildID:    opts.BotOpts.GuildID,
			SendOnly:   true,
		})
		if err != nil {
			return nil, fmt.Errorf("discord: participant bot %q: %w", botID, err)
		}
		pool.byID[botID] = bot
		pool.closer = append(pool.closer, bot.Close)
		log.Printf("discord participant bot ready (send-only) id=%q", botID)
	}
	return pool, nil
}

func (p *BotPool) boundBotID(participantID string) string {
	if p == nil || p.participantBotID == nil {
		return strings.TrimSpace(participantID)
	}
	return strings.TrimSpace(p.participantBotID(participantID))
}

// SenderFor returns the bot for a participant, or Default when unmapped.
func (p *BotPool) SenderFor(participantID string) ChannelSender {
	if p == nil {
		return nil
	}
	botID := p.boundBotID(participantID)
	if botID == "" {
		botID = strings.TrimSpace(participantID)
	}
	if botID != "" {
		if s, ok := p.byID[botID]; ok {
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
	botID := p.boundBotID(participantID)
	if botID == "" {
		botID = strings.TrimSpace(participantID)
	}
	if botID == "" {
		return false
	}
	_, ok := p.byID[botID]
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
