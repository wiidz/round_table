package engine

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/meeting"
)

func TestRenderMinutesWithRecap(t *testing.T) {
	s := meeting.State{
		Topic: "骑士",
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 1, Summary: "round 1"}},
		},
		SynthesisSummary: "## Draft\n\ncontent",
		Consensus:        &meeting.ConsensusState{ResolvedBy: "readiness"},
		MeetingMode:      meeting.MeetingModeDeliberation,
	}
	got := renderMinutesWithRecap(s, "## 会议回顾\n\n过程说明")
	if !strings.Contains(got, "## Executive Recap") || !strings.Contains(got, "过程说明") {
		t.Fatalf("missing recap section:\n%s", got)
	}
	if idx := strings.Index(got, "## Executive Recap"); idx < 0 || strings.Index(got, "## Synthesis") < idx {
		t.Fatalf("recap should precede synthesis:\n%s", got)
	}
}

func TestModeratorExecutiveRecap_disabled(t *testing.T) {
	e := &Engine{LLMModeratorExecutiveRecap: false, Model: roundSummaryModel{summary: "x"}}
	if got := e.moderatorExecutiveRecap(nil, meeting.State{}); got != "" {
		t.Fatalf("got=%q", got)
	}
}

func TestModeratorExecutiveRecap_llm(t *testing.T) {
	recap := "## 会议回顾\n\n### 目标与议程覆盖\n已覆盖定位与循环。\n\n### 过程脉络\n两轮研讨收敛。\n\n### 关键转折\n- 自由对话明确流派分工\n\n### 当前态势\n大体一致。\n\n### 进入合成前提示\n关注开放数值问题。"
	e := &Engine{
		Model:                      roundSummaryModel{summary: recap},
		ModelName:                  "test",
		LLMModeratorExecutiveRecap: true,
	}
	got := e.moderatorExecutiveRecap(context.Background(), meeting.State{Topic: "骑士", CurrentRound: 2})
	if !strings.Contains(got, "目标与议程覆盖") {
		t.Fatalf("got=%q", got)
	}
}
