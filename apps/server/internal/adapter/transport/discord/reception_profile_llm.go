package discord

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/platform/config"
)

func (r *Reception) resolveProfileContent(ctx context.Context, item config.ParticipantRosterItem, file, userInput string) (string, bool, error) {
	userInput = strings.TrimSpace(userInput)
	if userInput == "" {
		return "", false, errReceptionNeedProfileContent
	}
	if looksLikeProfileMarkdown(userInput) {
		return userInput, false, nil
	}
	if r.Model == nil {
		return "", false, fmt.Errorf("根据描述生成档案需要配置 model API key（与会议 LLM 相同）")
	}
	draft, err := r.generateProfileDraft(ctx, item, file, userInput)
	if err != nil {
		return "", false, err
	}
	return draft, true, nil
}

func (r *Reception) generateProfileDraft(ctx context.Context, item config.ParticipantRosterItem, file, direction string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	existing := r.readParticipantProfileFile(item.ID, file)
	template := r.readProfileTemplate(file)

	raw, err := r.Model.Complete(ctx, model.Request{
		Model: r.modelName(),
		Messages: []model.Message{
			{Role: "system", Content: profileDraftSystemPrompt(r.loc(), file)},
			{Role: "user", Content: profileDraftUserPrompt(item, file, direction, template, existing)},
		},
		Temperature: 0.5,
	})
	if err != nil {
		return "", err
	}
	content := stripJSONFence(strings.TrimSpace(raw.Content))
	if content == "" {
		return "", fmt.Errorf("模型未返回有效档案正文")
	}
	if n := utf8.RuneCountInString(content); n > maxProfileContentRunes {
		return "", fmt.Errorf("生成内容过长（%d 字）", n)
	}
	return content, nil
}

func profileDraftSystemPrompt(loc Locale, file string) string {
	role := profileFileLabel(loc, file)
	if loc == LocaleZH {
		return fmt.Sprintf(`你是 RoundTable 专家档案撰写助手（ADR-0010）。
任务：根据 Principal 的方向描述，为 Participant 撰写 **%s** Markdown 档案。
要求：
- 只输出 Markdown 正文，不要代码围栏，不要 JSON
- 保留与模板一致的一级/二级标题结构（如 SOUL 的 ## 语气 / ## 边界 / ## 价值观）
- 内容具体、可执行，贴合专家角色；中文为主
- 不要编造系统能力或虚假事实；不确定处用「待确认」
- 不要替其他专家发言；聚焦本 Participant 的人格与边界`, role)
	}
	return fmt.Sprintf(`You draft RoundTable participant **%s** in Markdown (ADR-0010).
Output Markdown only — no fences, no JSON. Match template section headings. Be specific to the role.`, role)
}

func profileDraftUserPrompt(item config.ParticipantRosterItem, file, direction, template, existing string) string {
	exp := strings.TrimSpace(item.Expertise)
	if exp == "" {
		exp = "general"
	}
	var b strings.Builder
	b.WriteString("Participant:\n")
	fmt.Fprintf(&b, "- id: %s\n- display_name: %s\n- expertise: %s\n", item.ID, item.DisplayName, exp)
	b.WriteString("- file: ")
	b.WriteString(file)
	b.WriteString("\n\nPrincipal direction:\n")
	b.WriteString(direction)
	if template != "" {
		b.WriteString("\n\nTemplate reference (structure):\n```\n")
		b.WriteString(truncateProfilePreview(template, 1200))
		b.WriteString("\n```\n")
	}
	if existing != "" {
		b.WriteString("\n\nCurrent file (revise/extend, do not discard useful parts unless direction says so):\n```\n")
		b.WriteString(truncateProfilePreview(existing, 2000))
		b.WriteString("\n```\n")
	}
	return b.String()
}

func looksLikeProfileMarkdown(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "#") {
		return true
	}
	if strings.Contains(s, "\n## ") || strings.Contains(s, "\n# ") {
		return true
	}
	if utf8.RuneCountInString(s) > 400 {
		return true
	}
	return false
}

func (r *Reception) readParticipantProfileFile(participantID, file string) string {
	if r.Participants == nil || r.Participants.Profile == nil {
		return ""
	}
	data, err := r.Participants.Profile.ReadParticipant(participantID, file)
	if err != nil {
		return ""
	}
	return string(data)
}

func (r *Reception) readProfileTemplate(file string) string {
	if r.Participants == nil || r.Participants.Profile == nil {
		return ""
	}
	path := filepath.Join(r.Participants.Profile.Templates, "participants", file)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
