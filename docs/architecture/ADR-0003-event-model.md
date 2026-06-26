# ADR-0003: Event 模型与持久化

**状态**: Accepted  
**日期**: 2026-06-26  
**关联**: [event.md](../domain/event.md), [ADR-0002-consensus-strategy.md](./ADR-0002-consensus-strategy.md), [ADR-0004-principal-confirmation.md](./ADR-0004-principal-confirmation.md), [CONSTITUTION.md](../CONSTITUTION.md)

---

## 背景

RoundTable 要求 Meeting State 由 Event 驱动，而非隐式 mutation。  
Domain 文档（D-E01 ~ D-E06）要求明确：持久化方案、Event payload、版本演进、回放范围、Pause/Resume、跨 Meeting 知识。

---

## 决策

### 1. 持久化：Event Sourcing + 周期性 Snapshot

采用 **Event Sourcing** 作为 source of truth，配合 **Snapshot** 优化读取性能。

```
Meeting State = fold(events[0..n])
Snapshot[k]   = fold(events[0..k])   // 物化缓存
```

| 规则 | 说明 |
|------|------|
| Event 只追加 | 不可修改、不可删除 |
| 写入顺序 | 单 Meeting 内严格有序（monotonic sequence） |
| Snapshot 触发 | 每 N 个 Event（默认 N=50）或 Meeting 进入 Paused/Completed 时 |
| 读取路径 | 优先 Snapshot + tail events；无 Snapshot 则全量 fold |

v0.1 存储：单进程内存 + 文件 append log。后续可换 PostgreSQL / SQLite，Domain 接口不变。

### 2. Event Envelope

所有 Event 共享统一信封：

```go
type EventEnvelope struct {
    ID          string    // UUID
    MeetingID   string
    Sequence    int       // Meeting 内单调递增，从 1 开始
    Type        EventType
    Version     int       // payload schema 版本，当前 = 1
    Payload     []byte    // JSON，按 Type + Version 反序列化
    OccurredAt  time.Time
    Actor       Actor     // principal | moderator | participant | system
}
```

**Version 字段**（D-E03）：payload schema 变更时递增 Version，旧 Event 仍可按 Version 反序列化。

### 3. v0.1 核心 Event 清单

在 domain 文档基础上，v0.1 **纳入**以下 Event：

| Event | 触发者 | Payload 要点 |
|-------|--------|--------------|
| `MeetingCreated` | Principal / System | topic, agenda, strategy, confirmation_mode |
| `ParticipantInvited` | Moderator | participant_id, role, expertise, goal |
| `RoundStarted` | Moderator | round_number, order[] |
| `ParticipantResponded` | Participant | participant_id, round_number, content, stance |
| `RoundCompleted` | Moderator | round_number, summary |
| `ConsensusReached` | Strategy / Moderator | strategy, dissent[], resolved_by |
| `ConsensusVetoed` | Principal | reason |
| `ConsensusForced` | Principal | reason |
| `ConfirmationPrepared` | Moderator | cycle, brief, items[] |
| `ConfirmationPresented` | Moderator | cycle |
| `ConfirmationApproved` | Principal | cycle, item_notes |
| `ConfirmationRejected` | Principal | cycle, feedback, item_notes |
| `ConfirmationSkipped` | Principal | reason |
| `ConfirmationForced` | Principal | cycle, reason |
| `MeetingPaused` | Principal / Moderator | reason |
| `MeetingResumed` | Principal / Moderator | — |
| `MeetingFinished` | Moderator / Principal | — |
| `ArtifactProduced` | Participant / Moderator | artifact_id, type, ref |
| `ActionItemCreated` | Moderator | action_item_id, assignee, description |

**v0.1 不纳入**：

| Event | 原因 |
|-------|------|
| `OpinionUpdated` | 合并进 `ParticipantResponded.stance`，避免冗余 |
| `ParticipantAssigned` | 合并进 `ParticipantInvited` 或后续 `ParticipantUpdated`（v0.2） |

### 4. ParticipantResponded Payload（D-E02）

