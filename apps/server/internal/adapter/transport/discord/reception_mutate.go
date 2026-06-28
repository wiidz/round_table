package discord

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"round_table/apps/server/internal/adapter/transport"
	"round_table/apps/server/internal/platform/config"
)

func (r *Reception) execMutatingTool(ctx context.Context, msg transport.Inbound, d receptionDecision) (string, error) {
	if reply, need := r.checkMutatingPrincipal(msg); need {
		return reply, nil
	}
	loc := r.loc()
	switch d.Tool {
	case receptionToolCreateParticipant:
		if r.Participants == nil {
			return expertStorageRequiredText(loc), nil
		}
		item, err := r.prepareCreateItem(d)
		if err != nil {
			if errors.Is(err, errReceptionNeedDisplayName) {
				r.storeClarifySession(msg, msg.Content, receptionDecision{
					Tool:        receptionToolClarify,
					PendingTool: receptionToolCreateParticipant,
				})
			}
			return receptionMutateClarifyText(loc, err), nil
		}
		return r.beginConfirm(msg, receptionConfirmSession{
			action:      receptionActionCreateParticipant,
			participant: item,
		})

	case receptionToolUpdateParticipant:
		if r.Participants == nil {
			return expertStorageRequiredText(loc), nil
		}
		oldID, item, err := r.prepareUpdateItem(d)
		if err != nil {
			return receptionMutateClarifyText(loc, err), nil
		}
		return r.beginConfirm(msg, receptionConfirmSession{
			action:           receptionActionUpdateParticipant,
			oldParticipantID: oldID,
			participant:      item,
		})

	case receptionToolDeleteParticipant:
		if r.Participants == nil {
			return expertStorageRequiredText(loc), nil
		}
		item, err := r.prepareDeleteItem(d)
		if err != nil {
			return receptionMutateClarifyText(loc, err), nil
		}
		return r.beginConfirm(msg, receptionConfirmSession{
			action:           receptionActionDeleteParticipant,
			oldParticipantID: item.ID,
			participant:      item,
		})

	case receptionToolStartMeeting:
		if r.Meet == nil {
			return receptionNoMeetingText(loc), nil
		}
		cfg, err := r.prepareStartMeeting(d)
		if err != nil {
			return receptionMutateClarifyText(loc, err), nil
		}
		return r.beginConfirm(msg, receptionConfirmSession{
			action:     receptionActionStartMeeting,
			meetConfig: cfg,
		})

	case receptionToolUpdateParticipantProfile:
		return r.execProfileUpdate(ctx, msg, d)

	default:
		return "", nil
	}
}

func (r *Reception) prepareCreateItem(d receptionDecision) (config.ParticipantRosterItem, error) {
	display := strings.TrimSpace(d.DisplayName)
	if display == "" {
		return config.ParticipantRosterItem{}, errReceptionNeedDisplayName
	}
	if err := config.ValidateParticipantDisplayName(display); err != nil {
		return config.ParticipantRosterItem{}, err
	}
	roster := r.Participants.roster()
	id := strings.TrimSpace(d.ParticipantID)
	if id == "" {
		id = suggestParticipantID(display, roster)
	}
	if err := config.ValidateParticipantID(id); err != nil {
		return config.ParticipantRosterItem{}, err
	}
	if rosterIndexIn(roster, id) >= 0 {
		return config.ParticipantRosterItem{}, fmt.Errorf("代号 %q 已存在", id)
	}
	exp := strings.TrimSpace(d.Expertise)
	if exp != "" {
		if err := validateExpertise(exp); err != nil {
			return config.ParticipantRosterItem{}, err
		}
	}
	return config.ParticipantRosterItem{ID: id, DisplayName: display, Expertise: exp}, nil
}

func (r *Reception) prepareUpdateItem(d receptionDecision) (string, config.ParticipantRosterItem, error) {
	ref := participantRefFromDecision(d)
	if ref == "" {
		return "", config.ParticipantRosterItem{}, errReceptionNeedParticipantRef
	}
	item, err := r.Participants.findParticipant(ref)
	if err != nil {
		return "", config.ParticipantRosterItem{}, err
	}
	oldID := item.ID
	if v := strings.TrimSpace(d.DisplayName); v != "" {
		if err := config.ValidateParticipantDisplayName(v); err != nil {
			return "", config.ParticipantRosterItem{}, err
		}
		item.DisplayName = v
	}
	if v := strings.TrimSpace(d.Expertise); v != "" {
		if err := validateExpertise(v); err != nil {
			return "", config.ParticipantRosterItem{}, err
		}
		item.Expertise = v
	}
	if v := strings.TrimSpace(d.ParticipantID); v != "" && v != oldID {
		if err := config.ValidateParticipantID(v); err != nil {
			return "", config.ParticipantRosterItem{}, err
		}
		if rosterIndexIn(r.Participants.roster(), v) >= 0 {
			return "", config.ParticipantRosterItem{}, fmt.Errorf("代号 %q 已存在", v)
		}
		item.ID = v
	}
	return oldID, item, nil
}

