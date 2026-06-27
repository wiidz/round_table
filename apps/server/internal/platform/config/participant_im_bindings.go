package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	ParticipantIMBindingsSetting = "ROUND_TABLE_PARTICIPANT_IM_BINDINGS"
	IMPlatformDiscord            = "discord"
)

// ParticipantIMBind is one expert → IM platform bot link (Discord uses application_id).
type ParticipantIMBind struct {
	Platform        string `json:"platform"`
	ApplicationID   string `json:"application_id"`
}

func (b ParticipantIMBind) MarshalJSON() ([]byte, error) {
	type alias struct {
		Platform      string `json:"platform"`
		ApplicationID string `json:"application_id"`
	}
	return json.Marshal(alias{
		Platform:      b.Platform,
		ApplicationID: strings.TrimSpace(b.ApplicationID),
	})
}

func (b *ParticipantIMBind) UnmarshalJSON(data []byte) error {
	var raw struct {
		Platform      string `json:"platform"`
		ApplicationID string `json:"application_id"`
		BotID         string `json:"bot_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	b.Platform = raw.Platform
	appID := normalizeDiscordApplicationID(raw.ApplicationID)
	if appID == "" {
		appID = normalizeDiscordApplicationID(raw.BotID)
	}
	if appID == "" {
		appID = strings.TrimSpace(raw.BotID)
	}
	b.ApplicationID = appID
	return nil
}

// ParticipantIMBindings maps participant id → platform bindings (expert 1-to-many IM).
type ParticipantIMBindings map[string][]ParticipantIMBind

func parseParticipantIMBindingsJSON(raw string) ParticipantIMBindings {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return make(ParticipantIMBindings)
	}
	var out ParticipantIMBindings
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return make(ParticipantIMBindings)
	}
	if out == nil {
		return make(ParticipantIMBindings)
	}
	return out
}

func formatParticipantIMBindingsJSON(bindings ParticipantIMBindings) string {
	if len(bindings) == 0 {
		return "{}"
	}
	b, _ := json.Marshal(bindings)
	return string(b)
}

func participantIMBindingsFromOverrides(overrides map[string]string) ParticipantIMBindings {
	if overrides == nil {
		return make(ParticipantIMBindings)
	}
	return parseParticipantIMBindingsJSON(overrides[ParticipantIMBindingsSetting])
}

func (b ParticipantIMBindings) clone() ParticipantIMBindings {
	out := make(ParticipantIMBindings, len(b))
	for pid, binds := range b {
		cp := make([]ParticipantIMBind, len(binds))
		copy(cp, binds)
		out[pid] = cp
	}
	return out
}

func normalizeIMPlatform(platform string) (string, error) {
	platform = strings.ToLower(strings.TrimSpace(platform))
	switch platform {
	case IMPlatformDiscord:
		return IMPlatformDiscord, nil
	case "":
		return "", fmt.Errorf("IM 平台不能为空")
	default:
		return "", fmt.Errorf("暂不支持 IM 平台 %q", platform)
	}
}

// DiscordBotForParticipant returns the Discord application id bound to an expert.
func DiscordBotForParticipant(bindings ParticipantIMBindings, participantID string) string {
	return discordApplicationForParticipant(bindings, participantID)
}

func discordApplicationForParticipant(bindings ParticipantIMBindings, participantID string) string {
	participantID = strings.TrimSpace(participantID)
	for _, bind := range bindings[participantID] {
		if bind.Platform == IMPlatformDiscord {
			return strings.TrimSpace(bind.ApplicationID)
		}
	}
	return ""
}

func participantForDiscordBot(bindings ParticipantIMBindings, applicationID string) string {
	ids := participantsForDiscordBot(bindings, applicationID)
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

func participantsForDiscordBot(bindings ParticipantIMBindings, applicationID string) []string {
	applicationID = strings.TrimSpace(applicationID)
	if applicationID == "" {
		return nil
	}
	var out []string
	for pid, binds := range bindings {
		for _, bind := range binds {
			if bind.Platform == IMPlatformDiscord && strings.TrimSpace(bind.ApplicationID) == applicationID {
				out = append(out, pid)
				break
			}
		}
	}
	return out
}

func setParticipantIMBindings(bindings ParticipantIMBindings, participantID string, binds []ParticipantIMBind) {
	participantID = strings.TrimSpace(participantID)
	if participantID == "" {
		return
	}
	normalized := normalizeParticipantIMBinds(binds)
	if len(normalized) == 0 {
		delete(bindings, participantID)
		return
	}
	bindings[participantID] = normalized
	enforceDiscordBotOwnership(bindings, participantID)
}

func enforceDiscordBotOwnership(bindings ParticipantIMBindings, ownerParticipantID string) {
	ownerParticipantID = strings.TrimSpace(ownerParticipantID)
	discordApp := discordApplicationForParticipant(bindings, ownerParticipantID)
	if discordApp == "" {
		return
	}
	for pid, binds := range bindings {
		if pid == ownerParticipantID {
			continue
		}
		next := binds[:0]
		for _, bind := range binds {
			if bind.Platform == IMPlatformDiscord && strings.TrimSpace(bind.ApplicationID) == discordApp {
				continue
			}
			next = append(next, bind)
		}
		if len(next) == 0 {
			delete(bindings, pid)
		} else {
			bindings[pid] = next
		}
	}
}

func normalizeParticipantIMBinds(binds []ParticipantIMBind) []ParticipantIMBind {
	byPlatform := make(map[string]string)
	for _, bind := range binds {
		platform, err := normalizeIMPlatform(bind.Platform)
		if err != nil {
			continue
		}
		appID := strings.TrimSpace(bind.ApplicationID)
		if appID == "" {
			delete(byPlatform, platform)
			continue
		}
		byPlatform[platform] = appID
	}
	if len(byPlatform) == 0 {
		return nil
	}
	out := make([]ParticipantIMBind, 0, len(byPlatform))
	if appID, ok := byPlatform[IMPlatformDiscord]; ok {
		out = append(out, ParticipantIMBind{Platform: IMPlatformDiscord, ApplicationID: appID})
	}
	return out
}

func setDiscordBotBinding(bindings ParticipantIMBindings, applicationID, participantID string) {
	assignDiscordPair(bindings, applicationID, participantID)
}

func assignDiscordPair(bindings ParticipantIMBindings, applicationID, participantID string) {
	applicationID = strings.TrimSpace(applicationID)
	participantID = strings.TrimSpace(participantID)
	if applicationID == "" {
		return
	}

	for pid, binds := range bindings {
		next := binds[:0]
		for _, bind := range binds {
			if bind.Platform == IMPlatformDiscord && strings.TrimSpace(bind.ApplicationID) == applicationID {
				continue
			}
			next = append(next, bind)
		}
		if len(next) == 0 {
			delete(bindings, pid)
		} else {
			bindings[pid] = next
		}
	}

	if participantID == "" {
		return
	}

	binds := bindings[participantID]
	next := binds[:0]
	for _, bind := range binds {
		if bind.Platform == IMPlatformDiscord {
			continue
		}
		next = append(next, bind)
	}
	next = append(next, ParticipantIMBind{Platform: IMPlatformDiscord, ApplicationID: applicationID})
	bindings[participantID] = next
}

func validateParticipantIMBindings(bindings ParticipantIMBindings, rosterIDs, applicationIDs map[string]struct{}) error {
	for pid, binds := range bindings {
		if _, ok := rosterIDs[pid]; !ok {
			return fmt.Errorf("绑定专家 %q 不在名录中", pid)
		}
		seenPlatform := make(map[string]bool)
		for _, bind := range binds {
			platform, err := normalizeIMPlatform(bind.Platform)
			if err != nil {
				return fmt.Errorf("专家 %s: %w", pid, err)
			}
			if seenPlatform[platform] {
				return fmt.Errorf("专家 %s: 平台 %q 重复绑定", pid, platform)
			}
			seenPlatform[platform] = true
			appID := strings.TrimSpace(bind.ApplicationID)
			if appID == "" {
				return fmt.Errorf("专家 %s: %q 绑定缺少 Bot", pid, platform)
			}
			if appID == ModeratorBotID {
				return fmt.Errorf("专家 %s: 不可绑定主持人 Bot", pid)
			}
			if _, ok := applicationIDs[appID]; !ok {
				return fmt.Errorf("专家 %s: Bot %q 不存在", pid, appID)
			}
		}
	}
	return validateDiscordBindingUniqueness(bindings)
}

func validateDiscordBindingUniqueness(bindings ParticipantIMBindings) error {
	botOwner := make(map[string]string)
	for pid, binds := range bindings {
		for _, bind := range binds {
			if bind.Platform != IMPlatformDiscord {
				continue
			}
			appID := strings.TrimSpace(bind.ApplicationID)
			if appID == "" {
				continue
			}
			if prev, ok := botOwner[appID]; ok && prev != pid {
				return fmt.Errorf("Discord Bot %q 已绑定专家 %q", appID, prev)
			}
			botOwner[appID] = pid
		}
	}
	return nil
}

func rosterIDSet(roster []ParticipantRosterItem) map[string]struct{} {
	out := make(map[string]struct{}, len(roster))
	for _, item := range roster {
		out[item.ID] = struct{}{}
	}
	return out
}

func migrateLegacyParticipantIMBindings(cfg Config) ParticipantIMBindings {
	bindings := make(ParticipantIMBindings)
	appSet := make(map[string]bool)
	for _, id := range splitCSV(cfg.Transport.Discord.ParticipantBots) {
		appSet[normalizeDiscordApplicationID(id)] = true
	}
	for _, pid := range MeetParticipantIDs(cfg.Transport.Discord.MeetParticipants) {
		if appSet[pid] {
			bindings[pid] = []ParticipantIMBind{{Platform: IMPlatformDiscord, ApplicationID: pid}}
		}
	}
	return bindings
}

func renameParticipantIMBindings(bindings ParticipantIMBindings, oldID, newID string) {
	oldID = strings.TrimSpace(oldID)
	newID = strings.TrimSpace(newID)
	if oldID == "" || newID == "" || oldID == newID {
		return
	}
	if binds, ok := bindings[oldID]; ok {
		delete(bindings, oldID)
		bindings[newID] = binds
	}
}

func deleteParticipantIMBindings(bindings ParticipantIMBindings, participantID string) {
	delete(bindings, strings.TrimSpace(participantID))
}
