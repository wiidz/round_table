package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"

	prin "round_table/apps/server/internal/adapter/principal"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

// ChannelPrincipal implements principal.Port for Discord Confirmation (ADR-0004).
type ChannelPrincipal struct {
	Bots *BotPool
	Loc  Locale

	mu       sync.Mutex
	sessions map[string]*meetingBind // meetingID → bind state
}

type meetingBind struct {
	channelID string
	authorID  string

	confirmCycle        int
	confirmLimitFallback bool
	confirmReply        chan confirmReply

	pendingRunning prin.RunningIntervention
	paused         bool
	pausedReply    chan prin.RunningIntervention

	inFreeDialogue              bool
	pendingPrincipalQuestion    string
	pendingPrincipalAnswerer    string
}

type confirmReply struct {
	resp prin.Response
	err  error
}

// NewChannelPrincipal returns a Principal port that blocks Confirm until the bound Discord user replies.
func NewChannelPrincipal(bots *BotPool, locale string) *ChannelPrincipal {
	return &ChannelPrincipal{
		Bots:     bots,
		Loc:      ParseLocale(locale),
		sessions: make(map[string]*meetingBind),
	}
}

// BindMeeting registers the Discord channel and Principal for a running meeting.
func (p *ChannelPrincipal) BindMeeting(meetingID, channelID, authorID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessions[meetingID] = &meetingBind{
		channelID:    channelID,
		authorID:     authorID,
		confirmReply: make(chan confirmReply, 1),
		pausedReply:  make(chan prin.RunningIntervention, 1),
	}
}

// UnbindMeeting removes a meeting binding when the run finishes.
func (p *ChannelPrincipal) UnbindMeeting(meetingID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.sessions, meetingID)
}

// PendingConfirmation reports whether a channel is waiting for Principal confirmation.
func (p *ChannelPrincipal) PendingConfirmation(channelID string) (meetingID string, authorID string, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, w := range p.sessions {
		if w.channelID == channelID && w.confirmCycle > 0 {
			return id, w.authorID, true
		}
	}
	return "", "", false
}

// PendingPaused reports whether a channel meeting is paused waiting for resume/abort.
func (p *ChannelPrincipal) PendingPaused(channelID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, w := range p.sessions {
		if w.channelID == channelID && w.paused {
			return true
		}
	}
	return false
}

// DeliverConfirmationReply parses a Principal message and unblocks Confirm.
func (p *ChannelPrincipal) DeliverConfirmationReply(channelID, authorID, content string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	wait := p.bindForChannelLocked(channelID)
	if wait == nil || wait.confirmCycle == 0 {
		return "", nil
	}
	if authorID != wait.authorID {
		return confirmNotOwnerText(p.Loc), nil
	}

	resp, err := parseConfirmationReplyForBind(content, wait.confirmLimitFallback)
	if err != nil {
		if wait.confirmLimitFallback {
			return confirmLimitParseErrorText(p.Loc, err), nil
		}
		return confirmParseErrorText(p.Loc, err), nil
	}

	select {
	case wait.confirmReply <- confirmReply{resp: resp}:
	default:
		return confirmAlreadyAnsweredText(p.Loc), nil
	}
	if wait.confirmLimitFallback {
		return confirmLimitReceivedText(p.Loc, resp.Decision), nil
	}
	if resp.Decision == prin.DecisionApproved {
		return confirmReceivedApproveText(p.Loc), nil
	}
	return confirmReceivedRejectText(p.Loc), nil
}

// MarkFreeDialogue updates whether a channel meeting is in the free-dialogue phase.
func (p *ChannelPrincipal) MarkFreeDialogue(channelID string, active bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if wait := p.bindForChannelLocked(channelID); wait != nil {
		wait.inFreeDialogue = active
		if !active {
			wait.pendingPrincipalQuestion = ""
		}
	}
}

// InFreeDialogue reports whether the channel's meeting is in free dialogue.
func (p *ChannelPrincipal) InFreeDialogue(channelID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	wait := p.bindForChannelLocked(channelID)
	return wait != nil && wait.inFreeDialogue
}

// DeliverFreeDialogueQuestion queues a Principal question during free dialogue.
func (p *ChannelPrincipal) DeliverFreeDialogueQuestion(channelID, authorID, content string) (string, error) {
	question, answererID, ok := parseFreeDialogueQuestion(content)
	if !ok {
		return "", nil
	}
	if strings.TrimSpace(question) == "" {
		return freeDialogueQuestionParseErrorText(p.Loc, errFreeDialogueQuestionEmpty), nil
	}

	p.mu.Lock()
	wait := p.bindForChannelLocked(channelID)
	if wait == nil {
		p.mu.Unlock()
		return "", nil
	}
	if authorID != wait.authorID {
		p.mu.Unlock()
		return freeDialogueQuestionNotOwnerText(p.Loc), nil
	}
	if wait.confirmCycle > 0 {
		p.mu.Unlock()
		return freeDialogueQuestionConfirmBlocksText(p.Loc), nil
	}
	if !wait.inFreeDialogue {
		p.mu.Unlock()
		return freeDialogueQuestionWrongPhaseText(p.Loc), nil
	}
	if wait.pendingPrincipalQuestion != "" {
		p.mu.Unlock()
		return freeDialogueQuestionAlreadyQueuedText(p.Loc), nil
	}
	wait.pendingPrincipalQuestion = strings.TrimSpace(question)
	wait.pendingPrincipalAnswerer = strings.TrimSpace(answererID)
	p.mu.Unlock()
	return freeDialogueQuestionAckText(p.Loc, strings.TrimSpace(question), answererID), nil
}

