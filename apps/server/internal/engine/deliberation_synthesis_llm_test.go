package engine

import (
	"context"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/domain/meeting"
)

type synthesisFakeModel struct {
	content string
	err     error
}

func (f synthesisFakeModel) Complete(_ context.Context, _ model.Request) (model.Response, error) {
	if f.err != nil {
		return model.Response{}, f.err
	}
	return model.Response{
		Content: f.content,
		Usage: model.Usage{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		},
	}, nil
}

func TestParseSynthesisOutput(t *testing.T) {
	raw := `{"core_scheme":["方案 A"],"decisions":["采用三连击"],"open_questions":["冷却时间？"]}`
	out, err := parseSynthesisOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.CoreScheme) != 1 || len(out.Decisions) != 1 || len(out.OpenQuestions) != 1 {
		t.Fatalf("unexpected output: %+v", out)
	}

	wrapped := "```json\n" + raw + "\n```"
	out, err = parseSynthesisOutput(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Decisions) != 1 {
		t.Fatalf("wrapped parse failed: %+v", out)
	}
}

func TestSynthesizeDeliberationFinal_noModelUsesRules(t *testing.T) {
	e := &Engine{}
	s := meeting.State{
		Topic:        "测试主题",
		CurrentRound: 1,
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {"a": {Content: "最终倾向：采用方案 A。"}},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 1, Summary: "round 1"}},
		},
	}
	summary, open, usage, err := e.synthesizeDeliberationFinal(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if usage != nil {
		t.Fatal("expected no token usage without model")
	}
	if !strings.Contains(summary, "Executive Summary") {
		t.Fatalf("missing executive summary: %s", summary)
	}
	_ = open
}

func TestSynthesizeDeliberationFinal_llmPath(t *testing.T) {
	e := &Engine{
		Model: synthesisFakeModel{content: `{
			"core_scheme": ["核心：三连击 + 位移"],
			"decisions": ["统一冷却 8 秒"],
			"open_questions": ["PVP 平衡如何验证？"]
		}`},
		ModelName: "test-model",
	}
	s := meeting.State{
		Topic:        "职业设计",
		Goal:         "形成草案",
		CurrentRound: 2,
		ParticipantOrder: []string{"designer"},
		RoundOrder:       []string{"designer"},
		Participants: map[string]meeting.ParticipantState{
			"designer": {ID: "designer", Role: "策划"},
		},
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			2: {"designer": {Content: "收束：三连击 + 位移，冷却 8 秒。"}},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 2, Summary: "Round 2 summary"}},
		},
	}
	summary, open, usage, err := e.synthesizeDeliberationFinal(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if usage == nil || usage.TotalTokens != 150 {
		t.Fatalf("usage = %+v", usage)
	}
	if !strings.Contains(summary, "三连击") {
		t.Fatalf("missing core scheme: %s", summary)
	}
	if !strings.Contains(summary, "统一冷却") {
		t.Fatalf("missing decision: %s", summary)
	}
	if len(open) != 1 || !strings.Contains(open[0], "PVP") {
		t.Fatalf("open = %v", open)
	}
}

func TestSplitTentativeDecisions(t *testing.T) {
	decisions := []string{
		"付费点限定为纯外观",
		"暗能上限是否扩展至7点留待讨论，designer倾向否",
	}
	open := []string{"斩杀系数待确认"}
	firm, merged := splitTentativeDecisions(decisions, open)
	if len(firm) != 1 || !strings.Contains(firm[0], "纯外观") {
		t.Fatalf("firm = %v", firm)
	}
	if len(merged) != 2 {
		t.Fatalf("open = %v", merged)
	}
}

func TestDedupeDecisionsAgainstCoreScheme(t *testing.T) {
	core := []string{
		"职业定位：高机动、高伤害、低容错纯刺客，放弃控制/辅助路线，聚焦爆发体验。",
		"核心资源：暗影能量，上限100点，脱战每秒自然回复5点。",
		"隐身设计：半透明轮廓+攻击显形，攻击出手瞬间立即显形。",
		"残影机制：释放技能后原地留下持续1.5秒的残影，最多同时存在1个。",
	}
	decisions := []string{
		"职业定位：高机动、高伤害、低容错纯刺客，放弃控制/辅助路线。",
		"能量溢出处理：溢出化为增伤buff——5%暴击率持续3秒",
		"禁止残影穿越地形：残影生成时进行碰撞检测",
		"隐身方案：半透明轮廓+攻击显形",
	}
	got := dedupeDecisionsAgainstCoreScheme(core, decisions)
	if len(got) != 2 {
		t.Fatalf("deduped = %d items: %v", len(got), got)
	}
	for _, want := range []string{"增伤buff", "穿越地形"} {
		found := false
		for _, g := range got {
			if strings.Contains(g, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing %q in %v", want, got)
		}
	}
}

func TestDedupeDecisionsAgainstCoreScheme_incrementalFallback(t *testing.T) {
	core := []string{
		"简化操作模式（可选）：放宽影步判定范围30%、加长换位前摇至0.35秒，但降低换位伤害系数。",
		"影步对Boss完全禁用换位，仅打断读条。",
	}
	decisions := []string{
		"简化操作模式：默认关闭，玩家可随时切换，伤害降低约20%",
		"不开放玩家自定义前摇时长，内测后通过配置表热更新",
		"Boss禁用换位，仅打断读条",
	}
	got := dedupeDecisionsAgainstCoreScheme(core, decisions)
	if len(got) == 0 {
		t.Fatalf("expected incremental fallback, got none")
	}
	if len(got) >= len(decisions) {
		t.Fatalf("expected dedupe, got all: %v", got)
	}
	hasCustomPref := false
	hasBoss := false
	for _, g := range got {
		if strings.Contains(g, "前摇") {
			hasCustomPref = true
		}
		if strings.Contains(g, "Boss") {
			hasBoss = true
		}
	}
	if !hasCustomPref {
		t.Fatalf("expected incremental preheat decision in %v", got)
	}
	if hasBoss {
		t.Fatalf("Boss duplicate should be removed: %v", got)
	}
}

func TestDedupeDecisionsAgainstCoreScheme_emptyCore(t *testing.T) {
	decisions := []string{"采用方案 A"}
	got := dedupeDecisionsAgainstCoreScheme(nil, decisions)
	if len(got) != 1 {
		t.Fatalf("got %v", got)
	}
}

func TestSynthesizeDeliberationFinal_llmErrorFallsBack(t *testing.T) {
	e := &Engine{
		Model: synthesisFakeModel{err: context.DeadlineExceeded},
	}
	s := meeting.State{
		Topic:        "fallback",
		CurrentRound: 1,
		RoundResponses: map[int]map[string]meeting.RoundResponse{
			1: {"a": {Content: "最终倾向：方案 B。"}},
		},
		Minutes: meeting.MinutesDraft{
			Rounds: []meeting.RoundSummary{{RoundNumber: 1, Summary: "r1"}},
		},
	}
	summary, _, usage, err := e.synthesizeDeliberationFinal(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if usage != nil {
		t.Fatal("expected no usage on error fallback")
	}
	if !strings.Contains(summary, "Executive Summary") {
		t.Fatalf("expected rule fallback summary: %s", summary)
	}
}
