package discord

import (
	"encoding/json"
	"fmt"
	"strings"
)

type receptionTool string

const (
	receptionToolNone               receptionTool = "none"
	receptionToolListParticipants   receptionTool = "list_participants"
	receptionToolMeetingStatus      receptionTool = "meeting_status"
	receptionToolGetArtifact        receptionTool = "get_artifact"
	receptionToolClarify            receptionTool = "clarify"
	receptionToolCreateParticipant  receptionTool = "create_participant"
	receptionToolUpdateParticipant  receptionTool = "update_participant"
	receptionToolDeleteParticipant  receptionTool = "delete_participant"
	receptionToolStartMeeting       receptionTool = "start_meeting"
	receptionToolUpdateParticipantProfile receptionTool = "update_participant_profile"
)

type receptionDecision struct {
	Tool             receptionTool `json:"tool"`
	PendingTool      receptionTool `json:"pending_tool"`
	Artifact         string        `json:"artifact"`
	Message          string        `json:"message"`
	DisplayName      string        `json:"display_name"`
	ParticipantID    string        `json:"participant_id"`
	ParticipantRef   string        `json:"participant_ref"`
	Expertise        string        `json:"expertise"`
	Topic            string        `json:"topic"`
	ParticipantQuery string        `json:"participant_query"`
	ProfileFile        string        `json:"profile_file"`
	ProfileContent     string        `json:"profile_content"`
}

const receptionJSONSchema = `Respond with JSON only:
{
  "tool": "list_participants|meeting_status|get_artifact|create_participant|update_participant|delete_participant|update_participant_profile|start_meeting|clarify|none",
  "pending_tool": "when tool is clarify: create_participant|update_participant|delete_participant|update_participant_profile|start_meeting",
  "artifact": "minutes|draft|open|conclusion",
  "message": "optional short reply when tool is clarify",
  "display_name": "for create/update participant",
  "participant_id": "new codename for create, or new id for update",
  "participant_ref": "existing expert id or display name for update/delete",
  "expertise": "optional expertise tag",
  "topic": "meeting topic for start_meeting",
  "participant_query": "optional roster pick for start_meeting, e.g. 策划、玩家",
  "profile_file": "soul|agents|tools for update_participant_profile",
  "profile_content": "markdown body when provided inline"
}
Rules:
- list_participants: user asks who experts are, roster, 有哪些专家
- meeting_status: user asks current state, progress, 会议状态, 开到哪了
- get_artifact: user wants summary, minutes, draft, open questions, conclusion, 纪要, 草案, 待决, 结论, 总结一下
- artifact: required for get_artifact; map 纪要/minutes/summary→minutes, 草案/draft→draft, 待决/open→open, 结论/conclusion→conclusion
- create_participant: 新增/创建/添加专家; fill display_name (required), optional participant_id, expertise
- update_participant: 修改/编辑专家; participant_ref required; optional display_name, expertise, participant_id (new id)
- delete_participant: 删除/移除专家; participant_ref required
- update_participant_profile: 给专家写/编辑 SOUL、AGENTS、TOOLS 档案; participant_ref + profile_file required; profile_content optional (ask follow-up if missing). NOT create_participant.
- start_meeting: user wants to start a meeting with topic (not already matched by 开个会 fast path); topic required; optional participant_query
- clarify: intent related to RoundTable but missing required fields; set pending_tool and ask one short question in message (same language as user)
- none: unrelated chit-chat; empty message
Mutating tools return a preview; Principal confirms with 1/0 before execution.`

func receptionSystemPrompt(loc Locale) string {
	if loc == LocaleZH {
		return `你是 RoundTable Discord 主持人 Bot 的「接待路由」模块。
根据 Principal 的自然语言选择 tool，输出 JSON。
不要编造会议内容。写操作会进入确认流程，你只负责解析意图与参数。给专家写 SOUL/AGENTS/TOOLS 用 update_participant_profile，不是 create_participant。` + "\n\n" + receptionJSONSchema
	}
	return `You are the RoundTable Discord host bot reception router.
Choose a tool and output JSON only. Do not invent meeting content. Mutating tools enter a 1/0 confirm flow.` + "\n\n" + receptionJSONSchema
}