// FreeDialogueQuestion implements principal.Port.
func (p *ChannelPrincipal) FreeDialogueQuestion(_ context.Context, meetingID string, _ meeting.State) (prin.FreeDialogueQuestionRequest, bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	wait, ok := p.sessions[meetingID]
	if !ok || wait.pendingPrincipalQuestion == "" {
		return prin.FreeDialogueQuestionRequest{}, false, nil
	}
	req := prin.FreeDialogueQuestionRequest{
		Question:   wait.pendingPrincipalQuestion,
		AnswererID: wait.pendingPrincipalAnswerer,
	}
	wait.pendingPrincipalQuestion = ""
	wait.pendingPrincipalAnswerer = ""
	return req, true, nil
}

// DeliverIntervention handles Principal running/paused control commands.
func (p *ChannelPrincipal) DeliverIntervention(channelID, authorID, content string) (string, error) {
	p.mu.Lock()
	wait := p.bindForChannelLocked(channelID)
	if wait == nil {
		p.mu.Unlock()
		return "", nil
	}
	if authorID != wait.authorID {
		p.mu.Unlock()
		return interventionNotOwnerText(p.Loc), nil
	}
	if wait.confirmCycle > 0 {
		p.mu.Unlock()
		return interventionConfirmBlocksText(p.Loc), nil
	}

	action, ok := parseIntervention(content)
	if !ok {
		p.mu.Unlock()
		return interventionParseErrorText(p.Loc, errInterventionUnrecognized), nil
	}

	if wait.paused {
		if action.Kind != prin.RunningInterventionResume && action.Kind != prin.RunningInterventionAbort {
			p.mu.Unlock()
			return interventionParseErrorText(p.Loc, fmt.Errorf("会议已暂停，请发送恢复或终止")), nil
		}
		select {
		case wait.pausedReply <- action:
		default:
			p.mu.Unlock()
			return interventionAlreadyQueuedText(p.Loc), nil
		}
		p.mu.Unlock()
		return interventionAckText(p.Loc, action.Kind), nil
	}

	if action.Kind == prin.RunningInterventionResume {
		p.mu.Unlock()
		return interventionNoMeetingText(p.Loc), nil
	}
	if wait.pendingRunning.Kind != "" {
		p.mu.Unlock()
		return interventionAlreadyQueuedText(p.Loc), nil
	}
	wait.pendingRunning = action
	p.mu.Unlock()
	return interventionAckText(p.Loc, action.Kind), nil
}

// Confirm implements principal.Port.
func (p *ChannelPrincipal) Confirm(ctx context.Context, meetingID string, brief event.ConfirmationBrief, cycle int) (prin.Response, error) {
	p.mu.Lock()
	wait, ok := p.sessions[meetingID]
	if !ok {
		p.mu.Unlock()
		return prin.Response{}, fmt.Errorf("discord: no channel binding for meeting %s", meetingID)
	}
	wait.confirmCycle = cycle
	wait.confirmLimitFallback = brief.LimitFallback
	channelID := wait.channelID
	replyCh := wait.confirmReply
	p.mu.Unlock()

	if p.Bots == nil || p.Bots.Default == nil {
		return prin.Response{}, fmt.Errorf("discord: bot sender not configured")
	}

	body := formatConfirmationBrief(p.Loc, meetingID, cycle, brief)
	SendLong(p.Bots.Default, ctx, channelID, body)

	select {
	case <-ctx.Done():
		return prin.Response{}, ctx.Err()
	case r := <-replyCh:
		p.mu.Lock()
		if w, ok := p.sessions[meetingID]; ok {
			w.confirmCycle = 0
			w.confirmLimitFallback = false
		}
		p.mu.Unlock()
		if r.err != nil {
			return prin.Response{}, r.err
		}
		return r.resp, nil
	}
}

// RunningAction implements principal.Port.
func (p *ChannelPrincipal) RunningAction(_ context.Context, meetingID string, _ meeting.State) (prin.RunningIntervention, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	wait, ok := p.sessions[meetingID]
	if !ok {
		return prin.RunningIntervention{}, nil
	}
	action := wait.pendingRunning
	if action.Kind == "" {
		return prin.RunningIntervention{}, nil
	}
	wait.pendingRunning = prin.RunningIntervention{}
	return action, nil
}

// PausedAction implements principal.Port.
func (p *ChannelPrincipal) PausedAction(ctx context.Context, meetingID string, _ meeting.State) (prin.RunningIntervention, error) {
	p.mu.Lock()
	wait, ok := p.sessions[meetingID]
	if !ok {
		p.mu.Unlock()
		return prin.RunningIntervention{}, fmt.Errorf("discord: no channel binding for meeting %s", meetingID)
	}
	wait.paused = true
	channelID := wait.channelID
	replyCh := wait.pausedReply
	p.mu.Unlock()

	if p.Bots != nil && p.Bots.Default != nil {
		_ = p.Bots.Default.Send(ctx, channelID, formatPausedWaitPrompt(p.Loc))
	}

	select {
	case <-ctx.Done():
		return prin.RunningIntervention{}, ctx.Err()
	case action := <-replyCh:
		p.mu.Lock()
		if w, ok := p.sessions[meetingID]; ok {
			w.paused = false
		}
		p.mu.Unlock()
		return action, nil
	}
}

func parseConfirmationReplyForBind(content string, limitFallback bool) (prin.Response, error) {
	if limitFallback {
		return parseConfirmationLimitReply(content)
	}
	return parseConfirmationReply(content)
}

func (p *ChannelPrincipal) bindForChannelLocked(channelID string) *meetingBind {
	for _, w := range p.sessions {
		if w.channelID == channelID {
			return w
		}
	}
	return nil
}
