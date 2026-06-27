package config

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	participantIDReserved       = "moderator"
	defaultParticipantExpertise = "general"
	// MeetParticipantsSetting persists expert roster (id:display:expertise) independently of IM bots.
	MeetParticipantsSetting = "ROUND_TABLE_MEET_PARTICIPANTS"
)

var participantIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// ParticipantRosterItem is one configured meeting expert (persisted via meet_participants).
type ParticipantRosterItem struct {
	ID          string              `json:"id"`
	DisplayName string              `json:"display_name"`
	Expertise   string              `json:"expertise,omitempty"`
	IMBindings  []ParticipantIMBind `json:"im_bindings,omitempty"`
}

// ParticipantRosterFromConfig builds roster items from effective transport config.
func ParticipantRosterFromConfig(cfg Config) []ParticipantRosterItem {
	meta := ParseMeetParticipants(cfg.Transport.Discord.MeetParticipants)
	seen := make(map[string]bool)
	var out []ParticipantRosterItem

	appendID := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		entry := meta[id]
		name := strings.TrimSpace(entry.DisplayName)
		if name == "" {
			name = id
		}
		exp := strings.TrimSpace(entry.Expertise)
		if exp == "" {
			exp = defaultParticipantExpertise
		}
		out = append(out, ParticipantRosterItem{
			ID:          id,
			DisplayName: name,
			Expertise:   exp,
		})
	}

	for _, id := range MeetParticipantIDs(cfg.Transport.Discord.MeetParticipants) {
		appendID(id)
	}
	return out
}

// ValidateParticipantID checks codename pattern and reserved ids.
func ValidateParticipantID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("代号不能为空")
	}
	if id == participantIDReserved || id == ModeratorBotID {
		return fmt.Errorf("代号 %q 为系统保留", id)
	}
	if !participantIDPattern.MatchString(id) {
		return fmt.Errorf("代号须以小写字母开头，仅含小写字母、数字、_、-")
	}
	if utf8.RuneCountInString(id) > 32 {
		return fmt.Errorf("代号不能超过 32 个字符")
	}
	return nil
}

// ValidateParticipantDisplayName checks display name length.
func ValidateParticipantDisplayName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("名称不能为空")
	}
	if utf8.RuneCountInString(name) > 40 {
		return fmt.Errorf("名称不能超过 40 个字符")
	}
	return nil
}

func normalizeParticipantNameKey(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if isParticipantASCII(name) {
		return strings.ToLower(name)
	}
	return name
}

func isParticipantASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

func validateParticipantRoster(roster []ParticipantRosterItem) error {
	ids := make(map[string]struct{}, len(roster))
	names := make(map[string]string, len(roster))
	for _, item := range roster {
		if err := ValidateParticipantID(item.ID); err != nil {
			return err
		}
		if err := ValidateParticipantDisplayName(item.DisplayName); err != nil {
			return fmt.Errorf("专家 %s: %w", item.ID, err)
		}
		if _, ok := ids[item.ID]; ok {
			return fmt.Errorf("代号 %q 重复", item.ID)
		}
		ids[item.ID] = struct{}{}

		nameKey := normalizeParticipantNameKey(item.DisplayName)
		if prev, ok := names[nameKey]; ok {
			return fmt.Errorf("名称 %q 与 %s 重复", item.DisplayName, prev)
		}
		names[nameKey] = item.ID

		exp := strings.TrimSpace(item.Expertise)
		if exp == "" {
			continue
		}
		if utf8.RuneCountInString(exp) > 32 {
			return fmt.Errorf("专家 %s: 专长不能超过 32 个字符", item.ID)
		}
	}
	return nil
}

func applyParticipantRoster(cfg *Config, roster []ParticipantRosterItem) error {
	return applyMeetParticipants(cfg, roster)
}

func applyMeetParticipants(cfg *Config, roster []ParticipantRosterItem) error {
	if err := validateParticipantRoster(roster); err != nil {
		return err
	}
	var meetParts []string
	for _, item := range roster {
		id := strings.TrimSpace(item.ID)
		label := strings.TrimSpace(item.DisplayName)
		if label == "" {
			label = id
		}
		exp := strings.TrimSpace(item.Expertise)
		if exp == "" {
			exp = defaultParticipantExpertise
		}
		meetParts = append(meetParts, fmt.Sprintf("%s:%s:%s", id, label, exp))
	}
	cfg.Transport.Discord.MeetParticipants = strings.Join(meetParts, ",")
	return nil
}

func meetParticipantsFromOverrides(overrides map[string]string, cfg Config) string {
	if overrides != nil {
		if raw, ok := overrides[MeetParticipantsSetting]; ok && strings.TrimSpace(raw) != "" {
			return strings.TrimSpace(raw)
		}
	}
	return cfg.Transport.Discord.MeetParticipants
}

func rosterIndex(roster []ParticipantRosterItem, id string) int {
	for i, item := range roster {
		if item.ID == id {
			return i
		}
	}
	return -1
}

