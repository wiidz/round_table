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
	Sender   ChannelSender
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
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	binding, ok := r.Registry.Get(scope)
	if !ok {
		return "请先绑定 Principal，再发起会议。", nil
	}
	if binding.ExternalID != msg.AuthorID {
		return "只有本范围的 Principal 可以发起会议。", nil
	}

	meetingID := fmt.Sprintf("mtg-dc-%d", time.Now().Unix())
	if err := r.sessions.tryStart(msg.ChannelID, meetingID); err != nil {
		return err.Error(), nil
	}

	go r.runMeeting(msg.ChannelID, meetingID, binding, parsed)

	return fmt.Sprintf("会议已启动\n- ID: `%s`\n- 主题: %s\n- 模式: %s\n- Principal: %s\n\n进度将推送到本频道。",
		meetingID, parsed.Topic, parsed.Mode, binding.DisplayName), nil
}

func (r *MeetRunner) runMeeting(channelID, meetingID string, binding principalbind.Binding, parsed meetParseResult) {
	defer r.sessions.clear(channelID)
	ctx := context.Background()

	chProgress := &channelProgress{sender: r.Sender, channelID: channelID}
	chStream := &channelStream{sender: r.Sender, channelID: channelID}
	eng, err := bootstrap.NewEngine(r.Cfg)
	if err != nil {
		_ = r.Sender.Send(ctx, channelID, "会议启动失败："+err.Error())
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
		_ = r.Sender.Send(ctx, channelID, "会议配置错误："+err.Error())
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
		_ = r.Sender.Send(ctx, channelID, "创建会议失败："+err.Error())
		return
	}

	log.Printf("discord meet started id=%s topic=%q principal=%s channel=%s",
		meetingID, parsed.Topic, binding.PrincipalID, channelID)

	final, err := eng.Run(ctx, meetingID)
	if err != nil {
		_ = r.Sender.Send(ctx, channelID, fmt.Sprintf("会议 `%s` 失败：%v", meetingID, err))
		return
	}

	summary := formatMeetDone(final, r.Cfg.Workspace.Root, meetingID)
	_ = r.Sender.Send(ctx, channelID, summary)
}

func formatMeetDone(s meeting.State, workspaceRoot, meetingID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**会议结束** `%s`\n", meetingID)
	fmt.Fprintf(&b, "- 状态: %s\n", s.Status)
	if s.Outcome != "" {
		fmt.Fprintf(&b, "- 结果: %s\n", s.Outcome)
	}
	if s.IsDeliberation() && s.Consensus != nil {
		fmt.Fprintf(&b, "- 合成: resolved_by=%s\n", s.Consensus.ResolvedBy)
	} else if s.Consensus != nil {
		fmt.Fprintf(&b, "- 共识: resolved_by=%s\n", s.Consensus.ResolvedBy)
	}
	if s.IsDeliberation() && len(s.SynthesisOpenQuestions) > 0 {
		fmt.Fprintf(&b, "- 开放问题: %d 条\n", len(s.SynthesisOpenQuestions))
		for i, q := range s.SynthesisOpenQuestions {
			if i >= 5 {
				fmt.Fprintf(&b, "- … 另有 %d 条\n", len(s.SynthesisOpenQuestions)-5)
				break
			}
			fmt.Fprintf(&b, "  %d. %s\n", i+1, q)
		}
	}
	fmt.Fprintf(&b, "- 辩论轮次: %d\n", s.DebateRoundCount())
	fmt.Fprintf(&b, "- Workspace: `%s/%s`", strings.TrimSuffix(workspaceRoot, "/"), meetingID)
	if s.TokenUsageTotals.CallCount > 0 {
		fmt.Fprintf(&b, "\n- Tokens: %d", s.TokenUsageTotals.TotalTokens)
	}
	return b.String()
}
