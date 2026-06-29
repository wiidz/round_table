package engine

import (
	"fmt"
	"strings"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

type briefGoalFields struct {
	Goal         string
	InScope      string
	OutOfScope   string
	DoneCriteria string
}

// parseBriefGoalFields splits the combined engine goal string (see discord formatBriefForEngineGoal)
// back into structured brief sections for MEETING.md display.
func parseBriefGoalFields(combined string) briefGoalFields {
	combined = strings.TrimSpace(combined)
	if combined == "" {
		return briefGoalFields{}
	}
	var goalParts []string
	var out briefGoalFields
	for _, block := range strings.Split(combined, "\n\n") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		switch {
		case strings.HasPrefix(block, "完成标准：") || strings.HasPrefix(block, "完成标准:"):
			out.DoneCriteria = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(block, "完成标准："), "完成标准:"))
		case strings.HasPrefix(block, "讨论范围：") || strings.HasPrefix(block, "讨论范围:"):
			out.InScope = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(block, "讨论范围："), "讨论范围:"))
		case strings.HasPrefix(block, "不在范围：") || strings.HasPrefix(block, "不在范围:"):
			out.OutOfScope = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(block, "不在范围："), "不在范围:"))
		case strings.HasPrefix(strings.ToLower(block), "done when:"):
			out.DoneCriteria = strings.TrimSpace(block[len("done when:"):])
		case strings.HasPrefix(strings.ToLower(block), "in scope:"):
			out.InScope = strings.TrimSpace(block[len("in scope:"):])
		case strings.HasPrefix(strings.ToLower(block), "out of scope:"):
			out.OutOfScope = strings.TrimSpace(block[len("out of scope:"):])
		default:
			goalParts = append(goalParts, block)
		}
	}
	out.Goal = strings.TrimSpace(strings.Join(goalParts, "\n\n"))
	if out.Goal == "" && (out.InScope != "" || out.OutOfScope != "" || out.DoneCriteria != "") {
		out.Goal = combined
		out.InScope, out.OutOfScope, out.DoneCriteria = "", "", ""
	}
	return out
}

// RenderMeetingDoc renders MEETING.md from folded meeting state.
func RenderMeetingDoc(s meeting.State) string {
	return renderMeetingDoc(s)
}

