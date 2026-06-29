package engine

import (
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/meeting"
)

func TestModeratorSynthesizeFinal_executiveSummary(t *testing.T) {
	s := meeting.State{
		Topic:               "API 网关拆分方案",
		Goal:                "形成草案",
		CurrentRound:        2,
		MaxRoundsPerSegment: 2,
		ParticipantOrder:    []string{"architect", "sre"},
		RoundOrder:          []string{"architect", "sre"},
		Participants: map[string]meeting.ParticipantState{
			"architect": {ID: "architect", Role: "架构师"},
			"sre":       {ID: "sre", Role: "SRE"},
		},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {
				"architect": {Content: "最终倾向：采用独立网关集群，实例数下限 3。\n\n**待决问题：**\n- 是否启用跨区域主动-主动？\n- 能否接受 50ms 额外延迟？"},
			},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 2, Summary: "round 2"}},
		},
		ModeratorSummaries: map[int]string{
			2: "- 核心架构：独立网关\n- 实例下限 3",
		},
	}
	summary, open, _ := moderatorSynthesizeFinal(s)
	if !strings.Contains(summary, "## Executive Summary") {
		t.Fatal("missing executive summary")
	}
	if !strings.Contains(summary, "### 已决要点") {
		t.Fatal("missing decisions section")
	}
	if strings.Contains(summary, "## 详细记录") {
		t.Fatal("design-draft should not include detailed record appendix")
	}
	if !strings.Contains(summary, "MINUTES.md") {
		t.Fatal("missing pointer to MINUTES.md")
	}
	if len(open) == 0 {
		t.Fatal("expected open questions")
	}
	for _, q := range open {
		if strings.HasPrefix(q, "待决问题") && !strings.Contains(q, "是否") {
			t.Fatalf("noise question: %q", q)
		}
	}
	found := false
	for _, q := range open {
		if strings.Contains(q, "跨区域") || strings.Contains(q, "延迟") {
			found = true
		}
	}
	if !found {
		t.Fatalf("open questions = %v", open)
	}
}

func TestExtractOpenQuestionsFromText(t *testing.T) {
	text := `基于讨论：

**待决问题：**
- 是否启用跨区域主动-主动？
- 3. 待决问题列表：

最终倾向：独立网关集群。`
	got := extractOpenQuestionsFromText(text)
	if len(got) != 1 {
		t.Fatalf("got %d questions: %v", len(got), got)
	}
	if !strings.Contains(got[0], "跨区域") {
		t.Fatalf("got %q", got[0])
	}
}

func TestCollectDeliberationDecisions(t *testing.T) {
	s := meeting.State{
		CurrentRound: 2,
		RoundOrder:   []string{"a", "b"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {
				"a": {Content: "1. 明确为「独立网关 + 边车模式」：北向流量统一入口。\n最终倾向：实例数下限 3。"},
				"b": {Content: "2. **议题 B**：强烈支持「先单区域落地，再扩展多活」方案。"},
			},
		},
	}
	got, _ := collectDeliberationDecisions(s)
	if len(got) < 2 {
		t.Fatalf("expected >=2 decisions, got %v", got)
	}
	foundArch := false
	foundRollout := false
	for _, d := range got {
		if strings.Contains(d, "独立网关") || strings.Contains(d, "明确为") {
			foundArch = true
		}
		if strings.Contains(d, "单区域") || strings.Contains(d, "强烈支持") {
			foundRollout = true
		}
	}
	if !foundArch {
		t.Fatalf("missing architecture decision: %v", got)
	}
	if !foundRollout {
		t.Fatalf("missing rollout decision: %v", got)
	}
}

func TestExtractSchemePoints_skipsParticipantMeta(t *testing.T) {
	text := `- (架构师): 结合前几轮讨论，我建议进一步完善。
- 方案 A：独立网关集群，三实例起步，北向流量统一入口。
- 部署拓扑（初版）：`
	points := extractSchemePoints(text)
	if len(points) == 0 {
		t.Fatal("expected scheme points")
	}
	for _, p := range points {
		if strings.Contains(p, "结合前几轮") || strings.Contains(p, "架构师") {
			t.Fatalf("noise in scheme points: %q", p)
		}
	}
	if !strings.Contains(points[0], "独立网关") {
		t.Fatalf("points = %v", points)
	}
}

