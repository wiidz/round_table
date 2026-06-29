package discord

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"

	profFS "round_table/apps/server/internal/adapter/profile/fs"
	"round_table/apps/server/internal/adapter/transport"
	"round_table/apps/server/internal/platform/config"
)

const defaultExpertiseTag = "general"

// ParticipantAdmin manages expert roster CRUD from Discord.
type ParticipantAdmin struct {
	ConfigSvc *config.Service
	Profile   *profFS.Store
	Locale    func() Locale
	Prefix    string
	setups    participantSetupSessions
}

func (a *ParticipantAdmin) locale() Locale {
	if a.Locale != nil {
		return a.Locale()
	}
	return LocaleEN
}

func (a *ParticipantAdmin) prefix() string {
	p := strings.TrimSpace(a.Prefix)
	if p == "" {
		return "!rt "
	}
	if !strings.HasSuffix(p, " ") {
		p += " "
	}
	return p
}

func (a *ParticipantAdmin) rosterRaw() string {
	if a.ConfigSvc == nil {
		return ""
	}
	return a.ConfigSvc.Current().Transport.Discord.MeetParticipants
}

func (a *ParticipantAdmin) roster() []config.ParticipantRosterItem {
	if a.ConfigSvc == nil {
		return nil
	}
	return config.ParticipantRosterFromConfig(a.ConfigSvc.Current())
}

func rosterIndexIn(roster []config.ParticipantRosterItem, id string) int {
	for i, item := range roster {
		if item.ID == id {
			return i
		}
	}
	return -1
}

func (a *ParticipantAdmin) findParticipant(ref string) (config.ParticipantRosterItem, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return config.ParticipantRosterItem{}, errExpertRefRequired
	}
	ids, err := resolveRosterPick(ref, a.rosterRaw())
	if err != nil {
		return config.ParticipantRosterItem{}, err
	}
	if len(ids) != 1 {
		return config.ParticipantRosterItem{}, fmt.Errorf("%w: %q", errSetupInvalidParticipants, ref)
	}
	for _, item := range a.roster() {
		if item.ID == ids[0] {
			return item, nil
		}
	}
	return config.ParticipantRosterItem{}, fmt.Errorf("专家 %q 不存在", ids[0])
}

func (a *ParticipantAdmin) HandleCommand(msg transport.Inbound, args []string) (string, error) {
	loc := a.locale()
	if a.ConfigSvc == nil {
		return expertStorageRequiredText(loc), nil
	}
	if len(args) == 0 {
		return expertUsageText(loc, a.prefix()), nil
	}
	sub := strings.ToLower(strings.TrimSpace(args[0]))
	rest := args[1:]

	switch sub {
	case "list", "ls", "列表":
		return a.formatList(loc), nil
	case "show", "view", "get", "查看", "详情":
		if len(rest) == 0 {
			return expertRefRequiredText(loc), nil
		}
		item, err := a.findParticipant(strings.Join(rest, " "))
		if err != nil {
			return expertErrorText(loc, err), nil
		}
		return a.formatShow(loc, item), nil
	case "new", "create", "add", "新建", "创建":
		return a.beginCreate(msg)
	case "edit", "update", "set", "编辑", "修改":
		if len(rest) == 0 {
			return expertRefRequiredText(loc), nil
		}
		return a.beginEdit(msg, strings.Join(rest, " "))
	case "delete", "del", "remove", "rm", "删除":
		if len(rest) == 0 {
			return expertRefRequiredText(loc), nil
		}
		return a.beginDelete(msg, strings.Join(rest, " "))
	case "cancel", "取消":
		if reply, ok := a.CancelSetup(msg.ChannelID, msg.AuthorID); ok {
			return reply, nil
		}
		return expertNothingToCancelText(loc), nil
	default:
		return expertUsageText(loc, a.prefix()), nil
	}
}

