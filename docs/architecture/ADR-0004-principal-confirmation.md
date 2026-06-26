# ADR-0004: Principal Confirmation（确认关）

**状态**: Accepted  
**日期**: 2026-06-26  
**关联**: [confirmation.md](../domain/confirmation.md), [principal.md](../domain/principal.md), [ADR-0002-consensus-strategy.md](./ADR-0002-consensus-strategy.md), [ADR-0003-event-model.md](./ADR-0003-event-model.md)

**修订 ADR-0002 §6**：`completion_mode: manual` 由本 ADR 的 Confirmation 模型取代。

---

## 背景

Participant 达成 Consensus 只表示专家团队内部一致，不代表结论符合 Principal 预期。  
需要在输出最终结论前，增加一个 Moderator 整理、Principal 审阅的环节；同时允许通过配置跳过，完全交由 Meeting 自行得出结论。

---

## 决策

### 1. 新增 Confirmation 阶段

Meeting 生命周期在 Consensus 与 Completed 之间插入 **Confirmation** 状态：

```
Running → Consensus → Confirmation → Completed
                          ↓ Rejected
                       Running（追加讨论）
```

Consensus 回答「专家们是否同意」；Confirmation 回答「Principal 是否接受」。

### 2. 配置：confirmation_mode

| 模式 | 行为 | 适用场景 |
|------|------|----------|
| **`required`**（默认） | Consensus 后进入 Confirmation，等 Principal 审阅 | 需要 Principal 把关的决策 |
| **`skip`** | Consensus 后直接 `MeetingFinished` | 完全交由 Meeting 得出结论 |

创建 Meeting 时配置，运行中 Principal 可将 `required` 改为 `skip`（产生 `ConfirmationSkipped` Event），不可反向（skip 后不可再启用，避免绕过审计）。

### 3. Confirmation Brief 流程

Consensus 达成后，Moderator 执行：

```
1. 归纳讨论结论
2. 拆解为编号 Confirmation Item（1, 2, 3, 4…）
3. 生成 Confirmation Brief
4. 呈现给 Principal，Meeting 进入 Confirmation 状态
```

**Confirmation Item** 结构：

```go
type ConfirmationItem struct {
    Index       int    // 1-based
    Title       string
    Description string
    Source      string // 可选，如 "Round 2 summary"
}
```

**Confirmation Brief** 另含一段 **Executive Summary**（整体结论概述），便于 Principal 快速理解。

Moderator 职责见 [moderator.md](../domain/moderator.md) — 新增「确认关整理（Prepare Confirmation）」。

### 4. Principal 响应

Principal 在 Confirmation 状态提交响应：

```go
type ConfirmationResponse struct {
    Decision    ConfirmationDecision // approved | rejected
    Feedback    string             // rejected 时必填，说明不符合预期之处
    ItemNotes   map[int]string     // 可选，单项备注
}
```

| Decision | 行为 |
|----------|------|
| **`approved`** | 产生 `ConfirmationApproved` → `MeetingFinished` → 输出最终 Artifacts / Minutes |
| **`rejected`** | 产生 `ConfirmationRejected` → Meeting 回到 `Running` |

### 5. 驳回后的追加讨论

Principal 驳回时：

1. `ConfirmationRejected` Event 记录 Feedback 与 ItemNotes
2. Meeting Status → `Running`
3. Moderator 将 Principal Feedback 作为**新上下文**注入下一轮讨论
4. Participant 针对 Feedback 继续发言，直至再次 Consensus
5. 再次进入 Confirmation（`cycle` 递增）

每轮驳回后追加的 Round 数量受 `rounds_per_cycle` 限制（默认与 Meeting 的 `max_rounds` 相同，详见 ADR-0005）。

### 6. 循环上限

| 参数 | 默认 | 说明 |
|------|------|------|
| `max_confirmation_cycles` | 3 | 最多经历 3 次 Confirmation（含首次） |

达到上限且 Principal 再次 Rejected 时，Moderator 向 Principal 呈现三个选项：

| 选项 | 行为 |
|------|------|
| **Force Approve** | 强制批准当前 Brief，进入 Completed |
| **Continue** | 重置 cycle 计数（或 +1 额外 cycle），继续讨论 |
| **Abort** | `MeetingFinished`，标记 `outcome: aborted`，输出部分 Minutes |

