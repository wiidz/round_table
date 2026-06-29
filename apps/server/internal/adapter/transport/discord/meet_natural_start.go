package discord

import (
	"strings"

	"round_table/apps/server/internal/adapter/transport"
	"round_table/apps/server/internal/platform/config"
)

type naturalMeetStart struct {
	Topic              string
	ParticipantQuery   string
	HasParticipantHint bool
}

var naturalMeetPrefixes = []string{
	"开个会",
	"开一场会",
	"开会议",
	"开会",
}

var naturalMeetTopicMarkers = []string{
	"聊聊",
	"讨论",
	"关于",
	"主题是",
	"主题：",
	"主题:",
}

func parseNaturalMeetStart(content string) (naturalMeetStart, bool) {
	content = strings.TrimSpace(content)
	if content == "" {
		return naturalMeetStart{}, false
	}
	rest, ok := stripNaturalMeetPrefix(content)
	if !ok {
		return naturalMeetStart{}, false
	}
	if rest == "" {
		return naturalMeetStart{HasParticipantHint: false}, true
	}

	markerIdx := -1
	markerLen := 0
	for _, marker := range naturalMeetTopicMarkers {
		if idx := strings.Index(rest, marker); idx >= 0 && (markerIdx < 0 || idx < markerIdx) {
			markerIdx = idx
			markerLen = len(marker)
		}
	}

	out := naturalMeetStart{}
	if markerIdx >= 0 {
		out.Topic = cleanNaturalMeetTopic(rest[markerIdx+markerLen:])
		participants := cleanNaturalParticipantSegment(rest[:markerIdx])
		if participants != "" {
			out.ParticipantQuery = participants
			out.HasParticipantHint = true
		}
		return out, true
	}

	// No explicit topic marker: treat whole tail as topic if it looks like one.
	if looksLikeMeetTopicOnly(rest) {
		out.Topic = cleanNaturalMeetTopic(rest)
		return out, true
	}

	participants := cleanNaturalParticipantSegment(rest)
	if participants != "" {
		out.ParticipantQuery = participants
		out.HasParticipantHint = true
	}
	return out, true
}

func stripNaturalMeetPrefix(content string) (string, bool) {
	content = strings.TrimSpace(content)
	for _, prefix := range naturalMeetPrefixes {
		if strings.HasPrefix(content, prefix) {
			rest := strings.TrimSpace(strings.TrimPrefix(content, prefix))
			rest = strings.TrimLeft(rest, "，,；;：: ")
			return rest, true
		}
	}
	return "", false
}

func cleanNaturalParticipantSegment(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimLeft(raw, "，,；; ")
	raw = strings.TrimRight(raw, "，,；; ")
	for _, suffix := range []string{"一起", "一块儿", "共同", "一同", "参会", "参加"} {
		raw = strings.TrimSuffix(raw, suffix)
	}
	return strings.TrimSpace(raw)
}

func cleanNaturalMeetTopic(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimLeft(raw, "，,；;：: ")
	raw = strings.TrimRight(raw, "。！？!?")
	return strings.TrimSpace(raw)
}

func looksLikeMeetTopicOnly(rest string) bool {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return false
	}
	for _, hint := range []string{"一起", "、", ",", "，", "和", "与"} {
		if strings.Contains(rest, hint) {
			return false
		}
	}
	return len([]rune(rest)) >= 2
}

func (r *MeetRunner) TryBeginNaturalMeet(msg transport.Inbound) (string, error) {
	parsed, ok := parseNaturalMeetStart(strings.TrimSpace(msg.Content))
	if !ok {
		return "", nil
	}

	loc := r.locale()
	if reply, ok := r.checkBeginSetup(msg, loc); ok {
		return reply, nil
	}

	cfg := r.defaultLaunchConfig(parsed.Topic, "")
	var participantIDs []string
	if parsed.HasParticipantHint {
		ids, err := resolveRosterPick(parsed.ParticipantQuery, r.dc().MeetParticipants)
		if err != nil {
			return meetParticipantsPickErrorText(loc, err), nil
		}
		participantIDs = ids
		cfg.ParticipantIDs = participantIDs
		cfg.ParticipantsSummary = summarizeParticipantIDs(r.dc().MeetParticipants, participantIDs)
	}

	step := setupStepAskTopic
	var reply string

	switch {
	case cfg.Topic != "" && parsed.HasParticipantHint:
		step = setupStepBriefGoal
		reply = formatNaturalMeetBriefStartPrompt(loc, cfg)
	case cfg.Topic != "":
		step = setupStepPickParticipants
		reply = formatNaturalMeetTopicOnlyPrompt(loc, cfg.Topic, r.dc().MeetParticipants, r.meetCasts())
	case parsed.HasParticipantHint:
		step = setupStepAskTopic
		reply = formatNaturalMeetParticipantsOnlyPrompt(loc, cfg.ParticipantsSummary)
	default:
		step = setupStepAskTopic
		reply = formatAskTopicPrompt(loc)
	}

	r.setups.put(msg.ChannelID, meetSetupSession{
		channelID: msg.ChannelID,
		authorID:  msg.AuthorID,
		config:    cfg,
		step:      step,
	})
	return reply, nil
}

func formatNaturalMeetBriefStartPrompt(loc Locale, cfg meetLaunchConfig) string {
	if loc == LocaleZH {
		return fmtNaturalMeetHeadZH(cfg) + "\n\n" + formatAskBriefGoalPrompt(loc, cfg.Topic)
	}
	return fmtNaturalMeetHeadEN(cfg) + "\n\n" + formatAskBriefGoalPrompt(loc, cfg.Topic)
}

func formatNaturalMeetReadyPrompt(loc Locale, cfg meetLaunchConfig, prefix string, presets []meetPreset) string {
	head := formatModeratorSetupPrompt(loc, prefix, presets)
	if loc == LocaleZH {
		return fmtNaturalMeetHeadZH(cfg) + "\n\n" + head
	}
	return fmtNaturalMeetHeadEN(cfg) + "\n\n" + head
}

func fmtNaturalMeetHeadZH(cfg meetLaunchConfig) string {
	return "🎙️ **已理解开会请求**\n- 📌 主题：" + cfg.Topic + "\n- 👥 " + cfg.ParticipantsSummary
}

func fmtNaturalMeetHeadEN(cfg meetLaunchConfig) string {
	return "🎙️ **Meeting request understood**\n- 📌 Topic: " + cfg.Topic + "\n- 👥 " + cfg.ParticipantsSummary
}

func formatNaturalMeetTopicOnlyPrompt(loc Locale, topic, rosterRaw string, casts []config.MeetCastConfig) string {
	pick := formatPickParticipantsPrompt(loc, rosterRaw, casts)
	if loc == LocaleZH {
		return "🎙️ **主题：**" + topic + "\n\n" + pick
	}
	return "🎙️ **Topic:** " + topic + "\n\n" + pick
}

func formatNaturalMeetParticipantsOnlyPrompt(loc Locale, summary string) string {
	if loc == LocaleZH {
		return "🎙️ **参会：**" + summary + "\n\n" + formatAskTopicPrompt(loc)
	}
	return "🎙️ **Participants:** " + summary + "\n\n" + formatAskTopicPrompt(loc)
}
