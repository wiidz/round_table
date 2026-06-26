# ADR-0007: Moderator 调度策略

**状态**: Accepted  
**日期**: 2026-06-26  
**关联**: [moderator.md](../domain/moderator.md), [ADR-0002-consensus-strategy.md](./ADR-0002-consensus-strategy.md), [ADR-0005-round-termination.md](./ADR-0005-round-termination.md)

---

## 背景

Moderator 负责 orchestration：发言顺序、总结、Consensus 检测、Confirmation 整理。  
Domain 文档（D-Mod01 ~ D-Mod06）尚未决议：Moderator 是否 LLM 驱动、如何 Choose Next Speaker、Summary 粒度等。

v0.1 目标是**可测试、可预测的 Meeting Engine**，调度逻辑必须先于 Model Adapter 实现。

---

## 决策

### 1. v0.1 Moderator = 纯规则引擎（D-Mod01）

| 能力 | v0.1 实现 |
|------|-----------|
| Choose Next Speaker | 规则 |
| Summarize | 模板拼接 + 可选 LLM Adapter（**非 Domain**） |
| Consensus Check | `ConsensusStrategy`（ADR-0002） |
| Prepare Confirmation | 模板 + 结构化提取（**非 Domain**） |

Domain 层 Moderator **不包含** LLM 调用。  
若 Summary / Brief 需要 LLM，通过 `ModeratorPort` 注入 Adapter，v0.1 默认 **stub 模板**。

LLM 辅助调度（动态 Order）推迟 v0.2。

### 2. 发言顺序：Fixed Registration Order（D-Mod02）

v0.1 默认策略：**FixedOrderStrategy**。

```
Order = ParticipantInvited 事件顺序（先邀请先发言）
每 Round 使用相同 Order
```

| 规则 | 说明 |
|------|------|
| 首轮 | 按 `participant_id` 注册顺序 |
| 后续 Round | 同一 Meeting 内 Order **不变** |
| 新 Participant | 仅 `Preparing` 或 Principal 显式允许时 `ParticipantInvited`；追加到 Order 末尾 |

**ChooseNextSpeaker** 伪代码：

```go
func (s FixedOrderStrategy) Next(order []ParticipantID, spoken map[ParticipantID]bool) ParticipantID {
    for _, id := range order {
        if !spoken[id] {
            return id
        }
    }
    return "" // round complete
}
```

v0.2 预留：`DynamicPriorityStrategy`（按 Agenda、反对意见优先等）。

### 3. 调度主循环（与 ADR-0005 对齐）

```
for each Running Segment:
  emit RoundStarted(round, order)
  for each participant in order:
    dispatch context → wait ParticipantResponded
  emit RoundCompleted(summary)
  if ConsensusStrategy.Evaluate → ConsensusReached; break
  if round >= max_rounds_per_segment → ModeratorDecision; break
  else → next round
```

Moderator 在 `ParticipantResponded` 后**不**自动 Summary，仅收集 Response；**Round 末**统一 Summary。

### 4. Summary 粒度（D-Mod03）

| 时机 | v0.1 |
|------|------|
| 每个 Participant 回应后 | ❌ 不 Summary |
| RoundCompleted 时 | ✅ 一次 Summary |

Summary 内容（stub 模板）：

- Topic / 当前 Round 编号
- 各 Participant 发言摘要（首 N 字或全文索引）
- 各 Stance 统计（agree / object / abstain）
- 可选：Principal Feedback（若本 Segment 由 ConfirmationRejected 触发）

LLM 润色 Summary 属于 Adapter，替换 stub 不改变 Event 结构。

### 5. Stance 收集（ADR-0002 衔接）

每 Round 所有 Participant 发言完成后，Moderator **再次按 Order** 请求 Stance（可合并进最后一次 respond 或单独 poll）：

v0.1 简化：**Stance 与 ParticipantResponded 同包提交**（发言时一并给出 stance）。

Moderator 在 Round 末汇总 Stance，交给 `ConsensusStrategy`。

### 6. Pause 行为（D-Mod04）

`MeetingPaused` 时：

