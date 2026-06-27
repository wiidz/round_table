package config

import (
	"context"
	"fmt"
	"strings"
)

// CreateParticipant adds a roster entry and persists meet_participants only.
func (s *Service) CreateParticipant(ctx context.Context, item ParticipantRosterItem) error {
	item.ID = strings.TrimSpace(item.ID)
	item.DisplayName = strings.TrimSpace(item.DisplayName)
	item.Expertise = strings.TrimSpace(item.Expertise)
	imBindings := item.IMBindings

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store == nil {
		return fmt.Errorf("participant roster requires app_settings storage")
	}

	roster := ParticipantRosterFromConfig(s.cfg)
	if rosterIndex(roster, item.ID) >= 0 {
		return fmt.Errorf("代号 %q 已存在", item.ID)
	}
	roster = append(roster, ParticipantRosterItem{
		ID:          item.ID,
		DisplayName: item.DisplayName,
		Expertise:   item.Expertise,
	})
	if err := s.persistParticipantRosterLocked(ctx, roster); err != nil {
		return err
	}

	bindings := s.effectiveParticipantIMBindingsLocked()
	setParticipantIMBindings(bindings, item.ID, imBindings)
	return s.persistParticipantIMBindingsLocked(ctx, bindings)
}

// UpdateParticipant updates metadata and optionally renames codename (oldID → item.ID).
func (s *Service) UpdateParticipant(ctx context.Context, oldID string, item ParticipantRosterItem) error {
	oldID = strings.TrimSpace(oldID)
	item.ID = strings.TrimSpace(item.ID)
	item.DisplayName = strings.TrimSpace(item.DisplayName)
	item.Expertise = strings.TrimSpace(item.Expertise)
	imBindings := item.IMBindings
	hasIMBindings := item.IMBindings != nil

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store == nil {
		return fmt.Errorf("participant roster requires app_settings storage")
	}

	roster := ParticipantRosterFromConfig(s.cfg)
	idx := rosterIndex(roster, oldID)
	if idx < 0 {
		return fmt.Errorf("专家 %q 不存在", oldID)
	}

	if item.ID != oldID {
		if rosterIndex(roster, item.ID) >= 0 {
			return fmt.Errorf("代号 %q 已存在", item.ID)
		}
	}

	roster[idx] = ParticipantRosterItem{
		ID:          item.ID,
		DisplayName: item.DisplayName,
		Expertise:   item.Expertise,
	}
	if err := s.persistParticipantRosterLocked(ctx, roster); err != nil {
		return err
	}

	if hasIMBindings || item.ID != oldID {
		bindings := s.effectiveParticipantIMBindingsLocked()
		if item.ID != oldID {
			renameParticipantIMBindings(bindings, oldID, item.ID)
		}
		if hasIMBindings {
			setParticipantIMBindings(bindings, item.ID, imBindings)
		}
		if err := s.persistParticipantIMBindingsLocked(ctx, bindings); err != nil {
			return err
		}
	}
	return nil
}

// DeleteParticipant removes a roster entry from meet_participants.
func (s *Service) DeleteParticipant(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store == nil {
		return fmt.Errorf("participant roster requires app_settings storage")
	}

	roster := ParticipantRosterFromConfig(s.cfg)
	idx := rosterIndex(roster, id)
	if idx < 0 {
		return fmt.Errorf("专家 %q 不存在", id)
	}
	if len(roster) <= 1 {
		return fmt.Errorf("至少保留一位专家")
	}

	roster = append(roster[:idx], roster[idx+1:]...)
	if err := s.persistParticipantRosterLocked(ctx, roster); err != nil {
		return err
	}

	bindings := s.effectiveParticipantIMBindingsLocked()
	deleteParticipantIMBindings(bindings, id)
	return s.persistParticipantIMBindingsLocked(ctx, bindings)
}

func (s *Service) persistParticipantRosterLocked(ctx context.Context, roster []ParticipantRosterItem) error {
	next := s.cfg
	if err := applyMeetParticipants(&next, roster); err != nil {
		return err
	}

	if s.store != nil {
		if err := s.store.SetSettings(ctx, map[string]string{
			MeetParticipantsSetting: next.Transport.Discord.MeetParticipants,
		}); err != nil {
			return err
		}
	}

	s.cfg = next
	return nil
}

// ParticipantIMBindingsView returns effective bindings for API enrichment.
func (s *Service) ParticipantIMBindingsView() ParticipantIMBindings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.effectiveParticipantIMBindingsLocked()
}
