# Round

Round 是 Meeting 内的**一轮有序讨论**。Meeting 由多个 Round 组成。

---

## 已定义

### 定义

每个 Round 是一次结构化的发言—回应—总结循环。Round 是时间上的切片，不是 Workflow 的 step node。

### 结构

```
Round
  → Order（发言顺序）
  → Responses（Participant 回应）
  → Summary（Moderator 总结）
  → Consensus Check（是否达成一致）
  → Next Round 或结束 Meeting
```

### 每轮产出

| 产出 | 说明 |
|------|------|
| **Order** | Moderator 确定的本次发言顺序 |
| **Responses** | 各 Participant 在被邀请后提交的内容 |
| **Summary** | Moderator 对本轮的汇总 |
| **State Update** | 更新后的 Meeting State（含 Opinion 演化） |

### 设计约束

- Round 由 Moderator 开启和结束，对应 Event：`RoundStarted`、`RoundCompleted`
- 一轮内 Participant 按 Order 依次发言，不抢话
- Consensus Check 在每轮 Summary 之后；未达成则进入 Next Round
- Round 数量应有上限或终止条件（待决策）

---

## 待决策

| 编号 | 问题 | 选项 / 备注 |
|------|------|-------------|
| D-R01 | Round 与 Agenda Item 映射 | 一对一 / 一对多 / 多 Agenda 共享一轮 |
| D-R02 | 单 Round 最大发言轮次 | 每人一次 vs 允许多轮 rebuttal |
| D-R03 | Round 终止条件 | 固定 N 轮 / Consensus 达成 / Moderator 判定 / Principal 叫停 |
| D-R04 | Round 内是否允许 Principal 插话 | 影响 Order 与 Event 模型 |
| D-R05 | Summary 是否写入 Minutes 逐条 | 全文 vs 摘要 vs 结构化字段 |
| D-R06 | 最后一轮无 Consensus 如何处理 | 强制 Moderator Decision / 标记 Partial Consensus / 回到 Running |

---

## 关联

- 父索引：[README.md](./README.md)
- 权威定义：[CONSTITUTION.md](../CONSTITUTION.md) § Core Concepts — Round
- 宿主：[meeting.md](./meeting.md)
- 调度：[moderator.md](./moderator.md)
- 共识：[consensus.md](./consensus.md)
- 事件：[event.md](./event.md) — RoundStarted, RoundCompleted