func TestOpenQuestions_filtersParagraphNoise(t *testing.T) {
	text := `**待决问题：**
- 是否启用跨区域主动-主动？
- 从安全角度，该方案在跨租户隔离上存在多个风险点需要逐一评估。第一，证书轮换窗口可能导致短暂不可用。第二，审计日志聚合延迟不确定。
- Q2: 灰度发布策略——是采用蓝绿还是金丝雀？`
	got := extractOpenQuestionsFromText(text)
	for _, q := range got {
		if strings.HasPrefix(q, "从安全角度") {
			t.Fatalf("paragraph noise leaked: %q", q)
		}
	}
}

func TestCollectDeliberationOpenQuestions_excludesResolved(t *testing.T) {
	s := meeting.State{
		CurrentRound: 2,
		RoundOrder:   []string{"b"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {
				"b": {Content: "强烈支持「先单区域落地，再扩展多活」方案。\n\nQ2: 多活策略——是否在本季度启动第二区域？"},
			},
		},
	}
	decisions, _ := collectDeliberationDecisions(s)
	open := collectDeliberationOpenQuestions(s, decisions, nil)
	for _, q := range open {
		if strings.Contains(q, "单区域落地") {
			t.Fatalf("resolved topic should be excluded: %q in %v", q, open)
		}
	}
}

func TestSummarizeCoreScheme_prefersRoundResponses(t *testing.T) {
	s := meeting.State{
		CurrentRound: 2,
		RoundOrder:   []string{"proposer", "reviewer"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {
				"proposer": {Content: "1. 方案 A：独立网关集群，三实例起步。\n2. 方案 B：复用现有 LB，渐进迁移。"},
			},
			2: {
				"reviewer": {Content: "1. 配置表热更新已有支持，无额外开发成本。\n2. 日志聚合需新增索引字段。"},
			},
		},
		ModeratorSummaries: map[int]string{
			2: "- 配置表热更新已有支持\n- 日志聚合需新增索引",
		},
	}
	got := summarizeCoreScheme(s)
	if !strings.Contains(got, "独立网关") {
		t.Fatalf("expected round-1 proposal, got:\n%s", got)
	}
	if strings.Contains(got, "配置表热更新") {
		t.Fatalf("should not prefer late-round feasibility bullets, got:\n%s", got)
	}
}

func TestSummarizeCoreScheme_prefersRevisionRound(t *testing.T) {
	s := meeting.State{
		CurrentRound: 2,
		RoundOrder:   []string{"proposer", "reviewer"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {
				"proposer": {Content: "1. 模块 A：独立网关集群。\n2. 模块 B：复用现有 LB。"},
			},
			2: {
				"proposer": {Content: "基于上一轮反馈，我收束核心框架如下：\n1. 模块 A：独立网关三实例起步。\n2. 模块 B：废弃复用 LB，改为边车模式。\n——这样北向流量统一入口。"},
			},
		},
	}
	got := summarizeCoreScheme(s)
	if !strings.Contains(got, "边车模式") {
		t.Fatalf("expected round-2 revision, got:\n%s", got)
	}
	if strings.Contains(got, "复用现有 LB") {
		t.Fatalf("round-1 proposal should be superseded, got:\n%s", got)
	}
}

func TestCollectDeliberationDecisions_revisionBlock(t *testing.T) {
	s := meeting.State{
		CurrentRound: 2,
		RoundOrder:   []string{"proposer", "reviewer"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {
				"proposer": {Content: "我收束核心框架如下：\n1. 上限 5 层资源槽，低层保底反馈。\n2. 副本节点不再采用独立寻路，仅做简单追踪。"},
				"reviewer": {Content: "② 控制效果必须与全局递减机制共用，避免连续控制链。"},
			},
		},
	}
	got, _ := collectDeliberationDecisions(s)
	if len(got) < 3 {
		t.Fatalf("expected >=3 decisions, got %v", got)
	}
}

func TestFinalizeDecisionLine_splitsTrailingQuestion(t *testing.T) {
	line := "影分身暴击继承方案已采纳打折方案（继承暴击率70%），但需确认是否允许触发装备特效？"
	decision, trailing := finalizeDecisionLine(line)
	if !strings.Contains(decision, "已采纳") {
		t.Fatalf("decision = %q", decision)
	}
	if trailing == "" || !strings.Contains(trailing, "装备特效") {
		t.Fatalf("trailing = %q", trailing)
	}
	if strings.Contains(decision, "？") {
		t.Fatalf("decision should not contain question: %q", decision)
	}
}

