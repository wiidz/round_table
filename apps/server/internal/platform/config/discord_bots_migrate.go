package config

import (
	"encoding/json"
	"strings"
)

const DiscordBotsMigrationV2Setting = "ROUND_TABLE_DISCORD_BOTS_MIGRATION_V2"

type legacyDiscordBotEntry struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	ApplicationID string `json:"application_id"`
}

func parseLegacyDiscordBotsJSON(raw string) []legacyDiscordBotEntry {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var entries []legacyDiscordBotEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil
	}
	return entries
}

// migrateDiscordBotsToApplicationIDs rekeys bot roster, tokens, profiles, and IM bindings by application_id.
func migrateDiscordBotsToApplicationIDs(overrides map[string]string) map[string]string {
	if overrides == nil {
		return nil
	}
	if strings.TrimSpace(overrides[DiscordBotsMigrationV2Setting]) == "done" {
		return nil
	}

	legacyEntries := parseLegacyDiscordBotsJSON(overrides[DiscordBotsSetting])
	if len(legacyEntries) == 0 {
		return nil
	}

	profiles := discordBotProfilesFromOverrides(overrides)
	tokens := discordBotTokensFromOverrides(overrides)
	bindings := participantIMBindingsFromOverrides(overrides)

	oldToApp := make(map[string]string)
	var newEntries []DiscordBotEntry
	seenApp := make(map[string]bool)

	for _, legacy := range legacyEntries {
		appID, _, err := resolveDiscordApplicationID(
			legacy.ApplicationID,
			legacy.ID,
			tokens.Participants[strings.TrimSpace(legacy.ID)],
			profiles,
			tokens.Participants,
		)
		if err != nil || appID == "" {
			continue
		}
		oldID := strings.TrimSpace(legacy.ID)
		if oldID != "" {
			oldToApp[oldID] = appID
		}
		if seenApp[appID] {
			continue
		}
		seenApp[appID] = true
		newEntries = append(newEntries, DiscordBotEntry{ApplicationID: appID})
	}

	newTokens := DiscordBotTokens{Participants: make(map[string]string)}
	if tok := strings.TrimSpace(tokens.Moderator); tok != "" {
		newTokens.Moderator = tok
	}
	for oldID, tok := range tokens.Participants {
		appID := oldToApp[oldID]
		if appID == "" {
			if appID = normalizeDiscordApplicationID(oldID); appID == "" {
				continue
			}
		}
		if strings.TrimSpace(tok) != "" {
			newTokens.Participants[appID] = strings.TrimSpace(tok)
		}
	}

	newProfiles := make(map[string]DiscordBotProfileCache)
	for oldID, profile := range profiles {
		appID := oldToApp[oldID]
		if appID == "" {
			appID = normalizeDiscordApplicationID(oldID)
			if appID == "" {
				appID = applicationIDFromProfile(profile)
			}
		}
		if appID == "" {
			continue
		}
		if profile.DiscordApplicationID == "" {
			profile.DiscordApplicationID = appID
		}
		newProfiles[appID] = profile
	}

	newBindings := make(ParticipantIMBindings)
	for pid, binds := range bindings {
		var next []ParticipantIMBind
		for _, bind := range binds {
			if bind.Platform != IMPlatformDiscord {
				next = append(next, bind)
				continue
			}
			appID := normalizeDiscordApplicationID(bind.ApplicationID)
			if appID == "" {
				appID = oldToApp[strings.TrimSpace(bind.ApplicationID)]
			}
			if appID == "" {
				continue
			}
			next = append(next, ParticipantIMBind{Platform: IMPlatformDiscord, ApplicationID: appID})
		}
		if len(next) > 0 {
			newBindings[pid] = next
		}
	}

	newPrimary := strings.TrimSpace(overrides[DiscordModeratorRoleSetting])
	if newPrimary != "" && newPrimary != ModeratorBotID {
		if mapped, ok := oldToApp[newPrimary]; ok {
			newPrimary = mapped
		} else if appID := normalizeDiscordApplicationID(newPrimary); appID != "" {
			newPrimary = appID
		}
	}

	out := map[string]string{
		DiscordBotsSetting:                formatDiscordBotsJSON(newEntries),
		DiscordBotTokensSetting:           formatDiscordBotTokensJSON(newTokens),
		DiscordBotProfilesSetting:         formatDiscordBotProfilesJSON(newProfiles),
		ParticipantIMBindingsSetting:    formatParticipantIMBindingsJSON(newBindings),
		DiscordBotsMigrationV2Setting:     "done",
	}
	if newPrimary != "" {
		out[DiscordModeratorRoleSetting] = newPrimary
	}
	return out
}
