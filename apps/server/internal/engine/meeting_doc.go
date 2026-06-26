package engine

import (
	"fmt"
	"strings"

	"round_table/apps/server/internal/domain/meeting"
)

func renderMeetingDoc(s meeting.State) string {
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
	b.WriteString(fmt.Sprintf("| 共识策略 | %s |\n", s.ConsensusStrategy))
	b.WriteString(fmt.Sprintf("| 确认模式 | %s |\n", s.ConfirmationMode))
	b.WriteString(fmt.Sprintf("| 辩论轮次上限 | %d（不含 Pre-meeting Round 0） |\n", s.MaxRoundsPerSegment))
	if s.FreeDialogueMaxQuestions > 0 {
		b.WriteString(fmt.Sprintf("| Round 1 后自由对话 | 每人最多 %d 轮提问 |\n", s.FreeDialogueMaxQuestions))
	}

	b.WriteString("\n## 会议主题\n\n")
	b.WriteString(s.Topic)
	b.WriteString("\n\n## 会议目标\n\n")
	b.WriteString(s.Goal)

	b.WriteString("\n\n## 参会人员\n\n")
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

	b.WriteString("\n## 议程\n\n")
	b.WriteString("1. **Pre-meeting（Round 0）**：各参会者独立提交初始观点（互不可见）\n")
	b.WriteString("2. **辩论讨论（Round 1+）**：按固定顺序发言\n")
	if s.FreeDialogueMaxQuestions > 0 {
		b.WriteString("3. **自由对话（Round 1 后）**：参会者互相提问与回答\n")
		b.WriteString("4. **Moderator 总结**：每轮辩论后提炼要点\n")
		b.WriteString("5. **共识判定**：No-objection 或达到轮次上限由 Moderator 裁决\n")
		if s.ConfirmationMode == meeting.ConfirmationModeRequired {
			b.WriteString("6. **Principal 确认**：共识结果提交 Principal 审批\n")
		}
	} else {
		b.WriteString("3. **Moderator 总结**：每轮辩论后提炼要点\n")
		b.WriteString("4. **共识判定**：No-objection 或达到轮次上限由 Moderator 裁决\n")
		if s.ConfirmationMode == meeting.ConfirmationModeRequired {
			b.WriteString("5. **Principal 确认**：共识结果提交 Principal 审批\n")
		}
	}

	if len(s.Agenda) > 0 {
		b.WriteString("\n## 附加议程项\n\n")
		for _, item := range s.Agenda {
			fmt.Fprintf(&b, "- %s", item.Title)
			if item.ID != "" {
				fmt.Fprintf(&b, " (`%s`)", item.ID)
			}
			b.WriteByte('\n')
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