func (a *ParticipantAdmin) beginCreate(msg transport.Inbound) (string, error) {
	loc := a.locale()
	if a.setups.pending(msg.ChannelID) {
		return expertSetupBusyText(loc), nil
	}
	a.setups.put(msg.ChannelID, participantSetupSession{
		channelID: msg.ChannelID,
		authorID:  msg.AuthorID,
		mode:      participantSetupCreate,
		step:      participantStepAskDisplayName,
	})
	return formatExpertAskDisplayName(loc), nil
}

func (a *ParticipantAdmin) beginEdit(msg transport.Inbound, ref string) (string, error) {
	loc := a.locale()
	if a.setups.pending(msg.ChannelID) {
		return expertSetupBusyText(loc), nil
	}
	item, err := a.findParticipant(ref)
	if err != nil {
		return expertErrorText(loc, err), nil
	}
	a.setups.put(msg.ChannelID, participantSetupSession{
		channelID: msg.ChannelID,
		authorID:  msg.AuthorID,
		mode:      participantSetupEdit,
		step:      participantStepAskEditFields,
		draft:     item,
		oldID:     item.ID,
	})
	return formatExpertAskEditFields(loc, item), nil
}

func (a *ParticipantAdmin) beginDelete(msg transport.Inbound, ref string) (string, error) {
	loc := a.locale()
	if a.setups.pending(msg.ChannelID) {
		return expertSetupBusyText(loc), nil
	}
	item, err := a.findParticipant(ref)
	if err != nil {
		return expertErrorText(loc, err), nil
	}
	a.setups.put(msg.ChannelID, participantSetupSession{
		channelID: msg.ChannelID,
		authorID:  msg.AuthorID,
		mode:      participantSetupDelete,
		step:      participantStepConfirmDelete,
		draft:     item,
		oldID:     item.ID,
	})
	return formatExpertConfirmDelete(loc, item), nil
}

func (a *ParticipantAdmin) CancelSetup(channelID, authorID string) (string, bool) {
	sess, ok := a.setups.get(channelID)
	if !ok {
		return "", false
	}
	if sess.authorID != authorID {
		return expertSetupNotOwnerText(a.locale()), true
	}
	a.setups.clear(channelID)
	return expertSetupCancelledText(a.locale()), true
}

func (a *ParticipantAdmin) HandleSetupReply(msg transport.Inbound) (string, error) {
	sess, ok := a.setups.get(msg.ChannelID)
	if !ok {
		return "", nil
	}
	loc := a.locale()
	if msg.AuthorID != sess.authorID {
		return expertSetupNotOwnerText(loc), nil
	}
	body := strings.TrimSpace(msg.Content)
	if isExpertCancelTrigger(body) {
		a.setups.clear(msg.ChannelID)
		return expertSetupCancelledText(loc), nil
	}

	switch sess.mode {
	case participantSetupCreate:
		return a.handleCreateStep(msg, sess, body)
	case participantSetupEdit:
		return a.handleEditStep(msg, sess, body)
	case participantSetupDelete:
		return a.handleDeleteStep(sess, body)
	default:
		a.setups.clear(msg.ChannelID)
		return "", nil
	}
}