func (r *Reception) prepareDeleteItem(d receptionDecision) (config.ParticipantRosterItem, error) {
	ref := participantRefFromDecision(d)
	if ref == "" {
		return config.ParticipantRosterItem{}, errReceptionNeedParticipantRef
	}
	return r.Participants.findParticipant(ref)
}

func (r *Reception) prepareStartMeeting(d receptionDecision) (meetLaunchConfig, error) {
	topic := strings.TrimSpace(d.Topic)
	if topic == "" {
		return meetLaunchConfig{}, errReceptionNeedTopic
	}
	cfg := r.Meet.defaultLaunchConfig(topic, "")
	query := strings.TrimSpace(d.ParticipantQuery)
	if query != "" {
		ids, err := resolveRosterPick(query, r.Meet.dc().MeetParticipants)
		if err != nil {
			return meetLaunchConfig{}, err
		}
		cfg.ParticipantIDs = ids
		cfg.ParticipantsSummary = summarizeParticipantIDs(r.Meet.dc().MeetParticipants, ids)
	}
	return cfg, nil
}

func participantRefFromDecision(d receptionDecision) string {
	if ref := strings.TrimSpace(d.ParticipantRef); ref != "" {
		return ref
	}
	return strings.TrimSpace(d.ParticipantID)
}

func isReceptionMutatingTool(tool receptionTool) bool {
	switch tool {
	case receptionToolCreateParticipant, receptionToolUpdateParticipant,
		receptionToolDeleteParticipant, receptionToolStartMeeting,
		receptionToolUpdateParticipantProfile:
		return true
	default:
		return false
	}
}

func formatReceptionConfirmPreview(loc Locale, sess receptionConfirmSession) string {
	switch sess.action {
	case receptionActionCreateParticipant:
		return formatExpertConfirmCreate(loc, sess.participant)
	case receptionActionUpdateParticipant:
		return formatExpertConfirmEdit(loc, sess.oldParticipantID, sess.participant)
	case receptionActionDeleteParticipant:
		return formatExpertConfirmDelete(loc, sess.participant)
	case receptionActionStartMeeting:
		return formatReceptionStartMeetingConfirm(loc, sess.meetConfig)
	case receptionActionUpdateParticipantProfile:
		return formatProfileUpdateConfirm(loc, sess.participant, sess.profileFile, sess.profileContent, sess.profileGenerated)
	default:
		return expertConfirmChoiceText(loc)
	}
}

func formatReceptionStartMeetingConfirm(loc Locale, cfg meetLaunchConfig) string {
	short := formatConfigShort(cfg, loc)
	if loc == LocaleZH {
		return fmt.Sprintf(`🎙️ **开会 · 确认**

- 📌 主题：%s
- 👥 %s
- ⚙️ %s

**1** — 确认启动 · **0** — 取消`, cfg.Topic, cfg.ParticipantsSummary, short)
	}
	return fmt.Sprintf(`🎙️ **Start meeting · confirm**

- 📌 Topic: %s
- 👥 %s
- ⚙️ %s

**1** — launch · **0** — cancel`, cfg.Topic, cfg.ParticipantsSummary, short)
}

func receptionMutateClarifyText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "🤔 " + err.Error()
	}
	return "🤔 " + err.Error()
}

func receptionConfirmPendingText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 上一条操作等待确认。回复 **1** 确认或 **0** 取消。"
	}
	return "⚠️ A pending action awaits confirmation. Reply **1** to confirm or **0** to cancel."
}

func receptionConfirmCancelledText(loc Locale) string {
	if loc == LocaleZH {
		return "↩️ 已取消。"
	}
	return "↩️ Cancelled."
}

func receptionConfirmNotOwnerText(loc Locale) string {
	if loc == LocaleZH {
		return "⚠️ 只有发起该操作的用户可以确认或取消。"
	}
	return "⚠️ Only the user who started this action can confirm or cancel."
}
