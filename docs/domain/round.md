# Round

Round 是 Meeting 内的**一轮有序讨论**。Meeting 由 Pre-meeting（Round 0）与多轮辩论（Round 1+）组成。

---

## 已定义

### 定义

每个 Round 是一次结构化的发言—回应—总结循环。Round 是时间上的切片，不是 Workflow 的 step node。

### Round 编号约定

| 轮次 | 名称 | 计入 `max_rounds` | 说明 |
|------|------|-------------------|------|
| **0** | Pre-meeting | 否 | 各 Participant 独立视角，互不可见；`stance: none` |
| **1+** | Debate | 是 | 按 Order 发言，需 `agree` / `object` / `abstain` |

### 结构（辩论轮 1+）

```
RoundStarted
  → Order（发言顺序）
  → Responses（Participant 回应，各一次）
  → RoundCompleted（本轮发言汇总）
  → [仅 Round 1] Free Dialogue（Q&A，可配置）
  → ModeratorSummarized（提炼摘要）
  → Consensus Check
  → Next Round 或结束 Meeting
```

Pre-meeting（Round 0）结束后直接进入 Round 1，无 Moderator 轮间摘要。

### Free Dialogue（Round 1 后）

- **时机**：Round 1 的 `RoundCompleted` 之后、Round 1 的 `ModeratorSummarized` 之前
- **次数**：`free_dialogue_max_questions` × 参会人数（每人按 Order 轮流提问，回答者为下一位）
- **关闭**：配置为 `0` 或参会者不足 2 人时跳过
- **事件**：`FreeDialogueStarted` → `FreeDialogueQuestionAsked` / `FreeDialogueAnswered` → `FreeDialogueCompleted`

### 每轮产出

| 产出 | 说明 |
|------|------|
| **Order** | Moderator 确定的本次发言顺序 |
| **Responses** | 各 Participant 在被邀请后提交的内容 + Stance |
| **Summary** | `RoundCompleted` 写入 Minutes 的本轮发言汇总 |
| **Moderator Summary** | 辩论轮 1+ 轮间提炼（非全文复制） |
| **State Update** | 更新后的 Meeting State |

### 设计约束

- Round 由 Moderator 开启和结束，对应 Event：`RoundStarted`、`RoundCompleted`
- 一轮内 Participant 按 Order 依次发言，不抢话
- Consensus Check 在每轮 **ModeratorSummarized**（及 Round 1 的 Free Dialogue）之后
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

### 实现补充（v0.2）

| 编号 | 决议 |
|------|------|
| D-R07 | Round 0 为 Pre-meeting，不计入 `max_rounds_per_segment` |
| D-R08 | Round 1 完成后固定一次 Free Dialogue，`free_dialogue_max_questions` 默认 1 |

---

## 关联

- 父索引：[README.md](./README.md)
- 权威定义：[CONSTITUTION.md](../CONSTITUTION.md) § Core Concepts — Round
- 宿主：[meeting.md](./meeting.md)
- 调度：[moderator.md](./moderator.md) — [ADR-0007](../architecture/ADR-0007-moderator-scheduling.md)
- 共识：[consensus.md](./consensus.md)
- 事件：[event.md](./event.md) — RoundStarted, RoundCompleted, FreeDialogue*
- Workspace：[workspace.md](./workspace.md) — `pre-meeting/`、`rounds/`、`free-dialogue/`