1. 若 Participant 处于 `Speaking`，**允许完成当前回应**
2. 完成后全部 Participant → `Waiting`，不发起下一 Speaker
3. 不 emit `RoundStarted` 直到 `MeetingResumed`
4. Resume 后从**暂停点**继续（同一 Round 内未发言者继续 Order）

### 7. 僵局与 Principal 介入（D-Mod05、D-Mod06）

| 场景 | Moderator 行为 |
|------|----------------|
| Round 内无人 `object` 但未达 unanimous（策略相关） | 进入 Next Round |
| 达 `max_rounds_per_segment` | Moderator Decision → ConsensusReached |
| Principal `Force Consensus` | 立即 ConsensusReached，`resolved_by: principal` |
| Principal `Pause` / `Abort` | 按 Event 定义，不自行 invent |

v0.1 Moderator **不**主动请求 Principal 介入；僵局走 Moderator Decision。  
Principal 主动 Veto / Force / Abort 不受限。

v0.1 Principal **不能** override 发言顺序或跳过 Order 中某人；`Force Consensus` 是合法捷径。

### 8. Confirmation Brief 整理

Consensus 后、Confirmation 前，Moderator：

1. 从各 Round Summary + 最终 Consensus 归纳 **Executive Summary**
2. 拆解为编号 **ConfirmationItem**（1..N）
3. emit `ConfirmationPrepared` → `ConfirmationPresented`

v0.1 使用 **规则模板**（按 Agenda + 最后 Summary 分段），LLM 归纳 optional via Adapter。

### 9. ModeratorPort 接口（Domain 边界）

```go
type ModeratorScheduler interface {
    Order(ctx ScheduleContext) []ParticipantID
    NextSpeaker(order []ParticipantID, spoken map[ParticipantID]bool) (ParticipantID, bool)
}

type ModeratorSummarizer interface {
    SummarizeRound(ctx SummarizeContext) (string, error)
}

type ModeratorBriefPreparer interface {
    PrepareConfirmation(ctx ConfirmationContext) (ConfirmationBrief, error)
}
```

v0.1 提供 `FixedOrderScheduler` + `TemplateSummarizer` + `TemplateBriefPreparer` 默认实现。

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| v0.1 LLM Moderator | 不可测试、与 Domain 耦合 |
| 每轮随机 Order | 不可预测，难调试 |
| 动态 Order 无 ADR | 需 Agenda 进度模型，v0.2 再做 |
| 每回应后 Summary | Event 膨胀、Moderator 职责过重 |
| Pause 中断 mid-response | 状态复杂；完成当前回应足够 |

---

## 后果

### 正面

- 调度可单测（Fixed Order + 确定性循环）
- Domain 与 LLM 解耦
- 与 ADR-0005 Round 循环一一对应

### 负面

- 固定 Order 可能不适合所有议题
- Template Summary 质量有限——Adapter 可后续替换

### 待实现

- [ ] `FixedOrderScheduler`
- [ ] Moderator 主循环（engine 层）
- [ ] `TemplateSummarizer` / `TemplateBriefPreparer`
- [ ] Pause / Resume 与 Order 进度恢复

---

## 决议项对照

| 编号 | 决议 |
|------|------|
| D-Mod01 | v0.1 纯规则引擎；LLM 仅 via Adapter |
| D-Mod02 | Fixed Registration Order，每 Round 相同 |
| D-Mod03 | 仅 RoundCompleted 时 Summary |
| D-Mod04 | Pause 完成当前回应后冻结 Order |
| D-Mod05 | 僵局 → Moderator Decision；不主动呼叫 Principal |
| D-Mod06 | Principal 可 Force/Pause/Abort；v0.1 不可 reorder |

---

## 与 ADR-0005 关系

| ADR-0005 | ADR-0007 |
|----------|----------|
| 何时结束 Round / Segment | 每 Round 内如何遍历 Order |
| max_rounds_per_segment | Fixed Order 跑完 = 一轮 |
| Moderator Decision 触发 | Moderator 执行 Decision 并 emit Event |

两者共同定义 Moderator Scheduler 的 v0.1 行为。
