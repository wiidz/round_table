package discord

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/adapter/transport"
	"round_table/apps/server/internal/platform/config"
)

const maxProfileContentRunes = 1 << 16 // 64Ki runes — aligned with reasonable Discord paste size

func (r *Reception) tryProfileUpdateFastPath(ctx context.Context, msg transport.Inbound, body string) (string, error) {
	d, ok := parseProfileUpdateIntent(body)
	if !ok {
		return "", nil
	}
	d.Tool = receptionToolUpdateParticipantProfile
	return r.execMutatingTool(ctx, msg, d)
}

func (r *Reception) execProfileUpdate(ctx context.Context, msg transport.Inbound, d receptionDecision) (string, error) {
	if r.Participants == nil {
		return expertStorageRequiredText(r.loc()), nil
	}
	item, file, brief, err := r.prepareProfileUpdate(d)
	if err != nil {
		if errors.Is(err, errReceptionNeedProfileContent) {
			r.putClarifySession(msg, receptionToolUpdateParticipantProfile, receptionDecision{
				ParticipantRef: item.ID,
				ProfileFile:    file,
			})
			return formatAskProfileContent(r.loc(), item, file), nil
		}
		return receptionMutateClarifyText(r.loc(), err), nil
	}
	content, generated, err := r.resolveProfileContent(ctx, item, file, brief)
	if err != nil {
		return receptionMutateClarifyText(r.loc(), err), nil
	}
	return r.beginConfirm(msg, receptionConfirmSession{
		action:           receptionActionUpdateParticipantProfile,
		participant:      item,
		profileFile:      file,
		profileContent:   content,
		profileGenerated: generated,
	})
}

func (r *Reception) prepareProfileUpdate(d receptionDecision) (config.ParticipantRosterItem, string, string, error) {
	ref := participantRefFromDecision(d)
	file := normalizeProfileFile(d.ProfileFile)
	content := strings.TrimSpace(d.ProfileContent)

	if ref == "" {
		return config.ParticipantRosterItem{}, "", "", errReceptionNeedParticipantRef
	}
	if file == "" {
		return config.ParticipantRosterItem{}, "", "", errReceptionNeedProfileFile
	}
	item, err := r.Participants.findParticipant(ref)
	if err != nil {
		return config.ParticipantRosterItem{}, "", "", err
	}
	if content == "" {
		return item, file, "", errReceptionNeedProfileContent
	}
	if n := utf8.RuneCountInString(content); n > maxProfileContentRunes {
		return config.ParticipantRosterItem{}, "", "", fmt.Errorf("档案正文过长（最多 %d 字）", maxProfileContentRunes)
	}
	return item, file, content, nil
}

func parseProfileUpdateIntent(body string) (receptionDecision, bool) {
	body = strings.TrimSpace(body)
	if body == "" {
		return receptionDecision{}, false
	}
	file := detectProfileFileKeyword(body)
	if file == "" {
		return receptionDecision{}, false
	}

	content := ""
	head := body
	if h, tail, ok := splitProfileContent(body); ok {
		content = tail
		head = h
	} else if !containsProfileActionWord(body) {
		return receptionDecision{}, false
	}

	ref := extractProfileParticipantRef(head, file)
	if ref == "" {
		return receptionDecision{}, false
	}
	return receptionDecision{
		ParticipantRef: ref,
		ProfileFile:    file,
		ProfileContent: content,
	}, true
}

func detectProfileFileKeyword(body string) string {
	lower := strings.ToLower(body)
	// Prefer longer matches first (agents before soul substring issues — none here)
	for _, pair := range []struct{ kw, file string }{
		{"agents", profile.FileAgents},
		{"tools", profile.FileTools},
		{"soul", profile.FileSoul},
		{"档案", profile.FileSoul},
	} {
		if strings.Contains(lower, pair.kw) {
			return pair.file
		}
	}
	return ""
}

func containsProfileActionWord(body string) bool {
	return containsAnySubstring(body,
		"添加", "写", "编写", "编辑", "更新", "设置", "配置", "修改", "填入", "上传", "生成",
		"add", "write", "edit", "update", "set", "configure", "generate", "draft",
	)
}

func splitProfileContent(body string) (head, content string, ok bool) {
	for _, sep := range []string{"\n---\n", "\n\n---\n\n", "：\n", ":\n"} {
		if idx := strings.Index(body, sep); idx >= 0 {
			return strings.TrimSpace(body[:idx]), strings.TrimSpace(body[idx+len(sep):]), true
		}
	}
	for _, sep := range []string{"内容：", "内容:", "正文：", "正文:", "content:", "content："} {
		if idx := strings.Index(body, sep); idx >= 0 {
			return strings.TrimSpace(body[:idx]), strings.TrimSpace(body[idx+len(sep):]), true
		}
	}
	// SOUL：# heading …
	lower := strings.ToLower(body)
	for _, tag := range []string{"soul", "agents", "tools"} {
		for _, punct := range []string{"：", ":"} {
			needle := tag + punct
			if idx := strings.Index(lower, needle); idx >= 0 {
				rest := strings.TrimSpace(body[idx+len(needle):])
				if rest != "" {
					return strings.TrimSpace(body[:idx]), rest, true
				}
			}
		}
	}
	return "", "", false
}

