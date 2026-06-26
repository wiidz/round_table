# Domain Concepts

RoundTable 的领域概念索引。权威定义见 [CONSTITUTION.md](../CONSTITUTION.md) v0.2。

> Everything is a Meeting.

---

## 概念层级

```
Meeting（聚合根）
├── Principal（委托人，发起者与最终验收者）
├── Topic
├── Agenda
├── Moderator
├── Participant[]
│   ├── Role / Expertise / Goal
│   ├── Opinion
│   └── Status
├── Round[]
│   ├── Order
│   ├── Responses[]
│   └── Summary
├── FreeDialogue（Round 1 后，可选）
├── TokenUsage（LLM 调用统计）
├── Consensus
├── Confirmation（Principal 确认关，可选）
├── Minutes
├── Artifact[]
├── ActionItem[]
├── Knowledge（跨 Meeting 引用）
└── Status
```

Event 横切所有状态变更，是 Meeting State 的唯一来源。

---

## 核心概念

| 概念 | 文档 | 一句话 |
|------|------|--------|
| Meeting | [meeting.md](./meeting.md) | 围绕一个主题的完整结构化讨论 |
| Principal | [principal.md](./principal.md) | 委托人，发起 Meeting 并拥有最终验收权 |
| Moderator | [moderator.md](./moderator.md) | 调度者，控场但不提供专业判断 |
| Participant | [participant.md](./participant.md) | 领域专家，只响应邀请、不直连 |
| Round | [round.md](./round.md) | Meeting 内的一轮有序讨论 |
| Consensus | [consensus.md](./consensus.md) | Participant 之间是否达成一致 |
| Confirmation | [confirmation.md](./confirmation.md) | Principal 确认关，审阅结论是否符合预期 |
| Event | [event.md](./event.md) | 领域事件，驱动状态变更与审计 |
| Workspace | [workspace.md](./workspace.md) | Meeting 文件产出区（Minutes、Artifacts、usage/） |
| Profile | [profile.md](./profile.md) | Participant/Principal 身份文件（SOUL、USER） |
| Knowledge | [knowledge.md](./knowledge.md) | 跨 Meeting 长期记忆 |

---

## 关联概念（见 CONSTITUTION）

| 概念 | 归属 | 说明 |
|------|------|------|
| Agenda | Meeting | 讨论目标，可多项 |
| Opinion | Participant | 当前观点，随 Round 演化 |
| Minutes | Meeting | 结构化纪要，不是 chat log |
| Artifact | Meeting | 产出物（文档、代码、设计等） |
| Action Item | Meeting | 后续待办 |
| Knowledge | 跨 Meeting | 长期持久知识 |

---

## 流程与状态

| 文档 | 内容 |
|------|------|
| [state_machine.md](../flow/state_achine.md) | Meeting 生命周期 |
| [participant flow](../flow/participant.md) | Participant 运行时状态 |
| [context_diagram.md](../flow/context_diagram.md) | 系统上下文 |

---

## 非 Domain 层

以下概念不属于核心领域，将在 adapter / infrastructure 层定义：

| 层 | 示例 | 计划位置 |
|----|------|----------|
| Runtime Adapter | OpenClaw, LangGraph, AutoGen | `docs/adapters/runtime.md` |
| Model Adapter | OpenAI, Anthropic, DeepSeek | `docs/adapters/model.md` |
| Transport Layer | Discord, Slack, Web UI | `docs/adapters/transport.md` |
| Memory Store | 向量库、持久化存储 | `docs/adapters/memory.md` |

Domain 层不得引用上述具体实现。

---

## 命名约定

完整说明见 [NAMING.md](../NAMING.md)。核心原则：**命名决定设计**。

| 使用 | 避免 |
|------|------|
| Meeting | Task, Workflow |
| Principal | Main Agent, Boss Agent |
| Moderator | Main Agent, Boss Agent, TaskManager |
| Participant | Sub Agent |
| Minutes | Chat History |
| Knowledge | Memory（对外命名） |
| Round | Step, Node |

---

## 架构决策（ADR）

| ADR | 主题 | 状态 |
|-----|------|------|
| [ADR-0002](../architecture/ADR-0002-consensus-strategy.md) | Consensus 判定策略 | Accepted |
| [ADR-0003](../architecture/ADR-0003-event-model.md) | Event 模型与持久化 | Accepted |
| [ADR-0004](../architecture/ADR-0004-principal-confirmation.md) | Principal Confirmation 确认关 | Accepted |
| [ADR-0005](../architecture/ADR-0005-round-termination.md) | Round 终止条件 | Accepted |
| [ADR-0006](../architecture/ADR-0006-knowledge-scope.md) | Knowledge 作用域 | Accepted |
| [ADR-0007](../architecture/ADR-0007-moderator-scheduling.md) | Moderator 调度策略 | Accepted |
| [ADR-0009](../architecture/ADR-0009-meeting-workspace.md) | Meeting Workspace | Accepted |
| [ADR-0010](../architecture/ADR-0010-agent-profiles.md) | Agent Profile | Accepted |

完整索引：[architecture/README.md](../architecture/README.md)

---

## 待决策汇总

| 编号 | 主题 | 涉及概念 | 状态 |
|------|------|----------|------|
| D-001 | Consensus 判定策略 | Consensus, Moderator | ✅ ADR-0002 |
| D-002 | Participant 知识作用域 | Participant, Knowledge | ✅ ADR-0006 |
| D-003 | Event 持久化与回放 | Event, Minutes | ✅ ADR-0003 |
| D-004 | Round 终止条件 | Round, Consensus | ✅ ADR-0005 |
| D-005 | Principal Confirmation 确认关 | Confirmation, Principal | ✅ ADR-0004 |
| D-006 | Moderator 调度策略 | Moderator, Scheduler | ✅ ADR-0007 |
| D-007 | Meeting Workspace | Meeting, Artifact, Minutes | ✅ ADR-0009 |
| D-008 | Agent Profile | Participant, Principal | ✅ ADR-0010 |

各概念文档中的「已决议 / 待决策」栏为详细说明。