func receptionUserPrompt(loc Locale, userText string, phase ChannelInputPhase, meetingID, rosterSummary string) string {
	phaseTitle := inputPhaseTitleEN(phase)
	if loc == LocaleZH {
		phaseTitle = inputPhaseTitleZH(phase)
	}
	var b strings.Builder
	b.WriteString("User message:\n")
	b.WriteString(userText)
	b.WriteString("\n\nContext:\n")
	b.WriteString("- input_phase: ")
	b.WriteString(string(phase))
	b.WriteString(" (")
	b.WriteString(phaseTitle)
	b.WriteString(")\n")
	if meetingID != "" {
		b.WriteString("- meeting_id: ")
		b.WriteString(meetingID)
		b.WriteByte('\n')
	}
	if rosterSummary != "" {
		b.WriteString("- roster: ")
		b.WriteString(rosterSummary)
		b.WriteByte('\n')
	}
	return b.String()
}

func parseReceptionDecision(raw string) (receptionDecision, error) {
	raw = strings.TrimSpace(raw)
	raw = stripJSONFence(raw)
	var d receptionDecision
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return receptionDecision{Tool: receptionToolNone}, fmt.Errorf("reception: parse json: %w", err)
	}
	d.Tool = receptionTool(strings.TrimSpace(string(d.Tool)))
	d.PendingTool = receptionTool(strings.TrimSpace(string(d.PendingTool)))
	d.Artifact = strings.TrimSpace(d.Artifact)
	d.Message = strings.TrimSpace(d.Message)
	d.DisplayName = strings.TrimSpace(d.DisplayName)
	d.ParticipantID = strings.TrimSpace(d.ParticipantID)
	d.ParticipantRef = strings.TrimSpace(d.ParticipantRef)
	d.Expertise = strings.TrimSpace(d.Expertise)
	d.Topic = strings.TrimSpace(d.Topic)
	d.ParticipantQuery = strings.TrimSpace(d.ParticipantQuery)
	d.ProfileFile = strings.TrimSpace(d.ProfileFile)
	d.ProfileContent = strings.TrimSpace(d.ProfileContent)
	if d.Tool == "" {
		d.Tool = receptionToolNone
	}
	return d, nil
}

func stripJSONFence(raw string) string {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "```") {
		return raw
	}
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(strings.TrimSpace(raw), "```")
	return strings.TrimSpace(raw)
}

func normalizeArtifactKind(kind string) string {
	k := strings.ToLower(strings.TrimSpace(kind))
	switch k {
	case "minutes", "minute", "summary", "纪要", "会议记录":
		return "minutes"
	case "draft", "design", "草案", "设计草案":
		return "draft"
	case "open", "open_questions", "questions", "待决", "开放问题", "遗留问题":
		return "open"
	case "conclusion", "decision", "结论", "已决":
		return "conclusion"
	default:
		return k
	}
}

func receptionErrorText(loc Locale, err error) string {
	if loc == LocaleZH {
		return "❌ 接待路由暂时不可用：" + err.Error()
	}
	return "❌ Reception router unavailable: " + err.Error()
}

func receptionFallbackClarifyText(loc Locale) string {
	if loc == LocaleZH {
		return "🤔 我没完全理解。你可以试试：\n- 有哪些专家\n- 会议状态\n- 获取纪要 / 草案\n- 新增专家「名称」\n- 测试专家 添加 SOUL\n- 或发送 **!rt help**"
	}
	return "🤔 I didn't quite get that. Try: list experts, meeting status, get minutes, add expert, or **!rt help**."
}

func receptionNoMeetingText(loc Locale) string {
	if loc == LocaleZH {
		return "ℹ️ 会议功能未就绪。"
	}
	return "ℹ️ Meeting features are not available."
}
