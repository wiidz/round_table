package discord

import (
	"context"
	"strings"
	"sync"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

type receptionConfirmAction int

const (
	receptionActionCreateParticipant receptionConfirmAction = iota
	receptionActionUpdateParticipant
	receptionActionDeleteParticipant
	receptionActionStartMeeting
	receptionActionUpdateParticipantProfile
)

type receptionConfirmSession struct {
	channelID        string
	authorID         string
	action           receptionConfirmAction
	participant      config.ParticipantRosterItem
	oldParticipantID string
	meetConfig       meetLaunchConfig
	profileFile      string
	profileContent   string
	profileGenerated bool
}

type receptionConfirmSessions struct {
	mu        sync.Mutex
	byChannel map[string]receptionConfirmSession
}

func (s *receptionConfirmSessions) put(channelID string, sess receptionConfirmSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byChannel == nil {
		s.byChannel = make(map[string]receptionConfirmSession)
	}
	s.byChannel[channelID] = sess
}

func (s *receptionConfirmSessions) get(channelID string) (receptionConfirmSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.byChannel[channelID]
	return sess, ok
}

func (s *receptionConfirmSessions) clear(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byChannel, channelID)
}

func (s *receptionConfirmSessions) pending(channelID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.byChannel[channelID]
	return ok
}

func (r *Reception) InputPhase(channelID string) ChannelInputPhase {
	if r != nil && r.confirms.pending(channelID) {
		return InputPhaseReceptionConfirm
	}
	if r != nil && r.clarifies.pending(channelID) {
		return InputPhaseReceptionClarify
	}
	return InputPhaseIdle
}

// HandleConfirmReply processes Principal 1/0 after a mutating reception preview.
func (r *Reception) HandleConfirmReply(ctx context.Context, msg transport.Inbound) (string, error) {
	if !r.enabled() {
		return "", nil
	}
	sess, ok := r.confirms.get(msg.ChannelID)
	if !ok {
		return "", nil
	}
	loc := r.loc()
	if msg.AuthorID != sess.authorID {
		return receptionConfirmNotOwnerText(loc), nil
	}
	body := strings.TrimSpace(msg.Content)
	if isReceptionCancelTrigger(body) {
		r.confirms.clear(msg.ChannelID)
		return receptionConfirmCancelledText(loc), nil
	}
	if !isExpertConfirmYes(body) {
		if isExpertConfirmNo(body) {
			r.confirms.clear(msg.ChannelID)
			return receptionConfirmCancelledText(loc), nil
		}
		return expertConfirmChoiceText(loc), nil
	}
	r.confirms.clear(msg.ChannelID)
	return r.executeConfirm(ctx, msg, sess)
}

func (r *Reception) executeConfirm(ctx context.Context, msg transport.Inbound, sess receptionConfirmSession) (string, error) {
	switch sess.action {
	case receptionActionCreateParticipant:
		if r.Participants == nil {
			return expertStorageRequiredText(r.loc()), nil
		}
		return r.Participants.executeCreate(ctx, sess.participant)
	case receptionActionUpdateParticipant:
		if r.Participants == nil {
			return expertStorageRequiredText(r.loc()), nil
		}
		return r.Participants.executeUpdate(ctx, sess.oldParticipantID, sess.participant)
	case receptionActionDeleteParticipant:
		if r.Participants == nil {
			return expertStorageRequiredText(r.loc()), nil
		}
		return r.Participants.executeDelete(ctx, sess.oldParticipantID)
	case receptionActionStartMeeting:
		if r.Meet == nil {
			return receptionNoMeetingText(r.loc()), nil
		}
		return r.Meet.launch(msg, sess.meetConfig)
	case receptionActionUpdateParticipantProfile:
		if r.Participants == nil {
			return expertStorageRequiredText(r.loc()), nil
		}
		return r.Participants.executeProfileUpdate(ctx, sess.participant.ID, sess.profileFile, sess.profileContent)
	default:
		return "", nil
	}
}

func (r *Reception) beginConfirm(msg transport.Inbound, sess receptionConfirmSession) (string, error) {
	if reply, busy := r.channelBusyForMutate(msg); busy {
		return reply, nil
	}
	sess.channelID = msg.ChannelID
	sess.authorID = msg.AuthorID
	r.confirms.put(msg.ChannelID, sess)
	return formatReceptionConfirmPreview(r.loc(), sess), nil
}

func (r *Reception) channelBusyForMutate(msg transport.Inbound) (string, bool) {
	loc := r.loc()
	if r.confirms.pending(msg.ChannelID) {
		return receptionConfirmPendingText(loc), true
	}
	if r.Participants != nil && r.Participants.setups.pending(msg.ChannelID) {
		return expertSetupBusyText(loc), true
	}
	if r.Meet != nil {
		if reply, ok := r.Meet.checkBeginSetup(msg, loc); ok {
			return reply, true
		}
	}
	return "", false
}

func (r *Reception) checkMutatingPrincipal(msg transport.Inbound) (string, bool) {
	if msg.Platform == "web" {
		if r.Meet != nil {
			if reply, ok := r.Meet.checkBeginSetup(msg, r.loc()); ok {
				return reply, true
			}
			return "", false
		}
		return WebChatMutatingUnavailableText(r.loc()), true
	}
	if r.Meet != nil {
		return r.Meet.checkBeginSetup(msg, r.loc())
	}
	if r.Registry == nil {
		return meetNeedBindText(r.loc()), true
	}
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	binding, ok := r.Registry.Get(scope)
	if !ok {
		return meetNeedBindText(r.loc()), true
	}
	if binding.ExternalID != msg.AuthorID {
		return meetNotScopePrincipalText(r.loc()), true
	}
	return "", false
}

func isReceptionCancelTrigger(content string) bool {
	s := strings.TrimSpace(content)
	if s == "" {
		return false
	}
	lower := strings.ToLower(normalizeASCIIForms(s))
	return matchExact(lower, "取消确认", "cancel confirm", "cancel reception")
}
