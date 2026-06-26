package discord

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/engine"
	"round_table/apps/server/internal/platform/bootstrap"
	"round_table/apps/server/internal/platform/config"
)

// ChannelSender posts outbound messages to a transport channel.
type ChannelSender interface {
	Send(ctx context.Context, channelID, content string) error
}

// MeetRunner starts Engine meetings from Discord commands.
type MeetRunner struct {
	Cfg       config.Config
	Discord   config.DiscordTransport
	Registry  *principalbind.Registry
	Bots      *BotPool
	Principal *ChannelPrincipal
	sessions  meetSessions
	setups    meetSetupSessions
}

type meetSessions struct {
	mu        sync.Mutex
	byChannel map[string]string
}

func (m *meetSessions) tryStart(channelID, meetingID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byChannel == nil {
		m.byChannel = make(map[string]string)
	}
	if existing, busy := m.byChannel[channelID]; busy {
		return fmt.Errorf("%w (meeting %s)", errChannelMeetingBusy, existing)
	}
	m.byChannel[channelID] = meetingID
	return nil
}

func (m *meetSessions) clear(channelID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.byChannel, channelID)
}

func (m *meetSessions) active(channelID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byChannel[channelID]
	return id, ok
}

// BeginSetupFromTrigger starts setup when Principal sends a natural-language trigger (no prefix).
func (r *MeetRunner) BeginSetupFromTrigger(msg transport.Inbound) (string, error) {
	loc := ParseLocale(r.Discord.Locale)
	if reply, ok := r.checkBeginSetup(msg, loc); ok {
		return reply, nil
	}

	cfg := r.defaultLaunchConfig("", "")
	r.setups.put(msg.ChannelID, meetSetupSession{
		channelID: msg.ChannelID,
		authorID:  msg.AuthorID,
		config:    cfg,
		step:      setupStepAskTopic,
	})
	return formatAskTopicPrompt(loc), nil
}

// BeginSetup asks the Principal to confirm or adjust meeting configuration.
func (r *MeetRunner) BeginSetup(msg transport.Inbound, parsed meetParseResult) (string, error) {
	loc := ParseLocale(r.Discord.Locale)
	if reply, ok := r.checkBeginSetup(msg, loc); ok {
		return reply, nil
	}

	cfg := r.defaultLaunchConfig(parsed.Topic, parsed.Mode)
	r.setups.put(msg.ChannelID, meetSetupSession{
		channelID: msg.ChannelID,
		authorID:  msg.AuthorID,
		config:    cfg,
		step:      setupStepPresetMenu,
	})
	prefix := strings.TrimSpace(r.Discord.CommandPrefix)
	if prefix == "" {
		prefix = "!rt"
	}
	return formatModeratorSetupPrompt(loc, prefix+" ", cfg), nil
}

func (r *MeetRunner) checkBeginSetup(msg transport.Inbound, loc Locale) (string, bool) {
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	binding, ok := r.Registry.Get(scope)
	if !ok {
		return meetNeedBindText(loc), true
	}
	if binding.ExternalID != msg.AuthorID {
		return meetNotScopePrincipalText(loc), true
	}
	if id, busy := r.sessions.active(msg.ChannelID); busy {
		return meetChannelBusyText(loc, id), true
	}
	if r.setups.pending(msg.ChannelID) {
		return meetSetupPendingText(loc), true
	}
	return "", false
}

// CancelSetup clears a pending meet configuration for a channel.
func (r *MeetRunner) CancelSetup(channelID, authorID string) (string, bool) {
	sess, ok := r.setups.get(channelID)
	if !ok {
		return "", false
	}
	if sess.authorID != authorID {
		loc := ParseLocale(r.Discord.Locale)
		return meetSetupNotOwnerText(loc), true
	}
	r.setups.clear(channelID)
	return meetSetupCancelledText(ParseLocale(r.Discord.Locale)), true
}

