package config

import "strings"

// ParticipantRosterEntry is parsed metadata from meet_participants (id:Role[:Expertise]).
type ParticipantRosterEntry struct {
	DisplayName string
	Expertise   string
}

// MeetParticipantIDs parses id from meet_participants roster (id:Role[:Expertise], …).
func MeetParticipantIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var ids []string
	seen := make(map[string]bool)
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		id := item
		if i := strings.Index(item, ":"); i > 0 {
			id = strings.TrimSpace(item[:i])
		}
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}

// ParseMeetParticipants maps participant id → display name and expertise.
func ParseMeetParticipants(raw string) map[string]ParticipantRosterEntry {
	out := make(map[string]ParticipantRosterEntry)
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		first := strings.Index(item, ":")
		if first <= 0 {
			continue
		}
		id := strings.TrimSpace(item[:first])
		rest := strings.TrimSpace(item[first+1:])
		entry := ParticipantRosterEntry{DisplayName: rest}
		if last := strings.LastIndex(rest, ":"); last > 0 {
			entry.DisplayName = strings.TrimSpace(rest[:last])
			entry.Expertise = strings.TrimSpace(rest[last+1:])
		}
		if entry.DisplayName == "" {
			entry.DisplayName = id
		}
		out[id] = entry
	}
	return out
}