func TestCollectDeliberationDecisions_weakConsensus(t *testing.T) {
	s := meeting.State{
		CurrentRound: 2,
		RoundOrder:   []string{"a", "b", "c"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {
				"a": {Content: "1. 学习梯度最终方案：低等级纯特效，30级解锁轻量实体。"},
				"b": {Content: "1. 易伤层数暂定2层支持，留待测试调优。"},
				"c": {Content: "1. 灰度策略（蓝绿 vs 金丝雀）：支持。"},
			},
		},
	}
	got, _ := collectDeliberationDecisions(s)
	if len(got) < 3 {
		t.Fatalf("expected >=3 decisions, got %v", got)
	}
}

func TestCollectDeliberationOpenQuestions_tokenOverlapDedup(t *testing.T) {
	decisions := []string{"影分身暴击继承方案已采纳打折方案（继承暴击率70%、暴伤系数0.5）"}
	s := meeting.State{
		CurrentRound: 1,
		RoundOrder:   []string{"r"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {
				"r": {Content: "影分身是否继承玩家的暴击/背刺加成？如果继承，数值膨胀风险高。"},
			},
		},
	}
	open := collectDeliberationOpenQuestions(s, decisions, nil)
	for _, q := range open {
		if strings.Contains(q, "暴击") && strings.Contains(q, "继承") {
			t.Fatalf("should dedupe by topic overlap: %q in %v", q, open)
		}
	}
}

func TestCollectDeliberationOpenQuestions_spilloverFromDecision(t *testing.T) {
	s := meeting.State{
		CurrentRound: 1,
		RoundOrder:   []string{"a"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {
				"a": {Content: "方案已采纳打折系数 0.5，但需确认是否允许叠加其他 buff？"},
			},
		},
	}
	decisions, spillover := collectDeliberationDecisions(s)
	open := collectDeliberationOpenQuestions(s, decisions, spillover)
	foundSpillover := false
	for _, q := range open {
		if strings.Contains(q, "buff") || strings.Contains(q, "叠加") {
			foundSpillover = true
		}
	}
	if !foundSpillover {
		t.Fatalf("expected spillover question in open list: decisions=%v open=%v spillover=%v", decisions, open, spillover)
	}
}

func TestIsTentativeStatementNotQuestion(t *testing.T) {
	if !isTentativeStatementNotQuestion("我倾向实体部署，但需主程评估性能") {
		t.Fatal("expected tentative statement to be filtered")
	}
	if isTentativeStatementNotQuestion("是否采用实体部署？") {
		t.Fatal("real question should not be filtered")
	}
}

func TestDedupeOpenQuestions(t *testing.T) {
	items := []string{
		"各方案优先级排序需明确（开发成本、玩家反馈、回收周期）",
		"各方案优先级排序（开发成本、玩家反馈、回收周期）待明确",
		"动态事件奖励是否可能演变成新日常负担？需要明确奖励类型与频次上限。",
		"动态事件奖励定位是否可能演变成新日常负担，需在设计中约束。",
	}
	got := dedupeOpenQuestions(items)
	if len(got) != 2 {
		t.Fatalf("deduped=%v", got)
	}
}

func TestModeratorSynthesizeFinal(t *testing.T) {
	s := meeting.State{
		Topic:               "数据平台选型",
		Goal:                "形成方案草案",
		CurrentRound:        2,
		MaxRoundsPerSegment: 2,
		ParticipantOrder:    []string{"owner", "reviewer"},
		Participants: map[string]meeting.ParticipantState{
			"owner":    {ID: "owner", Role: "负责人"},
			"reviewer": {ID: "reviewer", Role: "评审"},
		},
		RoundOrder: []string{"owner", "reviewer"},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {
				"owner": {Content: "**待决问题：**\n- 是否接受托管服务的 vendor lock-in 风险？"},
			},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 2, Summary: "Round 2 summary"}},
		},
		ModeratorSummaries: map[int]string{2: "提炼 round 2"},
	}
	summary, open, _ := moderatorSynthesizeFinal(s)
	if summary == "" {
		t.Fatal("empty summary")
	}
	if len(open) == 0 {
		t.Fatal("expected open questions")
	}
}
