# Meeting

Meeting 是 RoundTable 的**最高层抽象**（聚合根）。所有复杂问题都被建模为一场结构化会议。

> One problem. Many minds. One decision.

---

## 已定义

### 定义

Meeting 表示围绕一个 **Topic** 的完整讨论过程，由 **Principal** 发起，Moderator 调度多个 Participant，通过多轮 Round 推进，最终产出 Consensus、Minutes 与 Artifacts。

### 组成

| 属性 | 说明 |
|------|------|
| **Topic** | 讨论主题 |
| **MeetingMode** | `decision`（裁决型，默认）或 `deliberation`（研讨型，合成方案草案） |
| **Principal** | 委托人，发起者兼最终验收者（见 [principal.md](./principal.md)） |
| **Agenda** | 讨论目标，可包含一项或多项 |
| **Moderator** | 调度者，负责 orchestration |
| **Participants** | 领域专家集合 |
| **Rounds** | 多轮有序讨论 |
| **Consensus** | Participant 是否达成一致（Meeting 级属性） |
| **Confirmation** | Principal 确认关（可选，见 [confirmation.md](./confirmation.md)） |
| **Minutes** | 结构化纪要，非 chat history |
| **Artifacts** | 产出物（文档、代码、设计等） |
| **Action Items** | 后续待办 |
| **Knowledge** | 跨 Meeting 引用的持久知识 |
| **Status** | 生命周期状态 |
| **FreeDialogue** | Round 1 后固定一次的互相 Q&A（可配置关闭） |
| **TokenUsage** | 每次 LLM 调用的 token 统计（prompt / completion / total） |

### 会议阶段（Running 内）

`max_rounds_per_segment` **仅计辩论轮 1+**，不含 Pre-meeting。

```
Round 0 (Pre-meeting)     各 Participant 独立提交视角，互不可见
    ↓
Round 1 (Debate)          按 Order 发言 + Stance
    ↓
Free Dialogue             Round 1 后 Q&A（free_dialogue_max_questions，0=关闭）
    ↓
Moderator Summary         提炼 Round 1 要点
    ↓
Round 2 … N               重复辩论 →（每轮后 Moderator 总结）→ Consensus 检查
    ↓
Consensus → Confirmation? → Completed
```

配置：`meeting.free_dialogue_max_questions`（默认 1，每人提问轮数；总 Q&A 次数 = 该值 × 参会人数）。

### Meeting Mode（ADR-0011）

| 模式 | 说明 | 终止 |
|------|------|------|
| `decision` | 是否批准类议题，`agree/object` + Consensus | No Objection 或 Moderator 兜底 |
| `deliberation` | 方案共建，各角色贡献设计点 | Moderator 合成就绪检测或达 `max_rounds` → `artifacts/design-draft.md` |

创建时：`MeetingCreated.meeting_mode`；CLI：`meet -mode deliberation`。

### 生命周期

```
Created → Preparing → Running → Paused → Consensus → Confirmation → Completed → Archived
                                                      ↘ (skip) ↗
```

`Confirmation` 在 `confirmation_mode: required` 时启用；`skip` 时 Consensus 后直接 Completed。

详见 [state_machine.md](../flow/state_achine.md)。

### 设计约束

- Meeting 是产品，Agent 是实现细节
- 一个 Meeting 有且仅有一个 Principal
- 所有通信经 Moderator，Participant 不直连
- Minutes 是策展后的知识，不是原始对话记录
- Knowledge 长期持久；Minutes 仅属当前 Meeting
- 物理产出（md 文件）在 [Workspace](./workspace.md)（ADR-0009）

---

## 待决策

| 编号 | 问题 | 选项 / 备注 |
|------|------|-------------|
| D-M01 | Agenda 与 Topic 的关系 | Topic 是总主题，Agenda 是可逐项推进的子目标？ |
| D-M02 | 一个 Meeting 是否允许多 Topic | 单 Topic 简化模型；多 Topic 需子 Meeting 或 Agenda 拆分 |
| D-M03 | Paused 状态下是否允许修改 Participants | 影响 Preparing / Running 边界 |
| D-M04 | Archived 与 Completed 的数据保留策略 | Minutes / Artifacts 是否只读、是否可 fork 新 Meeting |
| D-M05 | Knowledge 注入时机 | Meeting 创建时加载 vs 每 Round 按需加载 |

---

## 关联

- 父索引：[README.md](./README.md)
- 委托人：[principal.md](./principal.md)
- 权威定义：[CONSTITUTION.md](../CONSTITUTION.md) § Core Concepts — Meeting
- 上下文：[context_diagram.md](../flow/context_diagram.md)
