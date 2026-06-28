package discord

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
	"round_table/apps/server/internal/platform/config"
)

// Reception routes natural-language Principal messages via LLM + tools (ADR-0012).
type Reception struct {
	Model        model.Port
	ModelName    string
	Enabled      bool
	Registry     *principalbind.Registry
	Meet         *MeetRunner
	Participants *ParticipantAdmin
	Phase        func(channelID string) ChannelInputPhase
	Locale       func() Locale
	confirms     receptionConfirmSessions
	clarifies    receptionClarifySessions
}

func (r *Reception) active() bool {
	return r != nil && r.Enabled && r.Model != nil
}

func (r *Reception) enabled() bool {
	return r != nil && r.Enabled
}

func (r *Reception) loc() Locale {
	if r.Locale != nil {
		return r.Locale()
	}
	return LocaleEN
}

// TryHandle runs reception when no earlier handler matched.
func (r *Reception) TryHandle(ctx context.Context, msg transport.Inbound) (string, error) {
	if !r.enabled() {
		return "", nil
	}
	body := strings.TrimSpace(msg.Content)
	if body == "" || utf8.RuneCountInString(body) < 2 {
		return "", nil
	}
	if shouldSkipReception(body) {
		return "", nil
	}
	if r.confirms.pending(msg.ChannelID) || r.clarifies.pending(msg.ChannelID) {
		return "", nil
	}

	if reply, err := r.tryProfileUpdateFastPath(ctx, msg, body); err != nil || reply != "" {
		return reply, err
	}

	if r.Model == nil {
		return "", nil
	}

	phase := InputPhaseIdle
	if r.Phase != nil {
		phase = r.Phase(msg.ChannelID)
	}

	decision, err := r.route(ctx, body, phase, msg.ChannelID)
	if err != nil {
		return receptionErrorText(r.loc(), err), nil
	}
	if decision.Tool == receptionToolNone {
		return "", nil
	}
	if decision.Tool == receptionToolClarify {
		msgText := strings.TrimSpace(decision.Message)
		if msgText == "" {
			msgText = receptionFallbackClarifyText(r.loc())
		}
		r.storeClarifySession(msg, body, decision)
		return msgText, nil
	}
	if isReceptionMutatingTool(decision.Tool) {
		return r.execMutatingTool(ctx, msg, decision)
	}

	return r.execTool(ctx, msg, decision, phase)
}

func shouldSkipReception(body string) bool {
	if isInputStatusTrigger(body) || isMeetStartTrigger(body) || isMeetCancelTrigger(body) {
		return true
	}
	if isExpertCancelTrigger(body) {
		return true
	}
	if _, ok := isArtifactFetchTrigger(body); ok {
		return true
	}
	if isInterventionTrigger(body) || isFreeDialogueQuestionTrigger(body) {
		return true
	}
	_, ok := parseNaturalMeetStart(body)
	return ok
}

func (r *Reception) execTool(ctx context.Context, msg transport.Inbound, d receptionDecision, phase ChannelInputPhase) (string, error) {
	loc := r.loc()
	switch d.Tool {
	case receptionToolListParticipants:
		if r.Participants == nil {
			return expertStorageRequiredText(loc), nil
		}
		return r.Participants.formatList(loc), nil

	case receptionToolMeetingStatus:
		meetingID := ""
		if r.Meet != nil {
			meetingID = r.Meet.meetingIDForPhase(msg.ChannelID, phase)
		}
		return formatInputPhaseStatus(loc, phase, meetingID), nil

	case receptionToolGetArtifact:
		if r.Meet == nil {
			return receptionNoMeetingText(loc), nil
		}
		kind := normalizeArtifactKind(d.Artifact)
		if kind == "" {
			return artifactFetchUsageText(loc), nil
		}
		meetingID, ok := r.Meet.sessions.last(msg.ChannelID)
		if !ok {
			return artifactFetchNoMeetingText(loc), nil
		}
		return r.Meet.fetchArtifact(ctx, msg.ChannelID, meetingID, kind, loc)

	default:
		return "", nil
	}
}

func (r *Reception) route(ctx context.Context, userText string, phase ChannelInputPhase, channelID string) (receptionDecision, error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	meetingID := ""
	if r.Meet != nil {
		meetingID = r.Meet.meetingIDForPhase(channelID, phase)
	}

	raw, err := r.Model.Complete(ctx, model.Request{
		Model: r.modelName(),
		Messages: []model.Message{
			{Role: "system", Content: receptionSystemPrompt(r.loc())},
			{Role: "user", Content: receptionUserPrompt(r.loc(), userText, phase, meetingID, r.rosterSummaryForPrompt())},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return receptionDecision{}, err
	}
	return parseReceptionDecision(raw.Content)
}

func (r *Reception) modelName() string {
	if strings.TrimSpace(r.ModelName) != "" {
		return strings.TrimSpace(r.ModelName)
	}
	return "deepseek-chat"
}

func (r *Reception) rosterSummaryForPrompt() string {
	if r.Participants != nil {
		return formatRosterSummaryLine(r.Participants.roster())
	}
	if r.Meet != nil {
		var items []config.ParticipantRosterItem
		for _, line := range buildRosterLines(r.Meet.dc().MeetParticipants) {
			items = append(items, config.ParticipantRosterItem{ID: line.id, DisplayName: line.display})
		}
		return formatRosterSummaryLine(items)
	}
	return ""
}

func formatRosterSummaryLine(roster []config.ParticipantRosterItem) string {
	var parts []string
	for _, item := range roster {
		parts = append(parts, item.ID+"·"+item.DisplayName)
	}
	return strings.Join(parts, ", ")
}
