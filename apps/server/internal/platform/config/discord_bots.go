package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	ModeratorBotID              = "moderator"
	ModeratorBotLabel           = "主持人"
	ModeratorEnvKey             = "DISCORD_BOT_TOKEN"
	DiscordBotsSetting          = "ROUND_TABLE_DISCORD_BOTS"
	DiscordBotProfilesSetting   = "ROUND_TABLE_DISCORD_BOT_PROFILES"
	DiscordModeratorRoleSetting = "ROUND_TABLE_DISCORD_MODERATOR_ROLE"
)

// DiscordBotEntry is one participant bot keyed by Discord Application ID.
type DiscordBotEntry struct {
	ApplicationID string `json:"application_id"`
}

// DiscordBotState is returned by GET /api/settings for Discord bot configuration.
type DiscordBotState struct {
	ID                   string `json:"id"`
	Label                string `json:"label,omitempty"`
	DisplayName          string `json:"display_name,omitempty"`
	Primary              bool   `json:"primary"`
	Deletable            bool   `json:"deletable"`
	EnvKey               string `json:"env_key"`
	Configured           bool   `json:"configured"`
	RestartRequired      bool   `json:"restart_required"`
	DiscordApplicationID string `json:"discord_application_id,omitempty"`
	DiscordUsername      string `json:"discord_username,omitempty"`
	AvatarURL            string `json:"avatar_url,omitempty"`
	ProfileFetchedAt     string `json:"profile_fetched_at,omitempty"`
	TokenMasked          string `json:"token_masked,omitempty"`
	Token                string `json:"token,omitempty"`
	BoundParticipantID   string `json:"bound_participant_id,omitempty"`
}

// DiscordBotsUpdate is the body for PUT /api/settings/discord-bots.
type DiscordBotsUpdate struct {
	ModeratorToken              string            `json:"moderator_token,omitempty"`
	ModeratorRoleToken          string            `json:"moderator_role_token,omitempty"`
	ModeratorRoleID             string            `json:"moderator_role_id,omitempty"`
	ModeratorBoundParticipantID string            `json:"moderator_bound_participant_id,omitempty"`
	Participants                []DiscordBotInput `json:"participants"`
}

// DiscordBotInput is one participant bot in a save request.
type DiscordBotInput struct {
	ApplicationID      string `json:"application_id,omitempty"`
	Token              string `json:"token,omitempty"`
	BoundParticipantID string `json:"bound_participant_id,omitempty"`
	// Legacy fields accepted during transition.
	LegacyID    string `json:"id,omitempty"`
	LegacyLabel string `json:"label,omitempty"`
}

// ParticipantBotEnvKey returns the env var name for a legacy participant bot token.
func ParticipantBotEnvKey(participantID string) string {
	id := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(participantID), "-", "_"))
	return "DISCORD_BOT_TOKEN_" + id
}

func parseDiscordBotsJSON(raw string) ([]DiscordBotEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var entries []DiscordBotEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, fmt.Errorf("invalid discord bots json: %w", err)
	}
	out := make([]DiscordBotEntry, 0, len(entries))
	for _, e := range entries {
		appID := normalizeDiscordApplicationID(e.ApplicationID)
		if appID == "" {
			continue
		}
		out = append(out, DiscordBotEntry{ApplicationID: appID})
	}
	return out, nil
}

func formatDiscordBotsJSON(entries []DiscordBotEntry) string {
	if len(entries) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(entries)
	return string(b)
}

func discordBotsFromTransport(dc DiscordTransport) []DiscordBotEntry {
	var out []DiscordBotEntry
	for _, id := range splitCSV(dc.ParticipantBots) {
		if appID := normalizeDiscordApplicationID(id); appID != "" {
			out = append(out, DiscordBotEntry{ApplicationID: appID})
		}
	}
	return out
}

