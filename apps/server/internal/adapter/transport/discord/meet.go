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
	Cfg      config.Config
	Discord  config.DiscordTransport
	Registry *principalbind.Registry
	Bots     *BotPool
	sessions meetSessions
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

// Start launches a meeting asynchronously and returns an immediate ack message.
func (r *MeetRunner) Start(msg transport.Inbound, parsed meetParseResult) (string, error) {
	loc := ParseLocale(r.Discord.Locale)
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	binding, ok := r.Registry.Get(scope)
	if !ok {
		return meetNeedBindText(loc), nil
	}
	if binding.ExternalID != msg.AuthorID {
		return meetNotScopePrincipalText(loc), nil
	}

	meetingID := fmt.Sprintf("mtg-dc-%d", time.Now().Unix())
	if err := r.sessions.tryStart(msg.ChannelID, meetingID); err != nil {
		if busy := extractBusyMeetingID(err); busy != "" {
			return meetChannelBusyText(loc, busy), nil
		}
		return err.Error(), nil
	}

	go r.runMeeting(msg.ChannelID, meetingID, binding, parsed)

	if loc == LocaleZH {
		return fmt.Sprintf("🚀 **会议已启动**\n- 🆔 `%s`\n- 📌 主题：%s\n- 🎯 模式：%s\n- 👤 Principal：%s\n\n进度将推送到本频道。",
			meetingID, parsed.Topic, meetingModeLabel(parsed.Mode, loc), binding.DisplayName), nil
	}
	return fmt.Sprintf("🚀 **Meeting started**\n- 🆔 `%s`\n- 📌 Topic: %s\n- 🎯 Mode: %s\n- 👤 Principal: %s\n\nProgress will post here.",
		meetingID, parsed.Topic, parsed.Mode, binding.DisplayName), nil
}

func (r *MeetRunner) runMeeting(channelID, meetingID string, binding principalbind.Binding, parsed meetParseResult) {
	defer r.sessions.clear(channelID)
	ctx := context.Background()

	loc := ParseLocale(r.Discord.Locale)
	chProgress := &channelProgress{pool: r.Bots, channelID: channelID, loc: loc}
	chStream := &channelStream{pool: r.Bots, channelID: channelID, loc: loc}
	eng, err := bootstrap.NewEngine(r.Cfg)
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

	rounds := r.Discord.MeetMaxRounds
	if rounds <= 0 {
		rounds = r.Cfg.Meeting.MaxRoundsPerSegment
	}
	minRounds := r.Cfg.Meeting.MinRoundsBeforeSynthesis
	if minRounds <= 0 {
		minRounds = 2
	}
	freeQ := r.Discord.MeetFreeDialogueQuestions
	confirmation := r.Discord.MeetConfirmation
	if confirmation == "" {
		confirmation = meeting.ConfirmationModeSkip
	}

	minPtr := minRounds
	freePtr := freeQ
	if _, err := eng.CreateMeeting(ctx, engine.CreateMeetingInput{
		MeetingID:                meetingID,
		Topic:                    parsed.Topic,
		MeetingMode:              parsed.Mode,
		ConfirmationMode:         confirmation,
		MaxRoundsPerSegment:      rounds,
		MinRoundsBeforeSynthesis: &minPtr,
		FreeDialogueMaxQuestions: &freePtr,
		Participants:             parts,
	}); err != nil {
		_ = r.Bots.Default.Send(ctx, channelID, meetCreateFailedText(loc, err))
		return
	}

	log.Printf("discord meet started id=%s topic=%q principal=%s channel=%s",
		meetingID, parsed.Topic, binding.PrincipalID, channelID)

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