func (a *ParticipantAdmin) handleCreateStep(msg transport.Inbound, sess participantSetupSession, body string) (string, error) {
	loc := a.locale()
	switch sess.step {
	case participantStepAskDisplayName:
		if body == "" {
			return expertDisplayNameRequiredText(loc), nil
		}
		if err := config.ValidateParticipantDisplayName(body); err != nil {
			return expertErrorText(loc, err), nil
		}
		sess.draft.DisplayName = body
		sess.step = participantStepAskID
		sess.draft.ID = suggestParticipantID(body, a.roster())
		a.setups.put(msg.ChannelID, sess)
		return formatExpertAskID(loc, sess.draft.ID), nil

	case participantStepAskID:
		idInput := body
		if isExpertSkipToken(idInput) {
			idInput = sess.draft.ID
		}
		if strings.EqualFold(idInput, "自动") || strings.EqualFold(idInput, "auto") {
			idInput = suggestParticipantID(sess.draft.DisplayName, a.roster())
		}
		idInput = strings.TrimSpace(idInput)
		if err := config.ValidateParticipantID(idInput); err != nil {
			return expertErrorText(loc, err), nil
		}
		if rosterIndexIn(a.roster(), idInput) >= 0 {
			return expertIDTakenText(loc, idInput), nil
		}
		sess.draft.ID = idInput
		sess.step = participantStepAskExpertise
		a.setups.put(msg.ChannelID, sess)
		return formatExpertAskExpertise(loc), nil

	case participantStepAskExpertise:
		exp := body
		if !isExpertSkipToken(exp) {
			exp = strings.TrimSpace(exp)
			if err := validateExpertise(exp); err != nil {
				return expertErrorText(loc, err), nil
			}
			sess.draft.Expertise = exp
		}
		sess.step = participantStepConfirmCreate
		a.setups.put(msg.ChannelID, sess)
		return formatExpertConfirmCreate(loc, sess.draft), nil

	case participantStepConfirmCreate:
		if !isExpertConfirmYes(body) {
			if isExpertConfirmNo(body) {
				a.setups.clear(msg.ChannelID)
				return expertSetupCancelledText(loc), nil
			}
			return expertConfirmChoiceText(loc), nil
		}
		a.setups.clear(msg.ChannelID)
		return a.executeCreate(context.Background(), sess.draft)
	default:
		a.setups.clear(msg.ChannelID)
		return "", nil
	}
}

func (a *ParticipantAdmin) handleEditStep(msg transport.Inbound, sess participantSetupSession, body string) (string, error) {
	loc := a.locale()
	switch sess.step {
	case participantStepAskEditFields:
		updates, err := parseExpertEditFields(body)
		if err != nil {
			return expertErrorText(loc, err), nil
		}
		draft := sess.draft
		if v, ok := updates["名称"]; ok {
			if err := config.ValidateParticipantDisplayName(v); err != nil {
				return expertErrorText(loc, err), nil
			}
			draft.DisplayName = v
		}
		if v, ok := updates["name"]; ok {
			if err := config.ValidateParticipantDisplayName(v); err != nil {
				return expertErrorText(loc, err), nil
			}
			draft.DisplayName = v
		}
		if v, ok := updates["专长"]; ok {
			if !isExpertSkipToken(v) {
				if err := validateExpertise(v); err != nil {
					return expertErrorText(loc, err), nil
				}
				draft.Expertise = v
			}
		}
		if v, ok := updates["expertise"]; ok {
			if !isExpertSkipToken(v) {
				if err := validateExpertise(v); err != nil {
					return expertErrorText(loc, err), nil
				}
				draft.Expertise = v
			}
		}
		if v, ok := updates["代号"]; ok {
			if !isExpertSkipToken(v) {
				if err := config.ValidateParticipantID(v); err != nil {
					return expertErrorText(loc, err), nil
				}
				if v != sess.oldID && rosterIndexIn(a.roster(), v) >= 0 {
					return expertIDTakenText(loc, v), nil
				}
				draft.ID = v
			}
		}
		if v, ok := updates["id"]; ok {
			if !isExpertSkipToken(v) {
				if err := config.ValidateParticipantID(v); err != nil {
					return expertErrorText(loc, err), nil
				}
				if v != sess.oldID && rosterIndexIn(a.roster(), v) >= 0 {
					return expertIDTakenText(loc, v), nil
				}
				draft.ID = v
			}
		}
		sess.draft = draft
		sess.step = participantStepConfirmEdit
		a.setups.put(msg.ChannelID, sess)
		return formatExpertConfirmEdit(loc, sess.oldID, sess.draft), nil

	case participantStepConfirmEdit:
		if !isExpertConfirmYes(body) {
			if isExpertConfirmNo(body) {
				a.setups.clear(msg.ChannelID)
				return expertSetupCancelledText(loc), nil
			}
			return expertConfirmChoiceText(loc), nil
		}
		a.setups.clear(msg.ChannelID)
		return a.executeUpdate(context.Background(), sess.oldID, sess.draft)
	default:
		a.setups.clear(msg.ChannelID)
		return "", nil
	}
}