func splitCSV(raw string) []string {
	var out []string
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func effectivePrimaryBotID(overrides map[string]string) string {
	if overrides != nil {
		if id := strings.TrimSpace(overrides[DiscordModeratorRoleSetting]); id != "" {
			return id
		}
	}
	return ModeratorBotID
}

// EffectivePrimaryBotID returns the configured host bot id (application id or "moderator").
func EffectivePrimaryBotID(overrides map[string]string) string {
	return effectivePrimaryBotID(overrides)
}

// FilterDiscordParticipantBotIDs returns participant bot application ids excluding the host bot.
func FilterDiscordParticipantBotIDs(cfg Config, overrides map[string]string) []string {
	ids := DiscordBotApplicationIDs(cfg)
	primary := effectivePrimaryBotID(overrides)
	if primary == "" || primary == ModeratorBotID {
		return ids
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != primary {
			out = append(out, id)
		}
	}
	return out
}

func swapPrimaryBotTokens(tokens DiscordBotTokens, oldPrimary, newPrimary string) DiscordBotTokens {
	oldPrimary = strings.TrimSpace(oldPrimary)
	newPrimary = strings.TrimSpace(newPrimary)
	if oldPrimary == "" {
		oldPrimary = ModeratorBotID
	}
	if newPrimary == "" || oldPrimary == newPrimary {
		return tokens
	}
	if tokens.Participants == nil {
		tokens.Participants = make(map[string]string)
	}

	oldPrimaryToken := strings.TrimSpace(tokens.Moderator)
	newPrimaryToken := strings.TrimSpace(tokens.Participants[newPrimary])

	tokens.Moderator = newPrimaryToken
	if oldPrimaryToken != "" {
		tokens.Participants[oldPrimary] = oldPrimaryToken
	} else {
		delete(tokens.Participants, oldPrimary)
	}
	delete(tokens.Participants, newPrimary)
	return tokens
}

func validateModeratorRoleID(roleID string, entries []DiscordBotEntry) error {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" || roleID == ModeratorBotID {
		return nil
	}
	for _, e := range entries {
		if e.ApplicationID == roleID {
			return nil
		}
	}
	return fmt.Errorf("moderator role %q not found in participant bots", roleID)
}

func effectiveDiscordBots(cfg Config, overrides map[string]string) []DiscordBotEntry {
	if overrides != nil {
		if raw, ok := overrides[DiscordBotsSetting]; ok && strings.TrimSpace(raw) != "" {
			entries, err := parseDiscordBotsJSON(raw)
			if err == nil {
				return entries
			}
		}
	}
	return discordBotsFromTransport(cfg.Transport.Discord)
}

func DiscordBotApplicationIDs(cfg Config) []string {
	entries := effectiveDiscordBots(cfg, nil)
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.ApplicationID)
	}
	return out
}

func buildDiscordBotStates(cfg Config, overrides map[string]string) []DiscordBotState {
	participants := effectiveDiscordBots(cfg, overrides)
	tokens := effectiveDiscordBotTokens(cfg, overrides)
	primaryID := effectivePrimaryBotID(overrides)
	profileCache := discordBotProfilesFromOverrides(overrides)
	bindings := cfg.Transport.Discord.ParticipantIMBindings
	if bindings == nil {
		bindings = participantIMBindingsFromOverrides(overrides)
	}
	roster := ParticipantRosterFromConfig(cfg)
	rosterNames := make(map[string]string, len(roster))
	for _, item := range roster {
		name := strings.TrimSpace(item.DisplayName)
		if name == "" {
			name = item.ID
		}
		rosterNames[item.ID] = name
	}

	modAppID := moderatorDiscordApplicationID(profileCache)
	modBound := ""
	if modAppID != "" {
		modBound = participantForDiscordBot(bindings, modAppID)
	}
	modProfile := profileCache[ModeratorBotID]
	modDisplayName := rosterNames[modBound]
	if modDisplayName == "" {
		modDisplayName = ModeratorBotLabel
	}

	states := make([]DiscordBotState, 0, 1+len(participants))

	states = append(states, DiscordBotState{
		ID:                   ModeratorBotID,
		Label:                modDisplayName,
		DisplayName:          modDisplayName,
		Primary:              primaryID == ModeratorBotID,
		Deletable:            false,
		EnvKey:               DiscordBotTokensSetting,
		Configured:           tokens.IsConfiguredFor(ModeratorBotID, primaryID),
		RestartRequired:      true,
		TokenMasked:          MaskSecretToken(tokens.TokenForBot(ModeratorBotID, primaryID)),
		Token:                tokens.TokenForBot(ModeratorBotID, primaryID),
		DiscordApplicationID: modAppID,
		DiscordUsername:      modProfile.DiscordUsername,
		AvatarURL:            modProfile.AvatarURL,
		ProfileFetchedAt:     modProfile.FetchedAt,
		BoundParticipantID:   modBound,
	})

	for _, p := range participants {
		boundPID := participantForDiscordBot(bindings, p.ApplicationID)
		displayName := rosterNames[boundPID]
		label := displayName
		if label == "" {
			label = p.ApplicationID
		}
		states = append(states, DiscordBotState{
			ID:                   p.ApplicationID,
			Label:                label,
			DisplayName:          displayName,
			Primary:              primaryID == p.ApplicationID,
			Deletable:            true,
			EnvKey:               DiscordBotTokensSetting,
			Configured:           tokens.IsConfiguredFor(p.ApplicationID, primaryID),
			RestartRequired:      true,
			TokenMasked:          MaskSecretToken(tokens.TokenForBot(p.ApplicationID, primaryID)),
			Token:                tokens.TokenForBot(p.ApplicationID, primaryID),
			DiscordApplicationID: p.ApplicationID,
			BoundParticipantID:   boundPID,
		})
	}
	return states
}