func renderMeetingDoc(s meeting.State) string {
	brief := parseBriefGoalFields(s.Goal)
	goalText := brief.Goal
	if goalText == "" {
		goalText = strings.TrimSpace(s.Goal)
	}

	var b strings.Builder
	b.WriteString("# 会议简报 · Meeting Brief\n\n")
	b.WriteString("| 项目 | 内容 |\n|------|------|\n")
	b.WriteString(fmt.Sprintf("| 会议编号 | `%s` |\n", s.ID))
	if !s.StartedAt.IsZero() {
		local := s.StartedAt.Local()
		b.WriteString(fmt.Sprintf("| 会议时间 | %s (%s) |\n",
			local.Format("2006-01-02 15:04"),
			local.Format("MST"),
		))
	}
	b.WriteString(fmt.Sprintf("| 会议状态 | %s |\n", meetingStatusLabel(s.Status)))
	b.WriteString(fmt.Sprintf("| 会议模式 | %s |\n", meetingModeLabel(s.MeetingMode)))
	if !s.IsDeliberation() {
		b.WriteString(fmt.Sprintf("| 共识策略 | %s |\n", s.ConsensusStrategy))
	}
	b.WriteString(fmt.Sprintf("| 确认模式 | %s |\n", s.ConfirmationMode))
	b.WriteString(fmt.Sprintf("| 辩论轮次上限 | %d（不含 Pre-meeting Round 0） |\n", s.MaxRoundsPerSegment))
	if s.FreeDialogueMaxQuestions > 0 {
		b.WriteString(fmt.Sprintf("| Round 1 后自由对话 | 每人最多 %d 轮提问 |\n", s.FreeDialogueMaxQuestions))
	}

	b.WriteString("\n## 会议主题\n\n")
	b.WriteString(s.Topic)
	b.WriteString("\n\n## 会议目标\n\n")
	b.WriteString(goalText)
	b.WriteByte('\n')

	writeBriefAgendaSection(&b, s.Agenda)
	writeBriefScopeSection(&b, "讨论范围", brief.InScope)
	writeBriefScopeSection(&b, "不在范围", brief.OutOfScope)
	writeBriefScopeSection(&b, "完成标准", brief.DoneCriteria)

	b.WriteString("\n## 参会人员\n\n")
	if len(s.ParticipantOrder) == 0 {
		b.WriteString("_待邀请_\n")
	} else {
		b.WriteString("| 参会者 | 角色 | 专长 | 参会目标 |\n")
		b.WriteString("|--------|------|------|----------|\n")
		for _, id := range s.ParticipantOrder {
			p := s.Participants[id]
			goal := p.Goal
			if goal == "" {
				goal = "—"
			}
			exp := p.Expertise
			if exp == "" {
				exp = "—"
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", id, p.Role, exp, goal)
		}
	}

	b.WriteString("\n## 会议流程\n\n")
	b.WriteString("1. **Pre-meeting（Round 0）**：各参会者独立提交初始观点（互不可见）\n")
	if s.IsDeliberation() {
		b.WriteString("2. **研讨讨论（Round 1+）**：各角色贡献设计点与约束（不投票）\n")
	} else {
		b.WriteString("2. **辩论讨论（Round 1+）**：按固定顺序发言\n")
	}
	if s.FreeDialogueMaxQuestions > 0 {
		b.WriteString("3. **自由对话（Round 1 后）**：参会者互相提问与回答\n")
		b.WriteString("4. **Moderator 总结**：每轮后提炼要点\n")
		if s.IsDeliberation() {
			b.WriteString("5. **方案合成**：达轮次上限后输出 design-draft\n")
		} else {
			b.WriteString("5. **共识判定**：No-objection 或达到轮次上限由 Moderator 裁决\n")
		}
		if s.ConfirmationMode == meeting.ConfirmationModeRequired {
			if s.IsDeliberation() {
				b.WriteString("6. **Principal 确认**：审阅方案草案是否足够进入下一环节\n")
			} else {
				b.WriteString("6. **Principal 确认**：共识结果提交 Principal 审批\n")
			}
		}
	} else {
		b.WriteString("3. **Moderator 总结**：每轮后提炼要点\n")
		if s.IsDeliberation() {
			b.WriteString("4. **方案合成**：达轮次上限后输出 design-draft\n")
		} else {
			b.WriteString("4. **共识判定**：No-objection 或达到轮次上限由 Moderator 裁决\n")
		}
		if s.ConfirmationMode == meeting.ConfirmationModeRequired {
			if s.IsDeliberation() {
				b.WriteString("5. **Principal 确认**：审阅方案草案是否足够进入下一环节\n")
			} else {
				b.WriteString("5. **Principal 确认**：共识结果提交 Principal 审批\n")
			}
		}
	}

	b.WriteString("\n---\n\n")
	if s.TokenUsageTotals.CallCount > 0 {
		fmt.Fprintf(&b, "**Token 用量**：共 %d tokens（%d 次 LLM 调用），详见 `usage/summary.md`。\n\n",
			s.TokenUsageTotals.TotalTokens, s.TokenUsageTotals.CallCount)
	}
	b.WriteString("_本文档由 RoundTable Engine 维护，随会议进展更新。详细发言见 `MINUTES.md` 与 `rounds/`。_\n")
	return b.String()
}

func writeBriefAgendaSection(b *strings.Builder, agenda []event.AgendaItem) {
	if len(agenda) == 0 {
		return
	}
	b.WriteString("\n## 讨论议题\n\n")
	for i, item := range agenda {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		fmt.Fprintf(b, "%d. %s\n", i+1, title)
	}
}

func writeBriefScopeSection(b *strings.Builder, heading, body string) {
	body = strings.TrimSpace(body)
	if body == "" {
		return
	}
	fmt.Fprintf(b, "\n## %s\n\n%s\n", heading, body)
}

func meetingModeLabel(mode string) string {
	switch mode {
	case meeting.MeetingModeDeliberation:
		return "研讨型（deliberation）"
	case meeting.MeetingModeDecision, "":
		return "裁决型（decision）"
	default:
		return mode
	}
}

func meetingStatusLabel(st meeting.Status) string {
	switch st {
	case meeting.StatusPreparing:
		return "准备中"
	case meeting.StatusRunning:
		return "进行中"
	case meeting.StatusPaused:
		return "已暂停"
	case meeting.StatusConsensus:
		return "共识达成"
	case meeting.StatusConfirmation:
		return "Principal 确认中"
	case meeting.StatusCompleted:
		return "已结束"
	case meeting.StatusArchived:
		return "已归档"
	default:
		return string(st)
	}
}