```go
type ParticipantRespondedPayload struct {
    ParticipantID string
    RoundNumber   int
    Content       string        // 发言全文
    Stance        Stance        // agree | object | abstain | none
    ObjectReason  string        // stance=object 时必填
}
```

- **Content**：存全文，Minutes 与审计需要原始发言
- **Stance**：供 ConsensusStrategy 判定，与 Content 分离
- 不在 Event 层存 Opinion delta；Opinion 由 fold 推导为 Meeting State 的投影字段

### 5. 回放能力（D-E04）

| 能力 | v0.1 | 后续 |
|------|------|------|
| **审计只读回放** | ✅ 按 sequence 重放 Event 列表 | — |
| **State 重建** | ✅ fold 验证 Snapshot 一致性 | — |
| **Fork Meeting** | ❌ | 从 sequence=k 分叉为新 Meeting（v0.2） |
| **Time travel 编辑** | ❌ 永不支持 | — |

### 6. Pause / Resume（D-E05）

**纳入 v0.1**。State machine 已定义 Paused 状态，缺少对应 Event 会导致状态不一致。

Pause 行为：

- 当前 Speaking 的 Participant 完成当前回应后暂停（不中断 mid-response）
- 暂停期间拒绝新的 `RoundStarted`、`ParticipantInvited`
- Resume 后从 Paused 前的 Round 状态继续

### 7. 跨 Meeting Knowledge（D-E06）

Knowledge 更新 **不属于 Meeting Event 流**。

```
Meeting Event Stream     →  Meeting 生命周期
Knowledge Event Stream   →  跨 Meeting 持久知识（独立 aggregate）
```

Meeting 仅持有 Knowledge 引用（ID 列表），不 embed 知识内容。  
Knowledge 的 CRUD 产生独立 Event，由 Memory Adapter 持久化，Domain 层只定义 `KnowledgeRef`。

---

## State 归约规则

Meeting State 由 Event fold 推导，核心字段：

```go
type MeetingState struct {
    Status       MeetingStatus
    Topic        string
    Agenda       []AgendaItem
    Participants map[string]ParticipantState
    CurrentRound int
    RoundOrder   []string
    Consensus    *ConsensusState  // nil until reached
    Minutes      MinutesDraft     // 累积中，Completed 时定稿
    Artifacts    []ArtifactRef
    ActionItems  []ActionItem
}
```

非法 Event 转换（如在 Completed 后追加 `RoundStarted`）在 apply 层拒绝，不写入 log。

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| 纯 Snapshot、无 Event Sourcing | 丢失审计与回放，违背 Domain 原则 |
| 全量 fold、无 Snapshot | v0.1 可行但长 Meeting 性能差 |
| OpinionUpdated 独立 Event | 与 ParticipantResponded 重复 |
| Knowledge 写入 Meeting Event | 污染 Meeting 边界，跨 Meeting 生命周期不同 |
| Event 可修改 | 破坏不可变审计链 |

---

## 后果

### 正面

- Source of truth 清晰，Minutes 可从 Event 策展生成
- Snapshot 保证读取性能可扩展
- Version 字段支持 schema 演进
- Meeting / Knowledge 生命周期解耦

### 负面

- Event Sourcing 增加实现复杂度（sequence、并发、Snapshot 一致性）
- ParticipantResponded 存全文，长发言 log 体积大——v0.2 可考虑 blob 外置

### 待实现

- [ ] `EventEnvelope` 与 v0.1 Event Type 常量
- [ ] `Apply(event) → MeetingState` 纯函数
- [ ] Append-only log 接口 + 文件实现
- [ ] Snapshot 触发与加载
- [ ] 非法转换校验

---

## 决议项对照

| 编号 | 决议 |
|------|------|
| D-E01 | Event Sourcing + 周期性 Snapshot |
| D-E02 | ParticipantResponded 存全文 + Stance |
| D-E03 | Envelope 含 Version 字段 |
| D-E04 | v0.1 仅审计回放与 State 重建 |
| D-E05 | Pause / Resume 纳入 v0.1 |
| D-E06 | Knowledge 独立 Event 流，Meeting 仅持引用 |
