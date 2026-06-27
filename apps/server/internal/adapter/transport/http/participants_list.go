package httptransport

import (
	"sort"

	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/platform/config"
)

func (h *Handler) listParticipantsForAPI() ([]profile.ParticipantIndex, error) {
	if h.config == nil {
		return h.profile.ListParticipants()
	}

	cfg := h.config.Current()
	rosterOrder := config.MeetParticipantIDs(cfg.Transport.Discord.MeetParticipants)

	diskByID := make(map[string]profile.ParticipantIndex)
	diskList, err := h.profile.ListParticipants()
	if err != nil {
		return nil, err
	}
	for _, p := range diskList {
		if config.IsMisplacedBotProfileID(p.ID) {
			continue
		}
		diskByID[p.ID] = p
	}

	out := make([]profile.ParticipantIndex, 0, len(rosterOrder)+len(diskByID))
	for _, id := range rosterOrder {
		if p, ok := diskByID[id]; ok {
			p.InRoster = true
			out = append(out, p)
			delete(diskByID, id)
			continue
		}
		if err := h.profile.EnsureParticipant(id); err != nil {
			return nil, err
		}
		p, err := h.participantIndexFromDisk(id)
		if err != nil {
			return nil, err
		}
		p.InRoster = true
		out = append(out, p)
	}

	extra := make([]profile.ParticipantIndex, 0, len(diskByID))
	for _, p := range diskByID {
		extra = append(extra, p)
	}
	sort.Slice(extra, func(i, j int) bool {
		return extra[i].ID < extra[j].ID
	})
	out = append(out, extra...)
	return out, nil
}

func (h *Handler) participantIndexFromDisk(id string) (profile.ParticipantIndex, error) {
	list, err := h.profile.ListParticipants()
	if err != nil {
		return profile.ParticipantIndex{}, err
	}
	for _, p := range list {
		if p.ID == id {
			return p, nil
		}
	}
	return profile.ParticipantIndex{ID: id}, nil
}

func (h *Handler) ensureConfiguredParticipants() {
	if h.config == nil {
		return
	}
	_ = h.profile.PruneMisplacedBotProfiles()
	cfg := h.config.Current()
	for _, id := range config.MeetParticipantIDs(cfg.Transport.Discord.MeetParticipants) {
		_ = h.profile.EnsureParticipant(id)
	}
}