func extractProfileParticipantRef(body, file string) string {
	s := body
	lower := strings.ToLower(s)
	for _, kw := range []string{"soul", "agents", "tools", "档案"} {
		s = strings.ReplaceAll(s, kw, " ")
		s = strings.ReplaceAll(s, strings.ToUpper(kw), " ")
	}
	_ = lower
	for _, word := range []string{
		"添加", "写", "编写", "编辑", "更新", "设置", "配置", "修改", "填入", "上传",
		"给", "帮", "把", "将", "的",
	} {
		s = strings.ReplaceAll(s, word, " ")
	}
	s = strings.ReplaceAll(s, file, " ")
	s = strings.ReplaceAll(s, strings.TrimSuffix(file, ".md"), " ")
	return strings.TrimSpace(collapseSpaces(s))
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func normalizeProfileFile(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	lower := strings.ToLower(raw)
	switch lower {
	case "soul", "soul.md", "人格", "档案":
		return profile.FileSoul
	case "agents", "agents.md", "行为", "规则":
		return profile.FileAgents
	case "tools", "tools.md", "工具":
		return profile.FileTools
	}
	if strings.HasSuffix(lower, ".md") {
		switch lower {
		case "soul.md", "agents.md", "tools.md":
			return lower
		}
	}
	return ""
}

func matchesProfileUpdateIntent(s string) bool {
	_, ok := parseProfileUpdateIntent(s)
	return ok
}

func formatAskProfileContent(loc Locale, item config.ParticipantRosterItem, file string) string {
	label := profileFileLabel(loc, file)
	if loc == LocaleZH {
		return fmt.Sprintf(`📝 **编辑专家档案** · `+"`%s`"+` · %s

请用一两句话描述 **%s** 的方向（角色、语气、边界等），我会用 AI 生成 Markdown 草稿供你确认。
也可直接粘贴完整 Markdown；取消发 **0** 或 **取消专家**。`, item.ID, item.DisplayName, label)
	}
	return fmt.Sprintf(`📝 **Edit expert profile** · `+"`%s`"+` · %s

Describe the desired **%s** in a few sentences — AI will draft Markdown for you to confirm.
Or paste full Markdown. Reply **0** to cancel.`, item.ID, item.DisplayName, label)
}

func formatProfileUpdateConfirm(loc Locale, item config.ParticipantRosterItem, file, content string, generated bool) string {
	label := profileFileLabel(loc, file)
	preview := truncateProfilePreview(content, 480)
	source := ""
	if generated {
		if loc == LocaleZH {
			source = "（AI 根据你的描述生成）\n\n"
		} else {
			source = "(AI draft from your description)\n\n"
		}
	}
	if loc == LocaleZH {
		return fmt.Sprintf(`📝 **编辑档案 · 确认**

- 专家：`+"`%s`"+` · %s
- 文件：%s

%s预览：
%s

**1** — 确认写入 · **0** — 取消`, item.ID, item.DisplayName, label, source, preview)
	}
	return fmt.Sprintf(`📝 **Profile update · confirm**

- Expert: `+"`%s`"+` · %s
- File: %s

%sPreview:
%s

**1** — write · **0** — cancel`, item.ID, item.DisplayName, label, source, preview)
}

func profileFileLabel(loc Locale, file string) string {
	if loc == LocaleZH {
		switch file {
		case profile.FileSoul:
			return "SOUL.md · 人格"
		case profile.FileAgents:
			return "AGENTS.md · 行为规则"
		case profile.FileTools:
			return "TOOLS.md · 工具"
		}
	}
	return file
}

func truncateProfilePreview(content string, maxRunes int) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return "（空）"
	}
	runes := []rune(content)
	if len(runes) <= maxRunes {
		return "```\n" + content + "\n```"
	}
	return "```\n" + string(runes[:maxRunes]) + "\n…（已截断）\n```"
}

func (a *ParticipantAdmin) executeProfileUpdate(_ context.Context, id, filename, content string) (string, error) {
	loc := a.locale()
	if a.Profile == nil {
		return expertStorageRequiredText(loc), nil
	}
	if err := a.Profile.EnsureParticipant(id); err != nil {
		return expertProfileErrorText(loc, err), nil
	}
	if err := a.Profile.WriteParticipant(id, filename, []byte(content)); err != nil {
		return expertProfileErrorText(loc, err), nil
	}
	item, _ := a.findParticipant(id)
	return formatProfileUpdated(loc, item, filename), nil
}

func formatProfileUpdated(loc Locale, item config.ParticipantRosterItem, file string) string {
	label := profileFileLabel(loc, file)
	if loc == LocaleZH {
		return fmt.Sprintf("✅ 已更新专家 `%s` · %s 的 %s。\n\n可在 Web 控制台继续编辑。", item.ID, item.DisplayName, label)
	}
	return fmt.Sprintf("✅ Updated %s for expert `%s` · %s.", label, item.ID, item.DisplayName)
}

// putClarifySession stores a pending clarify without inferring tool from user text.
func (r *Reception) putClarifySession(msg transport.Inbound, tool receptionTool, partial receptionDecision) {
	partial.Tool = tool
	r.clarifies.put(msg.ChannelID, receptionClarifySession{
		channelID: msg.ChannelID,
		authorID:  msg.AuthorID,
		tool:      tool,
		partial:   partial,
	})
}