// HandleSetupReply processes a Principal reply while setup is pending.
func (r *MeetRunner) HandleSetupReply(msg transport.Inbound) (string, error) {
	loc := ParseLocale(r.Discord.Locale)
	sess, ok := r.setups.get(msg.ChannelID)
	if !ok {
		return "", nil
	}
	if msg.AuthorID != sess.authorID {
		return meetSetupNotOwnerText(loc), nil
	}

	if sess.step == setupStepAskTopic {
		topic := strings.TrimSpace(msg.Content)
		if topic == "" {
			return meetTopicRequiredText(loc), nil
		}
		defaultCfg := r.defaultLaunchConfig(topic, "")
		sess.config = defaultCfg
		sess.step = setupStepPresetMenu
		r.setups.put(msg.ChannelID, sess)
		prefix := strings.TrimSpace(r.Discord.CommandPrefix)
		if prefix == "" {
			prefix = "!rt"
		}
		return formatModeratorSetupPrompt(loc, prefix+" ", defaultCfg), nil
	}

	prefix := strings.TrimSpace(r.Discord.CommandPrefix)
	if prefix == "" {
		prefix = "!rt"
	}
	prefix = prefix + " "
	defaultCfg := r.defaultLaunchConfig(sess.config.Topic, "")

	result, err := handleSetupStep(sess, msg.Content, loc, prefix, defaultCfg)
	if err != nil {
		return meetSetupParseErrorText(loc, err), nil
	}
	if result.launch {
		r.setups.clear(msg.ChannelID)
		return r.launch(msg, result.config)
	}

	sess.config = result.config
	sess.step = result.step
	r.setups.put(msg.ChannelID, sess)
	return result.reply, nil
}

// launch starts the meeting asynchronously after configuration is confirmed.
func (r *MeetRunner) launch(msg transport.Inbound, cfg meetLaunchConfig) (string, error) {
	loc := ParseLocale(r.Discord.Locale)
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	binding, ok := r.Registry.Get(scope)
	if !ok {
		return meetNeedBindText(loc), nil
	}

	meetingID := fmt.Sprintf("mtg-dc-%d", time.Now().Unix())
	if err := r.sessions.tryStart(msg.ChannelID, meetingID); err != nil {
		if busy := extractBusyMeetingID(err); busy != "" {
			return meetChannelBusyText(loc, busy), nil
		}
		return err.Error(), nil
	}

	go r.runMeeting(msg.ChannelID, meetingID, binding, cfg)
	return formatMeetLaunchAck(loc, meetingID, cfg, binding.DisplayName), nil
}

// HandleConfirmationReply processes Principal confirmation while a meeting waits.
func (r *MeetRunner) HandleConfirmationReply(msg transport.Inbound) (string, error) {
	if r.Principal == nil {
		return "", nil
	}
	return r.Principal.DeliverConfirmationReply(msg.ChannelID, msg.AuthorID, msg.Content)
}

func (r *MeetRunner) runMeeting(channelID, meetingID string, binding principalbind.Binding, cfg meetLaunchConfig) {
	defer r.sessions.clear(channelID)
	ctx := context.Background()

	loc := ParseLocale(r.Discord.Locale)
	if r.Principal != nil {
		r.Principal.BindMeeting(meetingID, channelID, binding.ExternalID)
		defer r.Principal.UnbindMeeting(meetingID)
	}

	chProgress := &channelProgress{pool: r.Bots, channelID: channelID, loc: loc}
	chStream := &channelStream{pool: r.Bots, channelID: channelID, loc: loc}

	var eng *engine.Engine
	var err error
	if r.Principal != nil && cfg.Confirmation == meeting.ConfirmationModeRequired {
		eng, err = bootstrap.NewEngineWithPrincipal(r.Cfg, r.Principal)
	} else {
		eng, err = bootstrap.NewEngine(r.Cfg)
	}
	if err != nil {
		_ = r.Bots.Default.Send(ctx, channelID, meetEngineFailedText(loc, err))
		return
	}
	eng.Progress = engine.TeeProgressLogger{
		Loggers: []engine.ProgressLogger{engine.StdProgressLogger{}, chProgress},
	}
	eng.Stream = engine.TeeStreamLogger{
		Loggers: []engine.StreamLogger{engine.StdStreamLogger{}, chStream},
	}

	parts, err := parseParticipants(r.Discord.MeetParticipants)
	if err != nil {
		_ = r.Bots.Default.Send(ctx, channelID, meetConfigErrorText(loc, err))
		return
	}

	minPtr := cfg.MinRoundsBeforeSynthesis
	freePtr := cfg.FreeDialogueQuestions
	if _, err := eng.CreateMeeting(ctx, engine.CreateMeetingInput{
		MeetingID:                meetingID,
		Topic:                    cfg.Topic,
		MeetingMode:              cfg.Mode,
		ConfirmationMode:         cfg.Confirmation,
		MaxRoundsPerSegment:      cfg.MaxRounds,
		MinRoundsBeforeSynthesis: &minPtr,
		FreeDialogueMaxQuestions: &freePtr,
		Participants:             parts,
	}); err != nil {
		_ = r.Bots.Default.Send(ctx, channelID, meetCreateFailedText(loc, err))
		return
	}

	log.Printf("discord meet started id=%s topic=%q mode=%s rounds=%d principal=%s channel=%s",
		meetingID, cfg.Topic, cfg.Mode, cfg.MaxRounds, binding.PrincipalID, channelID)

	final, err := eng.Run(ctx, meetingID)
	if err != nil {
		_ = r.Bots.Default.Send(ctx, channelID, meetRunFailedText(loc, meetingID, err))
		return
	}

	summary := formatMeetDone(final, r.Cfg.Workspace.Root, meetingID, loc)
	SendLong(r.Bots.Default, ctx, channelID, summary)
}

