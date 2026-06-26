# Consensus

Consensus 表示 Meeting 是否已达成** sufficient agreement**（足够的一致）。

---

## 已定义

### 定义

Consensus 是 **Meeting 级属性**，不是 Participant 的属性。单个 Participant 持有的是 Opinion；Consensus 描述整个 Meeting 是否可以在当前 Agenda 上做出决定。

### 与相关概念的区别

| 概念 | 粒度 | 说明 |
|------|------|------|
| **Opinion** | Participant | 个人当前观点，可变化 |
| **Consensus** | Meeting | 集体是否达成一致 |
| **Minutes** | Meeting | 一致后的结构化记录 |
| **Artifact** | Meeting | 一致后的产出物 |

### 触发时机

- 每 Round 的 Summary 之后进行 Consensus Check
- 达成时产生 Event：`ConsensusReached`
- Meeting Status 进入 `Consensus` 状态（见 [state_machine.md](../flow/state_achine.md)）
- Consensus 之后进入 [Confirmation](./confirmation.md)（Principal 确认关）或直接进入 Completed（`confirmation_mode: skip`）

### 设计约束

- 目标是一致（Consensus over Completion），不是某个 Agent 独自完成任务
- 判定逻辑应可插拔（策略模式），Domain 定义接口，具体算法由 ADR 决议
- Moderator 负责**检测与触发**，不一定拥有最终判定权（取决于策略）

---

## 已决议（ADR-0002）

详见 [ADR-0002-consensus-strategy.md](../architecture/ADR-0002-consensus-strategy.md)。

| 编号 | 决议 |
|------|------|
| D-C01 | 默认 **No Objection**，Round 上限僵局时 **Moderator Decision** 兜底 |
| D-C02 | `ConsensusStrategy` 可插拔，Meeting 创建时可配置，系统默认 No Objection |
| D-C03 | 部分一致算 Consensus；反对理由写入 Minutes 的 **Dissenting Opinions** |
| D-C04 | Principal 拥有 Veto、Force Consensus、Override Strategy 权限 |
| D-C05 | ~~`completion_mode`~~ → 由 [ADR-0004](../architecture/ADR-0004-principal-confirmation.md) 的 `confirmation_mode` 取代 |

### v0.1 实现的策略

| 策略 | 行为 |
|------|------|
| `NoObjection` | 无 Participant 标记 `object` → 通过（**默认**） |
| `Unanimous` | 所有 Participant 标记 `agree` → 通过 |
| `Supermajority` | `agree / (agree + object)` ≥ 阈值 → 通过 |

Participant 表态：`agree` | `object` | `abstain`。Vote 策略推迟至 v0.2。

---

## 关联

- 父索引：[README.md](./README.md)
- 权威定义：[CONSTITUTION.md](../CONSTITUTION.md) § Core Concepts — Consensus
- 原则：[PRINCIPLES.md](../PRINCIPLES.md) — Principle 4: Consensus over Completion
- 检测方：[moderator.md](./moderator.md)
- 委托人：[principal.md](./principal.md)
- 后续环节：[confirmation.md](./confirmation.md)
- 事件：[event.md](./event.md) — ConsensusReached
