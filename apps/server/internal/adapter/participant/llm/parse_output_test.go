package llm

import (
	"strings"
	"testing"
)

func TestParseOutput_contentWithInnerQuotes(t *testing.T) {
	raw := `{"content":"明确边界：在核心认证路径上，我无法接受"绕过撤销检查"的降级策略。如果Redis完全不可用，应拒绝所有需要验证撤销状态的请求（返回503或强制重新登录），而不是让已吊销令牌通过。RPO方面，我要求最多1秒的数据丢失窗口——Sentinel + 同步写入（WAIT命令）可实现，RTO在30秒内。审计日志方面，第一批迭代中本地文件+异步队列足够作为检查点，但必须保证事件序列化后不丢失（如使用持久化消息队列），且日志格式包含标准合规字段（时间戳、用户ID、操作类型、结果）。这两个刚性边界请写入架构设计文档。"}`

	out, err := parseOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !containsSubstring(out.Content, "绕过撤销检查") {
		t.Fatalf("content = %q", out.Content)
	}
	if !containsSubstring(out.Content, "架构设计文档") {
		t.Fatalf("content truncated: %q", out.Content)
	}
}

func TestParseOutput_debateWithStance(t *testing.T) {
	raw := `{"content":"方案有"重大"风险","stance":"object","object_reason":"缺少审计"}`
	out, err := parseOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out.Stance != "object" {
		t.Fatalf("stance = %q", out.Stance)
	}
	if out.ObjectReason != "缺少审计" {
		t.Fatalf("object_reason = %q", out.ObjectReason)
	}
	if !containsSubstring(out.Content, "重大") {
		t.Fatalf("content = %q", out.Content)
	}
}

func TestParseOutput_validJSON(t *testing.T) {
	out, err := parseOutput(`{"content":"同意","stance":"agree","object_reason":""}`)
	if err != nil {
		t.Fatal(err)
	}
	if out.Content != "同意" || out.Stance != "agree" {
		t.Fatalf("got %+v", out)
	}
}

func TestParseOutput_trailingCornerQuoteNoBrace(t *testing.T) {
	raw := `{"content":"我更倾向于「几乎不可见，只有特定天赋或技能才能察觉」的设计。隐身状态的核心爽感在于「未知的威胁」，如果轻微但普遍可见的淡影永久存在，对面就能持续追踪位置，隐身就退化成了加速buff，偷袭的威慑力和惊喜感会大幅下降。不过，我完全理解PVP需要反制空间——建议将「反制」放在主动预判层面：例如敌方可以通过使用侦测技能（如猎人标记、法师透视光环）短暂暴露影舞者，或者通过观察环境异动（如草丛晃动、音效提示）来推测位置，而不是默认给所有人一个永久标记。这样既保留了刺客偷袭的高光时刻，也让高水平的对手有机会通过技能交换和局势阅读来对抗，竞技深度体现在「技能博弈」而非「常驻视野压制」。后续运营活动可以侧重「隐身刺杀主题」的限时模式（如暗影狩猎），验证这类设计的实际体验。若测试数据表明隐身过于强势，再考虑加入更轻微的视觉反馈作为后续平衡手段，而非首发就弱化特色。」`
	out, err := parseOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !containsSubstring(out.Content, "几乎不可见") {
		t.Fatalf("content = %q", out.Content)
	}
	if strings.HasSuffix(out.Content, "」") {
		t.Fatalf("content should not end with corner quote: %q", out.Content)
	}
}

func TestParseOutput_trailingCornerQuote(t *testing.T) {
	raw := `{"content":"好问题。针对分身置换的延迟补偿，我们采用确定性帧同步 + 客户端预测 + 服务器回滚的组合方案。不会因分身数量影响同步质量。」}`
	out, err := parseOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !containsSubstring(out.Content, "帧同步") {
		t.Fatalf("content = %q", out.Content)
	}
	if strings.HasSuffix(out.Content, "」") {
		t.Fatalf("content should not end with corner quote: %q", out.Content)
	}
}

func containsSubstring(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && stringIndex(s, sub) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
