package config

import (
	"fmt"
	"regexp"
	"strings"
)

var discordApplicationIDPattern = regexp.MustCompile(`^\d{17,20}$`)

// normalizeDiscordApplicationID returns a Discord Application snowflake or "".
func normalizeDiscordApplicationID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "app-") {
		raw = strings.TrimPrefix(raw, "app-")
	}
	if discordApplicationIDPattern.MatchString(raw) {
		return raw
	}
	return ""
}

func applicationIDFromProfile(profile DiscordBotProfileCache) string {
	return normalizeDiscordApplicationID(profile.DiscordApplicationID)
}

func IsDiscordApplicationID(id string) bool {
	return normalizeDiscordApplicationID(id) != ""
}

// IsMisplacedBotProfileID reports ids that are Discord application ids or legacy app-{id} profile dirs.
func IsMisplacedBotProfileID(id string) bool {
	id = strings.TrimSpace(id)
	if IsDiscordApplicationID(id) {
		return true
	}
	if strings.HasPrefix(id, "app-") {
		return IsDiscordApplicationID(strings.TrimPrefix(id, "app-"))
	}
	return false
}

func resolveDiscordApplicationID(
	inputApplicationID, legacyID, token string,
	profiles map[string]DiscordBotProfileCache,
	tokens map[string]string,
) (string, DiscordBotProfileCache, error) {
	if appID := normalizeDiscordApplicationID(inputApplicationID); appID != "" {
		return appID, lookupProfileByApplicationID(profiles, appID), nil
	}
	if appID := normalizeDiscordApplicationID(legacyID); appID != "" {
		return appID, lookupProfileByApplicationID(profiles, appID), nil
	}
	for _, legacyKey := range []string{legacyID, inputApplicationID} {
		legacyKey = strings.TrimSpace(legacyKey)
		if legacyKey == "" || legacyKey == ModeratorBotID {
			continue
		}
		if cached, ok := profiles[legacyKey]; ok {
			if appID := applicationIDFromProfile(cached); appID != "" {
				return appID, cached, nil
			}
		}
		if tok := strings.TrimSpace(tokens[legacyKey]); tok != "" {
			profile := fetchDiscordBotProfile(tok)
			if appID := applicationIDFromProfile(profile); appID != "" {
				return appID, profile, nil
			}
		}
	}
	if tok := strings.TrimSpace(token); tok != "" {
		profile := fetchDiscordBotProfile(tok)
		if appID := applicationIDFromProfile(profile); appID != "" {
			return appID, profile, nil
		}
		return "", DiscordBotProfileCache{}, fmt.Errorf("无法从 Token 识别 Discord Application ID，请检查 Token 或网络")
	}
	return "", DiscordBotProfileCache{}, fmt.Errorf("请填写 Bot Token 或提供 Application ID")
}

func lookupProfileByApplicationID(cache map[string]DiscordBotProfileCache, appID string) DiscordBotProfileCache {
	if cached, ok := cache[appID]; ok {
		return cached
	}
	for _, profile := range cache {
		if normalizeDiscordApplicationID(profile.DiscordApplicationID) == appID {
			return profile
		}
	}
	return DiscordBotProfileCache{DiscordApplicationID: appID}
}

func validateDiscordBotExpertInputs(inputs []DiscordBotInput) error {
	expertBot := make(map[string]string)
	botExpert := make(map[string]string)
	for i, in := range inputs {
		appID := strings.TrimSpace(in.ApplicationID)
		if appID == "" {
			appID = normalizeDiscordApplicationID(in.LegacyID)
		}
		if appID == "" {
			appID = fmt.Sprintf("slot:%d", i)
		}
		pid := strings.TrimSpace(in.BoundParticipantID)
		if pid == "" {
			continue
		}
		if prev, ok := expertBot[pid]; ok && prev != appID {
			return fmt.Errorf("专家 %q 不能绑定多个 Discord Bot（%q 与 %q）", pid, prev, appID)
		}
		expertBot[pid] = appID
		if prev, ok := botExpert[appID]; ok && prev != pid {
			return fmt.Errorf("Discord Bot %q 不能绑定多个专家（%q 与 %q）", appID, prev, pid)
		}
		botExpert[appID] = pid
	}
	return nil
}

func validateUniqueDiscordBotTokens(inputs []DiscordBotInput, tokens DiscordBotTokens, primaryBotID string) error {
	seen := make(map[string]string)
	for _, in := range inputs {
		appID := strings.TrimSpace(in.ApplicationID)
		if appID == "" {
			appID = normalizeDiscordApplicationID(in.LegacyID)
		}
		if appID == "" {
			continue
		}
		tok := strings.TrimSpace(in.Token)
		if tok == "" {
			tok = tokens.TokenForBot(appID, primaryBotID)
		}
		if tok == "" {
			continue
		}
		if prev, ok := seen[tok]; ok && prev != appID {
			return fmt.Errorf("Bot Token 重复（%q 与 %q）", prev, appID)
		}
		seen[tok] = appID
	}
	return nil
}
