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
| [ADR-0005](./ADR-0005-round-termination.md) | Round 终止条件 | Accepted |
| [ADR-0006](./ADR-0006-knowledge-scope.md) | Knowledge 作用域与存储 | Accepted |
| [ADR-0007](./ADR-0007-moderator-scheduling.md) | Moderator 调度策略 | Accepted |
| [ADR-0008](./ADR-0008-project-structure.md) | Go 工程结构 | Accepted |
| [ADR-0009](./ADR-0009-meeting-workspace.md) | Meeting Workspace 文件产出 | Accepted |
| [ADR-0010](./ADR-0010-agent-profiles.md) | Agent Profile 身份层 | Accepted |
| [ADR-0011](./ADR-0011-meeting-mode.md) | Meeting Mode（decision / deliberation） | Accepted |
| [ADR-0012](./ADR-0012-transport-reception-agent.md) | Transport Reception Agent（自然语义接待层） | Accepted |
| [ADR-0013](./ADR-0013-web-round-table-live-ui.md) | Web 圆桌 Live 视图（围坐发言 + Drawer 历史） | Draft |

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
| — | Web 圆桌 Live v2（WS turn 与 Engine 对齐） | [ADR-0013](./ADR-0013-web-round-table-live-ui.md) §6 |
