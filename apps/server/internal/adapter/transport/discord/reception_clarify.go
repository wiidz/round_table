package discord

import (
	"context"
	"strings"
	"sync"

	"round_table/apps/server/internal/adapter/transport"
)

type receptionClarifySession struct {
	channelID string
	authorID  string
	tool      receptionTool
	partial   receptionDecision
}

type receptionClarifySessions struct {
	mu        sync.Mutex
	byChannel map[string]receptionClarifySession
}

func (s *receptionClarifySessions) put(channelID string, sess receptionClarifySession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byChannel == nil {
		s.byChannel = make(map[string]receptionClarifySession)
	}
	s.byChannel[channelID] = sess
}

func (s *receptionClarifySessions) get(channelID string) (receptionClarifySession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.byChannel[channelID]
	return sess, ok
}

func (s *receptionClarifySessions) clear(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byChannel, channelID)
}

func (s *receptionClarifySessions) pending(channelID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.byChannel[channelID]
	return ok
}

// HandleClarifyFollowUp continues a mutating flow after Reception asked for missing fields.
func (r *Reception) HandleClarifyFollowUp(ctx context.Context, msg transport.Inbound) (string, error) {
	if !r.enabled() {
		return "", nil
	}
	sess, ok := r.clarifies.get(msg.ChannelID)
	if !ok {
		return "", nil
	}
	loc := r.loc()
	if msg.AuthorID != sess.authorID {
		return receptionClarifyNotOwnerText(loc), nil
	}
	body := strings.TrimSpace(msg.Content)
	if isExpertCancelTrigger(body) || isReceptionCancelTrigger(body) {
		r.clarifies.clear(msg.ChannelID)
		return receptionClarifyCancelledText(loc), nil
	}
	if reply, ok := tryPrincipalNatural(msg, r.Registry, r.Profiles, loc); ok {
		r.clarifies.clear(msg.ChannelID)
		return reply, nil
	}
	d := r.mergeClarifyReply(sess, body)
	r.clarifies.clear(msg.ChannelID)
	if isReceptionMutatingTool(d.Tool) {
		return r.execMutatingTool(ctx, msg, d)
	}
	return receptionFallbackClarifyText(loc), nil
}

func (r *Reception) mergeClarifyReply(sess receptionClarifySession, body string) receptionDecision {
	d := sess.partial
	d.Tool = sess.tool
	switch sess.tool {
	case receptionToolCreateParticipant:
		if strings.TrimSpace(d.DisplayName) == "" {
			d.DisplayName = extractCreateDisplayName(body)
		}
		if strings.TrimSpace(d.Expertise) == "" {
			if exp := extractExpertiseFromReply(body); exp != "" {
				d.Expertise = exp
			}
		}
	case receptionToolUpdateParticipant, receptionToolDeleteParticipant:
		if strings.TrimSpace(d.ParticipantRef) == "" && strings.TrimSpace(d.ParticipantID) == "" {
			d.ParticipantRef = strings.TrimSpace(body)
		}
	case receptionToolStartMeeting:
		if strings.TrimSpace(d.Topic) == "" {
			d.Topic = strings.TrimSpace(body)
		}
	case receptionToolUpdateParticipantProfile:
		if strings.TrimSpace(d.ProfileContent) == "" {
			d.ProfileContent = body
		}
		if strings.TrimSpace(d.ProfileFile) == "" {
			d.ProfileFile = detectProfileFileKeyword(body)
		}
	}
	return d
}

func (r *Reception) storeClarifySession(msg transport.Inbound, userText string, decision receptionDecision) {
	tool := decision.PendingTool
	if tool == "" || tool == receptionToolNone {
		tool = inferReceptionPendingTool(userText)
	}
	if tool == "" || tool == receptionToolNone {
		return
	}
	partial := decision
	partial.Tool = tool
	r.clarifies.put(msg.ChannelID, receptionClarifySession{
		channelID: msg.ChannelID,
		authorID:  msg.AuthorID,
		tool:      tool,
		partial:   partial,
	})
}

func inferReceptionPendingTool(userText string) receptionTool {
	s := strings.TrimSpace(userText)
	if s == "" {
		return receptionToolNone
	}
	if matchesPrincipalBindIntent(s) {
		return receptionToolNone
	}
	if matchesProfileUpdateIntent(s) {
		return receptionToolUpdateParticipantProfile
	}
	if matchesCreateExpertIntent(s) {
		return receptionToolCreateParticipant
	}
	if containsAnySubstring(s, "删除专家", "移除专家", "删掉专家") ||
		(strings.Contains(s, "专家") && containsAnySubstring(s, "删除", "移除", "删掉")) {
		return receptionToolDeleteParticipant
	}
	if containsAnySubstring(s, "修改专家", "编辑专家", "更新专家") ||
		(strings.Contains(s, "专家") && containsAnySubstring(s, "修改", "编辑", "更新")) {
		return receptionToolUpdateParticipant
	}
	if containsAnySubstring(s, "开会", "启动会议", "开始会议") && !parseNaturalMeetStartOK(s) {
		return receptionToolStartMeeting
	}
	return receptionToolNone
}

func matchesCreateExpertIntent(s string) bool {
	if matchesPrincipalBindIntent(s) {
		return false
	}
	if matchesProfileUpdateIntent(s) {
		return false
	}
	if containsAnySubstring(s, "新增专家", "创建专家", "添加专家", "新建专家", "加一个专家", "增加专家") {
		return true
	}
	return false
}

func parseNaturalMeetStartOK(s string) bool {
	_, ok := parseNaturalMeetStart(s)
	return ok
}

func containsAnySubstring(s string, parts ...string) bool {
	for _, p := range parts {
		if p != "" && strings.Contains(s, p) {
			return true
		}
	}
	return false
}

func extractCreateDisplayName(body string) string {
	body = strings.TrimSpace(body)
	if matchesPrincipalBindIntent(body) {
		return ""
	}
	for _, prefix := range []string{"新增专家", "创建专家", "添加专家", "新建专家", "增加专家", "加一个专家"} {
		if strings.HasPrefix(body, prefix) {
			if rest := trimExpertQuotes(strings.TrimSpace(strings.TrimPrefix(body, prefix))); rest != "" {
				return rest
			}
		}
	}
	for _, prefix := range []string{"新增", "创建", "添加", "新建", "增加"} {
		if strings.HasPrefix(body, prefix) {
			rest := trimExpertQuotes(strings.TrimSpace(strings.TrimPrefix(body, prefix)))
			if rest != "" && !strings.HasPrefix(rest, "专家") {
				return rest
			}
		}
	}
	return trimExpertQuotes(body)
}

func extractExpertiseFromReply(body string) string {
	for _, sep := range []string{"专长：", "专长:", "expertise:", "expertise："} {
		if idx := strings.Index(body, sep); idx >= 0 {
			return strings.TrimSpace(body[idx+len(sep):])
		}
	}
	return ""
}

func trimExpertQuotes(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'「」『』""''`)
	return strings.TrimSpace(s)
}

func receptionClarifyCancelledText(loc Locale) string {
	if loc == LocaleZH {
		return "↩️ 已取消。"
	}
	return "↩️ Cancelled."
}

func receptionClarifyNotOwnerText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 只有发起该操作的用户可以继续回复。"
	}
	return "⚠️ Only the user who started this action can continue."
}