func applyDiscordBots(cfg *Config, entries []DiscordBotEntry) error {
	ids := make([]string, 0, len(entries))
	seen := make(map[string]bool)
	for _, e := range entries {
		appID := strings.TrimSpace(e.ApplicationID)
		if appID == "" {
			return fmt.Errorf("application_id required")
		}
		if !discordApplicationIDPattern.MatchString(appID) {
			return fmt.Errorf("invalid application_id %q", appID)
		}
		if seen[appID] {
			return fmt.Errorf("duplicate application_id %q", appID)
		}
		seen[appID] = true
		ids = append(ids, appID)
	}

	cfg.Transport.Discord.ParticipantBots = strings.Join(ids, ",")
	return nil
}

func normalizeDiscordBotInputs(
	inputs []DiscordBotInput,
	profiles map[string]DiscordBotProfileCache,
	existingTokens map[string]string,
) ([]DiscordBotEntry, map[string]string, map[string]string, map[string]DiscordBotProfileCache, error) {
	entries := make([]DiscordBotEntry, 0, len(inputs))
	tokens := make(map[string]string)
	boundByApp := make(map[string]string)
	fetchedProfiles := make(map[string]DiscordBotProfileCache)
	seen := make(map[string]bool)

	for _, in := range inputs {
		token := strings.TrimSpace(in.Token)
		appID, profile, err := resolveDiscordApplicationID(
			in.ApplicationID,
			in.LegacyID,
			token,
			profiles,
			existingTokens,
		)
		if err != nil {
			if token == "" && strings.TrimSpace(in.BoundParticipantID) == "" {
				continue
			}
			return nil, nil, nil, nil, err
		}
		if seen[appID] {
			return nil, nil, nil, nil, fmt.Errorf("duplicate application_id %q", appID)
		}
		seen[appID] = true

		if profile.DiscordApplicationID == "" {
			profile.DiscordApplicationID = appID
		}
		fetchedProfiles[appID] = profile

		entries = append(entries, DiscordBotEntry{ApplicationID: appID})
		if pid := strings.TrimSpace(in.BoundParticipantID); pid != "" {
			boundByApp[appID] = pid
		}
		if token != "" {
			tokens[appID] = token
		}
	}
	return entries, tokens, boundByApp, fetchedProfiles, nil
}

func discordApplicationIDSet(entries []DiscordBotEntry) map[string]struct{} {
	out := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		out[e.ApplicationID] = struct{}{}
	}
	return out
}

func moderatorDiscordApplicationID(profileCache map[string]DiscordBotProfileCache) string {
	if cached, ok := profileCache[ModeratorBotID]; ok {
		return normalizeDiscordApplicationID(cached.DiscordApplicationID)
	}
	return ""
}
