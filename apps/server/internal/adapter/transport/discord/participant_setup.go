package discord

import (
	"sync"

	"round_table/apps/server/internal/platform/config"
)

type participantSetupMode int

const (
	participantSetupCreate participantSetupMode = iota
	participantSetupEdit
	participantSetupDelete
)

type participantSetupStep int

const (
	participantStepAskDisplayName participantSetupStep = iota
	participantStepAskID
	participantStepAskExpertise
	participantStepConfirmCreate
	participantStepAskEditFields
	participantStepConfirmEdit
	participantStepConfirmDelete
)

type participantSetupSession struct {
	channelID string
	authorID  string
	mode      participantSetupMode
	step      participantSetupStep
	draft     config.ParticipantRosterItem
	oldID     string
}

type participantSetupSessions struct {
	mu        sync.Mutex
	byChannel map[string]participantSetupSession
}

func (s *participantSetupSessions) put(channelID string, sess participantSetupSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byChannel == nil {
		s.byChannel = make(map[string]participantSetupSession)
	}
	s.byChannel[channelID] = sess
}

func (s *participantSetupSessions) get(channelID string) (participantSetupSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.byChannel[channelID]
	return sess, ok
}

func (s *participantSetupSessions) clear(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byChannel, channelID)
}

func (s *participantSetupSessions) pending(channelID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.byChannel[channelID]
	return ok
}