func (a *ParticipantAdmin) handleDeleteStep(sess participantSetupSession, body string) (string, error) {
	loc := a.locale()
	if !isExpertConfirmYes(body) {
		if isExpertConfirmNo(body) {
			a.setups.clear(sess.channelID)
			return expertSetupCancelledText(loc), nil
		}
		return expertConfirmChoiceText(loc), nil
	}
	a.setups.clear(sess.channelID)
	return a.executeDelete(context.Background(), sess.oldID)
}

func (a *ParticipantAdmin) executeCreate(ctx context.Context, item config.ParticipantRosterItem) (string, error) {
	loc := a.locale()
	if err := a.ConfigSvc.CreateParticipant(ctx, item); err != nil {
		return expertErrorText(loc, err), nil
	}
	if a.Profile != nil {
		if err := a.Profile.EnsureParticipant(item.ID); err != nil {
			return expertProfileErrorText(loc, err), nil
		}
	}
	return formatExpertCreated(loc, item), nil
}

func (a *ParticipantAdmin) executeUpdate(ctx context.Context, oldID string, item config.ParticipantRosterItem) (string, error) {
	loc := a.locale()
	if err := a.ConfigSvc.UpdateParticipant(ctx, oldID, item); err != nil {
		return expertErrorText(loc, err), nil
	}
	if a.Profile != nil && item.ID != oldID {
		if err := a.Profile.RenameParticipant(oldID, item.ID); err != nil {
			return expertProfileErrorText(loc, err), nil
		}
	}
	return formatExpertUpdated(loc, oldID, item), nil
}

func (a *ParticipantAdmin) executeDelete(ctx context.Context, id string) (string, error) {
	loc := a.locale()
	item, _ := a.findParticipant(id)
	if err := a.ConfigSvc.DeleteParticipant(ctx, id); err != nil {
		return expertErrorText(loc, err), nil
	}
	if a.Profile != nil {
		if err := a.Profile.DeleteParticipant(id); err != nil {
			return expertProfileErrorText(loc, err), nil
		}
	}
	return formatExpertDeleted(loc, item), nil
}

func (a *ParticipantAdmin) formatList(loc Locale) string {
	roster := a.roster()
	if len(roster) == 0 {
		return expertListEmptyText(loc)
	}
	var b strings.Builder
	if loc == LocaleZH {
		b.WriteString(fmt.Sprintf("📋 **专家名录**（%d）\n\n", len(roster)))
	} else {
		b.WriteString(fmt.Sprintf("📋 **Expert roster** (%d)\n\n", len(roster)))
	}
	for i, item := range roster {
		exp := strings.TrimSpace(item.Expertise)
		if exp == "" {
			exp = defaultExpertiseTag
		}
		b.WriteString(fmt.Sprintf("%d · `%s` · %s · %s\n", i+1, item.ID, item.DisplayName, exp))
	}
	b.WriteString("\n")
	b.WriteString(expertListFooterText(loc, a.prefix()))
	return b.String()
}

// FormatList returns the expert roster for web chat and other callers outside discord handlers.
func (a *ParticipantAdmin) FormatList(loc Locale) string {
	return a.formatList(loc)
}