func extractBusyMeetingID(err error) string {
	if err == nil {
		return ""
	}
	const prefix = "(meeting "
	msg := err.Error()
	idx := strings.Index(msg, prefix)
	if idx < 0 {
		return ""
	}
	rest := msg[idx+len(prefix):]
	end := strings.IndexByte(rest, ')')
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func formatMeetDone(s meeting.State, workspaceRoot, meetingID string, loc Locale) string {
	var b strings.Builder
	if loc == LocaleZH {
		fmt.Fprintf(&b, "🏁 **会议结束** `%s`\n", meetingID)
		fmt.Fprintf(&b, "- 📊 状态：%s\n", statusLabel(string(s.Status), loc))
		if s.Outcome != "" {
			fmt.Fprintf(&b, "- ✅ 结果：%s\n", outcomeLabel(s.Outcome, loc))
		}
		if s.IsDeliberation() && s.Consensus != nil {
			fmt.Fprintf(&b, "- 📝 合成方式：%s\n", resolvedByLabel(s.Consensus.ResolvedBy, loc))
		} else if s.Consensus != nil {
			fmt.Fprintf(&b, "- 🤝 共识：%s\n", resolvedByLabel(s.Consensus.ResolvedBy, loc))
		}
		if s.IsDeliberation() && len(s.SynthesisOpenQuestions) > 0 {
			fmt.Fprintf(&b, "- ❓ 开放问题：%d 条\n", len(s.SynthesisOpenQuestions))
			for i, q := range s.SynthesisOpenQuestions {
				fmt.Fprintf(&b, "  %d. %s\n", i+1, q)
			}
		}
		fmt.Fprintf(&b, "- 🔄 研讨轮次：%d\n", s.DebateRoundCount())
		fmt.Fprintf(&b, "- 📁 工作区：`%s/%s`", strings.TrimSuffix(workspaceRoot, "/"), meetingID)
		if s.TokenUsageTotals.CallCount > 0 {
			fmt.Fprintf(&b, "\n- 🔢 Token 用量：%d", s.TokenUsageTotals.TotalTokens)
		}
		return b.String()
	}

	fmt.Fprintf(&b, "🏁 **Meeting finished** `%s`\n", meetingID)
	fmt.Fprintf(&b, "- Status: %s\n", s.Status)
	if s.Outcome != "" {
		fmt.Fprintf(&b, "- Outcome: %s\n", s.Outcome)
	}
	if s.IsDeliberation() && s.Consensus != nil {
		fmt.Fprintf(&b, "- Synthesis: resolved_by=%s\n", s.Consensus.ResolvedBy)
	} else if s.Consensus != nil {
		fmt.Fprintf(&b, "- Consensus: resolved_by=%s\n", s.Consensus.ResolvedBy)
	}
	if s.IsDeliberation() && len(s.SynthesisOpenQuestions) > 0 {
		fmt.Fprintf(&b, "- Open questions: %d\n", len(s.SynthesisOpenQuestions))
		for i, q := range s.SynthesisOpenQuestions {
			fmt.Fprintf(&b, "  %d. %s\n", i+1, q)
		}
	}
	fmt.Fprintf(&b, "- Debate rounds: %d\n", s.DebateRoundCount())
	fmt.Fprintf(&b, "- Workspace: `%s/%s`", strings.TrimSuffix(workspaceRoot, "/"), meetingID)
	if s.TokenUsageTotals.CallCount > 0 {
		fmt.Fprintf(&b, "\n- Tokens: %d", s.TokenUsageTotals.TotalTokens)
	}
	return b.String()
}
