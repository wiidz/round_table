package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const discordAPIBase = "https://discord.com/api/v10"

type discordUserProfile struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type discordApplicationProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DiscordBotProfileCache is persisted profile metadata for one bot (keyed by bot id).
type DiscordBotProfileCache struct {
	DiscordApplicationID string `json:"discord_application_id,omitempty"`
	DiscordUsername      string `json:"discord_username,omitempty"`
	AvatarURL            string `json:"avatar_url,omitempty"`
	FetchedAt            string `json:"fetched_at,omitempty"`
}

func parseDiscordBotProfilesJSON(raw string) map[string]DiscordBotProfileCache {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]DiscordBotProfileCache{}
	}
	var out map[string]DiscordBotProfileCache
	if err := json.Unmarshal([]byte(raw), &out); err != nil || out == nil {
		return map[string]DiscordBotProfileCache{}
	}
	return out
}

func formatDiscordBotProfilesJSON(cache map[string]DiscordBotProfileCache) string {
	if len(cache) == 0 {
		return "{}"
	}
	b, _ := json.Marshal(cache)
	return string(b)
}

func discordBotProfilesFromOverrides(overrides map[string]string) map[string]DiscordBotProfileCache {
	if overrides == nil {
		return map[string]DiscordBotProfileCache{}
	}
	return parseDiscordBotProfilesJSON(overrides[DiscordBotProfilesSetting])
}

func applyCachedDiscordBotProfiles(states []DiscordBotState, cache map[string]DiscordBotProfileCache) []DiscordBotState {
	if len(cache) == 0 {
		return states
	}
	out := make([]DiscordBotState, len(states))
	copy(out, states)
	for i := range out {
		cached, ok := cache[out[i].ID]
		if !ok {
			continue
		}
		out[i].DiscordApplicationID = cached.DiscordApplicationID
		out[i].DiscordUsername = cached.DiscordUsername
		out[i].AvatarURL = cached.AvatarURL
		out[i].ProfileFetchedAt = cached.FetchedAt
	}
	return out
}

func pruneDiscordBotProfilesCache(cache map[string]DiscordBotProfileCache, activeIDs map[string]bool) {
	for id := range cache {
		if !activeIDs[id] {
			delete(cache, id)
		}
	}
}

func activeDiscordBotIDs(states []DiscordBotState) map[string]bool {
	out := make(map[string]bool, len(states))
	for _, s := range states {
		out[s.ID] = true
	}
	return out
}

func fetchDiscordBotProfile(token string) DiscordBotProfileCache {
	token = strings.TrimSpace(token)
	if token == "" {
		return DiscordBotProfileCache{}
	}

	client := &http.Client{Timeout: 4 * time.Second}
	auth := "Bot " + token

	var user discordUserProfile
	if discordAPIGet(client, auth, "/users/@me", &user) {
		avatar := discordAvatarURL(user.ID, user.Avatar)
		var app discordApplicationProfile
		appID := ""
		if discordAPIGet(client, auth, "/oauth2/applications/@me", &app) {
			appID = strings.TrimSpace(app.ID)
		}
		return DiscordBotProfileCache{
			DiscordApplicationID: appID,
			DiscordUsername:      strings.TrimSpace(user.Username),
			AvatarURL:            avatar,
		}
	}
	return DiscordBotProfileCache{}
}

// ApplicationIDFromToken resolves the Discord Application snowflake for a bot token.
func ApplicationIDFromToken(token string) string {
	return applicationIDFromProfile(fetchDiscordBotProfile(token))
}

// HostApplicationID returns the application id of the configured host bot.
func HostApplicationID(overrides map[string]string) string {
	primary := effectivePrimaryBotID(overrides)
	if primary != ModeratorBotID {
		return normalizeDiscordApplicationID(primary)
	}
	cache := discordBotProfilesFromOverrides(overrides)
	if cached, ok := cache[ModeratorBotID]; ok {
		return normalizeDiscordApplicationID(cached.DiscordApplicationID)
	}
	return ""
}

// BotShouldHandleCommands reports whether a connected bot application should handle commands.
// Deprecated: prefer BotShouldHandleCommandsForToken for host checks.
func BotShouldHandleCommands(botAppID string, overrides map[string]string) bool {
	botAppID = normalizeDiscordApplicationID(botAppID)
	hostAppID := HostApplicationID(overrides)
	if botAppID == "" || hostAppID == "" {
		return false
	}
	return botAppID == hostAppID
}

func discordAPIGet(client *http.Client, auth, path string, dest any) bool {
	req, err := http.NewRequest(http.MethodGet, discordAPIBase+path, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", auth)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return false
	}
	return json.Unmarshal(body, dest) == nil
}

func fetchDiscordBotProfiles(states []DiscordBotState, tokens DiscordBotTokens, primaryBotID string) map[string]DiscordBotProfileCache {
	now := time.Now().UTC().Format(time.RFC3339)
	out := make(map[string]DiscordBotProfileCache)
	for _, state := range states {
		if !state.Configured {
			continue
		}
		token := tokens.TokenForBot(state.ID, primaryBotID)
		if token == "" {
			continue
		}
		profile := fetchDiscordBotProfile(token)
		if profile.DiscordApplicationID == "" && profile.DiscordUsername == "" {
			continue
		}
		profile.FetchedAt = now
		out[state.ID] = profile
	}
	return out
}

func discordAvatarURL(userID, avatarHash string) string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ""
	}
	if hash := strings.TrimSpace(avatarHash); hash != "" {
		return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png?size=128", userID, hash)
	}
	index := (parseDiscordSnowflake(userID) >> 22) % 6
	return fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", index)
}

func parseDiscordSnowflake(id string) uint64 {
	n, _ := strconv.ParseUint(id, 10, 64)
	return n
}
