package httptransport

import (
	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/platform/config"
)

func (h *Handler) enrichParticipantIndex(p profile.ParticipantIndex) profile.ParticipantIndex {
	roster := h.participantRoster()
	if entry, ok := roster[p.ID]; ok {
		p.InRoster = true
		if entry.DisplayName != "" {
			p.DisplayName = entry.DisplayName
		}
		if entry.Expertise != "" {
			p.Expertise = entry.Expertise
		}
	}
	if p.DisplayName == "" {
		p.DisplayName = p.ID
	}
	p.IMBindings = h.participantIMBindings()[p.ID]
	return p
}

func (h *Handler) enrichParticipantDetail(d profile.ParticipantDetail) profile.ParticipantDetail {
	idx := h.enrichParticipantIndex(profile.ParticipantIndex{ID: d.ID})
	d.DisplayName = idx.DisplayName
	d.Expertise = idx.Expertise
	d.IMBindings = idx.IMBindings
	return d
}

func (h *Handler) participantRoster() map[string]config.ParticipantRosterEntry {
	if h.config == nil {
		return nil
	}
	return config.ParseMeetParticipants(h.config.Current().Transport.Discord.MeetParticipants)
}

func (h *Handler) participantIMBindings() config.ParticipantIMBindings {
	if h.config == nil {
		return nil
	}
	return h.config.ParticipantIMBindingsView()
}
