package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

const MeetCastsSetting = "ROUND_TABLE_MEET_CASTS"

// MeetCastConfig is a named participant subset for Discord meet setup.
type MeetCastConfig struct {
	ID             string   `json:"id"`
	NameZH         string   `json:"name_zh"`
	NameEN         string   `json:"name_en"`
	ParticipantIDs []string `json:"participant_ids"`
}

func formatMeetCastsJSON(casts []MeetCastConfig) string {
	if len(casts) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(casts)
	return string(b)
}

func parseMeetCastsJSON(raw string) ([]MeetCastConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var out []MeetCastConfig
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("invalid meet casts json: %w", err)
	}
	return out, nil
}

func normalizeMeetCasts(in []MeetCastConfig, roster []ParticipantRosterItem) ([]MeetCastConfig, error) {
	rosterIDs := rosterIDSet(roster)
	out := make([]MeetCastConfig, 0, len(in))
	seenID := make(map[string]struct{})
	for _, c := range in {
		id := strings.TrimSpace(c.ID)
		if id == "" {
			return nil, fmt.Errorf("meet cast id required")
		}
		if _, ok := seenID[id]; ok {
			return nil, fmt.Errorf("duplicate meet cast id %q", id)
		}
		seenID[id] = struct{}{}
		nameZH := strings.TrimSpace(c.NameZH)
		nameEN := strings.TrimSpace(c.NameEN)
		ids := make([]string, 0, len(c.ParticipantIDs))
		seenPID := make(map[string]struct{})
		for _, pid := range c.ParticipantIDs {
			pid = strings.TrimSpace(pid)
			if pid == "" {
				continue
			}
			if _, ok := rosterIDs[pid]; !ok {
				return nil, fmt.Errorf("meet cast %q references unknown participant %q", id, pid)
			}
			if _, ok := seenPID[pid]; ok {
				continue
			}
			seenPID[pid] = struct{}{}
			ids = append(ids, pid)
		}
		if len(ids) == 0 {
			return nil, fmt.Errorf("meet cast %q requires at least one participant", id)
		}
		if nameZH == "" && nameEN == "" {
			nameZH = meetCastAutoName(ids, roster)
			nameEN = nameZH
		}
		if nameZH == "" && nameEN == "" {
			return nil, fmt.Errorf("meet cast %q requires name_zh or name_en", id)
		}
		if nameZH == "" {
			nameZH = nameEN
		}
		if nameEN == "" {
			nameEN = nameZH
		}
		out = append(out, MeetCastConfig{
			ID:             id,
			NameZH:         nameZH,
			NameEN:         nameEN,
			ParticipantIDs: ids,
		})
	}
	return out, nil
}

func meetCastAutoName(ids []string, roster []ParticipantRosterItem) string {
	byID := make(map[string]string, len(roster))
	for _, item := range roster {
		name := strings.TrimSpace(item.DisplayName)
		if name == "" {
			name = item.ID
		}
		byID[item.ID] = name
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if name, ok := byID[id]; ok {
			parts = append(parts, name)
		} else {
			parts = append(parts, id)
		}
	}
	return strings.Join(parts, "+")
}

func meetCastsFromOverrides(overrides map[string]string, cfg Config) []MeetCastConfig {
	if overrides == nil {
		return nil
	}
	raw, ok := overrides[MeetCastsSetting]
	if !ok || strings.TrimSpace(raw) == "" {
		return nil
	}
	parsed, err := parseMeetCastsJSON(raw)
	if err != nil {
		return nil
	}
	roster := ParticipantRosterFromConfig(cfg)
	normalized, err := normalizeMeetCasts(parsed, roster)
	if err != nil {
		return nil
	}
	return normalized
}

func applyMeetCastsToConfig(cfg *Config, casts []MeetCastConfig) {
	cfg.Meeting.MeetCasts = casts
}
