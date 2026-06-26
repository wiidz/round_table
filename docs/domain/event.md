# Event

Event 是 Domain 中**状态变更的唯一表达方式**。系统内部不通过隐式 mutation 改变 Meeting State，而是追加 Event。

---

## 已定义

### 定义

每个 Event 记录 Meeting 生命周期中发生的一件不可变事实。Event 序列是 Meeting State 的 source of truth，也是 Minutes 生成与审计回放的基础。

### 原则

> 系统里面，永远只有 Event。

状态 = Event 的 fold（归约）。当前 Meeting State 由 Event 历史推导或物化缓存。

### v0.1 Event 清单

详见 [ADR-0003-event-model.md](../architecture/ADR-0003-event-model.md)。

| Event | 触发者 | 含义 |
|-------|--------|------|
| **MeetingCreated** | Principal / System | Meeting 创建 |
| **ParticipantInvited** | Moderator | Participant 加入 |
| **RoundStarted** | Moderator | 新 Round 开始，含 Order |
| **ParticipantResponded** | Participant | 发言全文 + Stance（agree/object/abstain） |
| **RoundCompleted** | Moderator | Round 结束，含 Summary |
| **ModeratorSummarized** | Moderator | 辩论轮结束后提炼摘要（非全文复制） |
| **FreeDialogueStarted** | Moderator | Round 1 后 Q&A 开始 |
| **FreeDialogueQuestionAsked** | Participant | 自由对话提问（含 optional `token_usage`） |
| **FreeDialogueAnswered** | Participant | 自由对话回答（含 optional `token_usage`） |
| **FreeDialogueCompleted** | Moderator | 自由对话结束 |
| **ConsensusReached** | Strategy / Moderator | 达成一致，含 dissent |
| **ConsensusVetoed** | Principal | Principal 否决共识 |
| **ConsensusForced** | Principal | Principal 强制共识 |
| **ConfirmationPrepared** | Moderator | 整理 Confirmation Brief 与 Item 列表 |
| **ConfirmationPresented** | Moderator | 呈现给 Principal 审阅 |
| **ConfirmationApproved** | Principal | Principal 批准，进入最终结论 |
| **ConfirmationRejected** | Principal | Principal 驳回，附 Feedback |
| **ConfirmationSkipped** | Principal | 运行中将 required 改为 skip |
| **ConfirmationForced** | Principal | 达循环上限后强制批准 |
| **MeetingPaused** | Principal / Moderator | 暂停 |
| **MeetingResumed** | Principal / Moderator | 恢复 |
| **MeetingFinished** | Moderator / Principal | Meeting 完成 |
| **ArtifactProduced** | Participant / Moderator | 产出物就绪 |
| **ActionItemCreated** | Moderator | 后续待办 |

### v0.2 预留

| Event | 说明 |
|-------|------|
| ParticipantUpdated | 角色或 Goal 变更 |
| OpinionUpdated | 若与 Responded 拆分 |

OpinionUpdated 不纳入 v0.1——Stance 合并进 `ParticipantResponded`。

### TokenUsage（payload 字段）

以下 Event 的 payload 可携带 `token_usage`（Model Adapter 返回的 prompt / completion / total tokens）：

| Event | 记录对象 |
|-------|----------|
| `ParticipantResponded` | 该轮发言的 LLM 调用 |
| `FreeDialogueQuestionAsked` | 提问方 LLM 调用 |
| `FreeDialogueAnswered` | 回答方 LLM 调用 |

Fold 后累计至 `Meeting.State.TokenUsageTotals`，并投影至 workspace `usage/`。Stub Participant 无 LLM 调用时不写入。

### 设计约束

- Event 不可变、只追加
- Event payload 使用 Domain 类型，不含 LLM / Transport 细节
- Minutes 由 Event 策展生成，不是 Event 的简单拼接

---

## 已决议（ADR-0003）

详见 [ADR-0003-event-model.md](../architecture/ADR-0003-event-model.md)。

| 编号 | 决议 |
|------|------|
| D-E01 | **Event Sourcing + 周期性 Snapshot**（每 50 Event 或 Paused/Completed 时） |
| D-E02 | `ParticipantResponded` 存**全文** + **Stance**，不单独存 Opinion delta |
| D-E03 | `EventEnvelope` 含 **Version** 字段，支持 schema 演进 |
| D-E04 | v0.1 支持审计回放与 State 重建；Fork Meeting 推迟 v0.2 |
| D-E05 | **Pause / Resume 纳入 v0.1** |
| D-E06 | Knowledge 使用**独立 Event 流**；Meeting 仅持 `KnowledgeRef` |

Apply 实现见 `apps/server/internal/domain/meeting/apply.go`。

Event Envelope 的 `Actor` 字段取值：`principal` | `moderator` | `participant` | `system`。

---

## 关联

- 父索引：[README.md](./README.md)
- Actor 命名：[principal.md](./principal.md)
- 权威定义：[CONSTITUTION.md](../CONSTITUTION.md) § Development Principles — Keep the domain pure
- Meeting 状态：[state_machine.md](../flow/state_achine.md)
- 各概念文档中的 Event 引用
