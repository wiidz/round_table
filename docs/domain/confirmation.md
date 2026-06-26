# Confirmation

Confirmation（用户确认关）是 Consensus 达成之后、输出最终结论之前的**可选环节**。

Participant 之间的 Consensus 只代表「专家团队内部达成一致」；Confirmation 代表 **Principal 对结论是否符合预期的终审**。

---

## 已定义

### 定义

Moderator 在 Consensus 后整理一份 **Confirmation Brief（确认清单）**，列出 Principal 需要逐项确认的事项。Principal 审阅后给出批准或驳回；批准则输出最终结论，驳回则 Meeting 回到讨论阶段继续推进。

### 与相关概念的区别

| 概念 | 谁参与 | 回答的问题 |
|------|--------|------------|
| **Consensus** | Participant 之间 | 专家们是否达成一致？ |
| **Confirmation** | Principal + Moderator | 结论是否符合 Principal 的预期？ |
| **Completed** | — | 最终结论与 Artifacts 已输出 |

> Consensus 是集体决策；Confirmation 是 Principal 验收。

### 组成

| 属性 | 说明 |
|------|------|
| **Confirmation Brief** | 本次确认清单的整体描述 |
| **Confirmation Item[]** | 编号待确认事项（1、2、3、4…） |
| **Principal Response** | Principal 的批准 / 驳回及反馈 |
| **Cycle** | 第几次确认（驳回后重新讨论再确认，cycle 递增） |

### Confirmation Item 结构

| 字段 | 说明 |
|------|------|
| **序号** | 1, 2, 3, 4… |
| **Title** | 简短标题 |
| **Description** | 该项的具体内容（Moderator 归纳） |
| **Source** | 来源说明（如「综合 Round 2 讨论」） |

Principal 可按整体批准，也可对单项附注（`item_notes`）。

### 生命周期

```
ConsensusReached
  → [confirmation_mode = skip]    → MeetingFinished → Completed
  → [confirmation_mode = required]  → Confirmation → Principal 审阅
       → Approved  → MeetingFinished → Completed
       → Rejected  → Running（继续讨论）→ 再次 Consensus → 再次 Confirmation
```

详见 [state_machine.md](../flow/state_achine.md)。

### 配置

创建 Meeting 时设置 `confirmation_mode`：

| 模式 | 行为 |
|------|------|
| **`skip`** | Consensus 后直接输出最终结论，跳过 Confirmation |
| **`required`**（推荐） | 必须经过 Principal 确认关 |

配套参数：

| 参数 | 默认 | 说明 |
|------|------|------|
| `max_confirmation_cycles` | 3 | Principal 驳回后最多重新确认的次数 |
| `rounds_per_cycle` | 由 ADR-0005 定义 | 每轮驳回后追加的讨论 Round 上限 |

### 设计约束

- Confirmation 由 **Moderator 整理**，Participant 不直接与 Principal 对话
- Principal 驳回时的 **Feedback** 作为下一轮讨论的输入，写入 Event 并注入 Participant 上下文
- 跳过 Confirmation 不等于跳过 Principal——`skip` 模式下 Principal 仍可在 Consensus 前通过 Veto 介入（见 ADR-0002）
- Confirmation Brief 是 Minutes 的一部分，Completed 后只读

---

## 已决议（ADR-0004）

详见 [ADR-0004-principal-confirmation.md](../architecture/ADR-0004-principal-confirmation.md)。

| 编号 | 决议 |
|------|------|
| D-CF01 | 新增 **Confirmation** 状态，位于 Consensus 与 Completed 之间 |
| D-CF02 | `confirmation_mode: skip \| required`，默认 `required` |
| D-CF03 | Moderator 整理 **Confirmation Brief**，含编号 Confirmation Item |
| D-CF04 | Principal **Approved** → 输出最终结论；**Rejected** → 回到 Running |
| D-CF05 | 默认最多 **3 次** Confirmation 循环，超出后 Principal 选择强制批准 / 继续 / 终止 |
| D-CF06 | 取代 ADR-0002 的 `completion_mode: manual`，统一为 Confirmation 模型 |

---

## 关联

- 父索引：[README.md](./README.md)
- 委托人：[principal.md](./principal.md)
- 前置：[consensus.md](./consensus.md)
- 调度：[moderator.md](./moderator.md)
- 事件：[event.md](./event.md)
- 状态机：[state_machine.md](../flow/state_achine.md)
