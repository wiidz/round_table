# Moderator

Moderator 是 Meeting 的**调度者**，不是领域专家。

> The Moderator owns orchestration. The Moderator does not own expertise.

---

## 已定义

### 定义

Moderator 控制整场 Meeting 的进行方式：谁发言、何时发言、如何汇总、何时结束。Moderator 不提供领域专业判断，只负责 orchestration。

### 职责

| 类别 | 能力 |
|------|------|
| **调度** | 决定发言顺序（Choose Next Speaker）、分配 Participant（Assign） |
| **上下文** | 向 Participant 分发当前 Round 所需上下文 |
| **总结** | 汇总每轮讨论（Summarize），更新 Meeting State |
| **共识** | 检测是否可进入 Consensus（具体策略见 [consensus.md](./consensus.md)） |
| **轮次** | 开启 / 结束 Round（RoundStarted / RoundCompleted） |
| **状态** | 管理 Meeting 生命周期：End / Pause / Resume |
| **确认关** | Consensus 后整理 Confirmation Brief，呈现给 Principal（见 [confirmation.md](./confirmation.md)） |

### 通信模型

所有消息必须经过 Moderator：

```
Principal → Moderator → Participant → Moderator → … → Consensus → Confirmation → Completed
```

Participant 之间**禁止**直接通信。

### 设计约束

- Moderator 不是 Main Agent、Boss Agent 或 TaskManager
- Moderator 不替代 Participant 做专业决策
- 调度逻辑属于 Domain；LLM 调用属于 Model Adapter（见 [ADR-0007](../architecture/ADR-0007-moderator-scheduling.md)）

---

## 已决议（ADR-0007）

详见 [ADR-0007-moderator-scheduling.md](../architecture/ADR-0007-moderator-scheduling.md)。

| 编号 | 决议 |
|------|------|
| D-Mod01 | v0.1 **纯规则引擎**；Summary/Brief 可选 LLM Adapter |
| D-Mod02 | **Fixed Registration Order**，每 Round 相同顺序 |
| D-Mod03 | 仅 **RoundCompleted** 时 Summary |
| D-Mod04 | Pause 完成当前回应后冻结，Resume 从断点继续 |
| D-Mod05 | 僵局 → Moderator Decision；不主动请求 Principal |
| D-Mod06 | Principal 可 Force/Pause/Abort；v0.1 不可 reorder |

---

## 关联

- 父索引：[README.md](./README.md)
- 权威定义：[CONSTITUTION.md](../CONSTITUTION.md) § Core Concepts — Moderator
- 委托人：[principal.md](./principal.md)
- 共识检测：[consensus.md](./consensus.md)
- 确认关：[confirmation.md](./confirmation.md)
- 轮次终止：[ADR-0005](../architecture/ADR-0005-round-termination.md)
