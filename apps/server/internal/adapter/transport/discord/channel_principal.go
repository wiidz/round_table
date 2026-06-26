package discord

import (
	"context"
	"fmt"
	"sync"

	prin "round_table/apps/server/internal/adapter/principal"
	prinstub "round_table/apps/server/internal/adapter/principal/stub"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

// ChannelPrincipal implements principal.Port for Discord Confirmation (ADR-0004).
type ChannelPrincipal struct {
	Bots *BotPool
	Loc  Locale

	mu       sync.Mutex
	sessions map[string]*confirmWait // meetingID → wait state
	stub     prinstub.Principal
}

type confirmWait struct {
	channelID string
	authorID  string
	cycle     int
	replyCh   chan confirmReply
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
		sessions: make(map[string]*confirmWait),
	}
}

// BindMeeting registers the Discord channel and Principal for a running meeting.
func (p *ChannelPrincipal) BindMeeting(meetingID, channelID, authorID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessions[meetingID] = &confirmWait{
		channelID: channelID,
		authorID:  authorID,
		replyCh:   make(chan confirmReply, 1),
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
		if w.channelID == channelID && w.cycle > 0 {
			return id, w.authorID, true
		}
	}
	return "", "", false
}

// DeliverConfirmationReply parses a Principal message and unblocks Confirm.
func (p *ChannelPrincipal) DeliverConfirmationReply(channelID, authorID, content string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var wait *confirmWait
	var meetingID string
	for id, w := range p.sessions {
		if w.channelID == channelID && w.cycle > 0 {
			wait = w
			meetingID = id
			break
		}
	}
	if wait == nil {
		return "", nil
	}
	if authorID != wait.authorID {
		return confirmNotOwnerText(p.Loc), nil
	}

	resp, err := parseConfirmationReply(content)
	if err != nil {
		return confirmParseErrorText(p.Loc, err), nil
	}

	select {
	case wait.replyCh <- confirmReply{resp: resp}:
	default:
		return confirmAlreadyAnsweredText(p.Loc), nil
	}
	_ = meetingID
	if resp.Decision == prin.DecisionApproved {
		return confirmReceivedApproveText(p.Loc), nil
	}
	return confirmReceivedRejectText(p.Loc), nil
}

// Confirm implements principal.Port.
func (p *ChannelPrincipal) Confirm(ctx context.Context, meetingID string, brief event.ConfirmationBrief, cycle int) (prin.Response, error) {
	p.mu.Lock()
	wait, ok := p.sessions[meetingID]
	if !ok {
		p.mu.Unlock()
		return prin.Response{}, fmt.Errorf("discord: no channel binding for meeting %s", meetingID)
	}
	wait.cycle = cycle
	channelID := wait.channelID
	replyCh := wait.replyCh
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
		if r.err != nil {
			return prin.Response{}, r.err
		}
		return r.resp, nil
	}
}

// RunningAction delegates to stub (CLI-style interventions not wired on Discord yet).
func (p *ChannelPrincipal) RunningAction(ctx context.Context, meetingID string, s meeting.State) (prin.RunningIntervention, error) {
	return p.stub.RunningAction(ctx, meetingID, s)
}

// PausedAction delegates to stub.
func (p *ChannelPrincipal) PausedAction(ctx context.Context, meetingID string, s meeting.State) (prin.RunningIntervention, error) {
	return p.stub.PausedAction(ctx, meetingID, s)
}
