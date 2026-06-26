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
- Round 数量受 `max_rounds_per_segment` 约束（见 [ADR-0005](../architecture/ADR-0005-round-termination.md)）

---

## 已决议（ADR-0005）

详见 [ADR-0005-round-termination.md](../architecture/ADR-0005-round-termination.md)。

| 编号 | 决议 |
|------|------|
| D-R01 | v0.1 Round 不映射 Agenda Item |
| D-R02 | 每 Participant 每 Round 发言一次 |
| D-R03 | 优先 Consensus；达上限 `max_rounds_per_segment`（默认 5）→ Moderator Decision |
| D-R04 | v0.1 Principal 仅 Turn boundary 操作 |
| D-R05 | Round Summary 结构化写入 Minutes |
| D-R06 | 达上限用 Moderator Decision |

---

## 关联

- 父索引：[README.md](./README.md)
- 权威定义：[CONSTITUTION.md](../CONSTITUTION.md) § Core Concepts — Round
- 宿主：[meeting.md](./meeting.md)
- 调度：[moderator.md](./moderator.md) — [ADR-0007](../architecture/ADR-0007-moderator-scheduling.md)
- 共识：[consensus.md](./consensus.md)
- 事件：[event.md](./event.md) — RoundStarted, RoundCompleted
