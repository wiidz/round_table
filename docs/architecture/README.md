# Architecture Decision Records

本目录记录 RoundTable 的架构决策（ADR）。  
Domain 概念见 [docs/domain/](../domain/README.md)，权威原则见 [CONSTITUTION.md](../CONSTITUTION.md)。

---

## 索引

| ADR | 标题 | 状态 |
|-----|------|------|
| [ADR-0001](./ADR-0001-project-vision.md) | 项目愿景：为什么是 Meeting Engine | Draft |
| [ADR-0002](./ADR-0002-consensus-strategy.md) | Consensus 判定策略 | Accepted |
| [ADR-0003](./ADR-0003-event-model.md) | Event 模型与持久化 | Accepted |
| [ADR-0004](./ADR-0004-principal-confirmation.md) | Principal Confirmation 确认关 | Accepted |

---

## ADR 格式

每份 ADR 包含：

1. **背景** — 要解决什么问题
2. **决策** — 具体选择了什么
3. **拒绝的选项** — 考虑过但未采纳的方案及原因
4. **后果** — 正负面影响与待实现项

状态：`Draft` | `Accepted` | `Superseded` | `Deprecated`

---

## 待起草

| 编号 | 主题 | 来源 |
|------|------|------|
| ADR-0004 | Moderator 调度策略 | [moderator.md](../domain/moderator.md) D-Mod* |
| ADR-0005 | Round 终止条件 | [round.md](../domain/round.md) D-R* |
| ADR-0006 | Knowledge 作用域 | [participant.md](../domain/participant.md) D-P01 |
