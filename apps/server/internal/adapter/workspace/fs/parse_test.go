package fs

import (
	"testing"

	"round_table/apps/server/internal/adapter/workspace"
)

func TestEnrichFromMeetingDoc(t *testing.T) {
	doc := `# 会议简报

| 项目 | 内容 |
|------|------|
| 会议状态 | 已结束 |
| 会议模式 | 裁决型（decision） |
| 辩论轮次上限 | 3（不含 Pre-meeting Round 0） |
| Round 1 后自由对话 | 每人最多 2 轮提问 |

## 会议主题

Topic A

## 参会人员

| 参会者 | 角色 | 专长 | 参会目标 |
|--------|------|------|----------|
| a | a | x | — |
| b | b | y | — |
`
	idx := workspaceMeetingIndex()
	EnrichFromMeetingDoc(&idx, doc)

	if idx.Topic != "Topic A" {
		t.Fatalf("topic=%q", idx.Topic)
	}
	if idx.ModeKind != "decision" {
		t.Fatalf("mode_kind=%q", idx.ModeKind)
	}
	if idx.MaxRounds != 3 {
		t.Fatalf("max_rounds=%d", idx.MaxRounds)
	}
	if !idx.FreeDialogue {
		t.Fatal("free_dialogue want true")
	}
	if idx.ParticipantCount != 2 {
		t.Fatalf("participant_count=%d", idx.ParticipantCount)
	}
}

func TestEnrichFromMeetingDocDeliberationNoFree(t *testing.T) {
	doc := `| 会议模式 | 研讨型（deliberation） |
| 辩论轮次上限 | 1（不含 Pre-meeting Round 0） |

## 会议主题

Delib
`
	idx := workspaceMeetingIndex()
	EnrichFromMeetingDoc(&idx, doc)

	if idx.ModeKind != "deliberation" {
		t.Fatalf("mode_kind=%q", idx.ModeKind)
	}
	if idx.MaxRounds != 1 {
		t.Fatalf("max_rounds=%d", idx.MaxRounds)
	}
	if idx.FreeDialogue {
		t.Fatal("free_dialogue want false")
	}
}

func workspaceMeetingIndex() workspace.MeetingIndex {
	return workspace.MeetingIndex{ID: "mtg-test"}
}