func (a *ParticipantAdmin) formatShow(loc Locale, item config.ParticipantRosterItem) string {
	exp := strings.TrimSpace(item.Expertise)
	if exp == "" {
		exp = defaultExpertiseTag
	}
	bot := expertBotBindingLabel(a.ConfigSvc, item.ID, loc)
	if loc == LocaleZH {
		return fmt.Sprintf(`👤 **专家** `+"`%s`"+`
- 名称：%s
- 专长：%s
- Discord Bot：%s

档案（SOUL/AGENTS）请在 Web 控制台编辑。`, item.ID, item.DisplayName, exp, bot)
	}
	return fmt.Sprintf(`👤 **Expert** `+"`%s`"+`
- Display: %s
- Expertise: %s
- Discord bot: %s

Edit SOUL/AGENTS in the Web console.`, item.ID, item.DisplayName, exp, bot)
}

func expertBotBindingLabel(svc *config.Service, participantID string, loc Locale) string {
	if svc == nil {
		if loc == LocaleZH {
			return "未知"
		}
		return "unknown"
	}
	appID := config.DiscordBotForParticipant(svc.ParticipantIMBindingsView(), participantID)
	if appID == "" {
		if loc == LocaleZH {
			return "未绑定"
		}
		return "not bound"
	}
	return "`" + appID + "`"
}

func suggestParticipantID(displayName string, roster []config.ParticipantRosterItem) string {
	base := slugParticipantID(displayName)
	if base == "" {
		base = hashParticipantID(displayName)
	}
	if err := config.ValidateParticipantID(base); err != nil {
		base = hashParticipantID(displayName)
	}
	candidate := base
	for i := 2; rosterIndexIn(roster, candidate) >= 0; i++ {
		candidate = fmt.Sprintf("%s_%d", base, i)
	}
	return candidate
}

func hashParticipantID(displayName string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return "expert"
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(displayName))
	return fmt.Sprintf("p_%x", h.Sum32())
}

func slugParticipantID(displayName string) string {
	displayName = strings.TrimSpace(strings.ToLower(displayName))
	var b strings.Builder
	prevUnderscore := false
	for _, r := range displayName {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevUnderscore = false
		case r == ' ', r == '-', r == '_':
			if b.Len() > 0 && !prevUnderscore {
				b.WriteByte('_')
				prevUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

func validateExpertise(exp string) error {
	exp = strings.TrimSpace(exp)
	if exp == "" {
		return nil
	}
	if len([]rune(exp)) > 32 {
		return fmt.Errorf("专长不能超过 32 个字符")
	}
	return nil
}

func parseExpertEditFields(body string) (map[string]string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, errExpertEditEmpty
	}
	repl := strings.NewReplacer("，", ",", "；", ";", "：", ":", "＝", "=")
	body = repl.Replace(body)
	out := make(map[string]string)
	for _, part := range strings.Split(body, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("%w: %q", errExpertEditInvalid, part)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			return nil, fmt.Errorf("%w: %q", errExpertEditInvalid, part)
		}
		out[key] = val
	}
	if len(out) == 0 {
		return nil, errExpertEditEmpty
	}
	return out, nil
}

func isExpertSkipToken(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == "-" || s == "跳过" || strings.EqualFold(s, "skip")
}

func isExpertConfirmYes(s string) bool {
	s = strings.TrimSpace(s)
	return s == "1" || strings.EqualFold(s, "yes") || s == "是" || s == "确认"
}

func isExpertConfirmNo(s string) bool {
	s = strings.TrimSpace(s)
	return s == "0" || strings.EqualFold(s, "no") || s == "否" || s == "取消"
}

func isExpertCancelTrigger(content string) bool {
	s := strings.TrimSpace(content)
	if s == "" {
		return false
	}
	lower := strings.ToLower(normalizeASCIIForms(s))
	return matchExact(lower, "取消专家", "expert cancel", "cancel expert")
}

func (a *ParticipantAdmin) InputPhase(channelID string) ChannelInputPhase {
	if a.setups.pending(channelID) {
		return InputPhaseExpertSetup
	}
	return InputPhaseIdle
}