Force Approve 产生 `ConfirmationForced` Event（区别于 Consensus 的 `ConsensusForced`）。

### 7. 与 ADR-0002 的关系

| ADR-0002 原设计 | ADR-0004 修订 |
|-----------------|---------------|
| `completion_mode: auto` | `confirmation_mode: skip` |
| `completion_mode: manual` | `confirmation_mode: required`（结构化 Brief，非简单等待） |
| `ConsensusVetoed` | 保留；Principal 在 Consensus 状态仍可 Veto 回到 Running |
| `ConsensusForced` | 保留；Principal 在 Consensus 状态可强制跳过 Participant 异议 |

Confirmation 是 Consensus **之后**的独立关卡，两者不互斥。

### 8. 新增 Event

| Event | 触发者 | Payload 要点 |
|-------|--------|--------------|
| `ConfirmationPrepared` | Moderator | cycle, brief, items[] |
| `ConfirmationPresented` | Moderator | cycle（呈现给 Principal） |
| `ConfirmationApproved` | Principal | cycle, item_notes |
| `ConfirmationRejected` | Principal | cycle, feedback, item_notes |
| `ConfirmationSkipped` | Principal | reason（运行中跳过） |
| `ConfirmationForced` | Principal | cycle, reason（达上限强制批准） |

`confirmation_mode: skip` 且创建时已配置时，不产生 `ConfirmationSkipped`——直接从 `ConsensusReached` 到 `MeetingFinished`（与 ADR-0002 auto 路径一致）。

---

## 完整流程图

```
                    ┌─────────────────────────────────┐
                    │         Running                  │
                    │   Round → Summary → Check      │
                    └──────────────┬──────────────────┘
                                   │
                          ConsensusReached
                                   │
                    ┌──────────────┴──────────────────┐
                    │     confirmation_mode?           │
                    └──────────────┬──────────────────┘
                          skip     │      required
                    ┌──────────────┼──────────────────┐
                    ▼                             ▼
            MeetingFinished              ConfirmationPrepared
                    │                             │
                    ▼                      ConfirmationPresented
               Completed                          │
                                    ┌─────────────┴─────────────┐
                                    ▼                           ▼
                        Principal Approved            Principal Rejected
                                    │                           │
                                    ▼                           ▼
                            MeetingFinished                  Running
                                    │                    (inject feedback)
                                    ▼                           │
                               Completed              Consensus → Confirmation
                                                              (cycle + 1)
```

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| Confirmation 在 Consensus 之前 | Principal 无法评估尚未形成的结论 |
| Participant 直接与 Principal 对话 | 违背 Moderator 中转通信模型 |
| 驳回后无限循环 | 需上限 + 兜底选项 |
| 用自由文本代替编号 Item | 不利于 Principal 逐项确认，难以追踪变更 |
| 默认 skip | Principal 明确要求此环节；默认 required 更安全，skip 显式配置 |

---

## 后果

### 正面

- 区分「专家共识」与「Principal 验收」，职责清晰
- 编号 Item 让 Principal 可精准反馈哪项不符合预期
- skip 模式保留全自动路径
- 驳回 → 追加讨论 → 再确认，形成闭环

### 负面

- 状态机与 Event 数量增加
- Moderator 需额外能力归纳 Confirmation Brief
- 多次循环可能延长 Meeting 时长——靠 `max_confirmation_cycles` 控制

### 待实现

- [ ] `Confirmation` 状态与 `confirmation_mode` 配置
- [ ] Moderator Prepare Confirmation 流程
- [ ] Confirmation Event 类型与 payload
- [ ] Principal Feedback 注入 Round 上下文
- [ ] 达上限时的 Principal 选项 UI / API

---

## 决议项对照

| 编号 | 决议 |
|------|------|
| D-CF01 | 新增 Confirmation 状态 |
| D-CF02 | `confirmation_mode: skip \| required`，默认 required |
| D-CF03 | Moderator 整理 Confirmation Brief + 编号 Item |
| D-CF04 | Approved → Completed；Rejected → Running |
| D-CF05 | 默认 max_confirmation_cycles = 3，达上限提供 Force / Continue / Abort |
| D-CF06 | 取代 ADR-0002 completion_mode |
