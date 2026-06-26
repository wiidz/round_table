# RoundTable

> **Build AI Teams, not AI Agents.**

*One problem. Many minds. One decision.*

RoundTable 是一个 **Multi-Agent Meeting Engine（多智能体会议引擎）**——用结构化会议的方式，让多个 AI 专家讨论、辩论并达成共识，而不是堆叠一个更强的单体 Agent。

---

## 它是什么

RoundTable 把复杂问题建模为一场 **Meeting（会议）**：

- **Principal（委托人）** 提出议题，拥有最终验收权
- **Moderator（司仪）** 控制发言顺序、总结讨论、检测共识
- **Participant（专家）** 各司其职，只在被邀请时发言
- 多轮 **Round** 推进讨论，产出 **Consensus**、**Minutes** 与 **Artifacts**

目标不是「Agent 自动干完」，而是**协作推理、集体决策**。

## 它不是什么

| | |
|---|---|
| ❌ 另一个 AI Agent | ✅ 协调多个专家的 Meeting Engine |
| ❌ Agent Runtime（LangGraph / AutoGen / CrewAI） | ✅ 独立于 Runtime 的领域引擎 |
| ❌ Workflow Engine | ✅ 结构化讨论，不是 DAG 任务流 |
| ❌ 聊天机器人 | ✅ Chat 只是界面，Discussion 才是架构 |

---

## 核心流程

```
Principal → Moderator → Participant → … → Consensus → Confirmation → Decision
                                              ↑              │
                                              └── Rejected ──┘
                                                   （继续讨论）
```

1. **Principal** 创建 Meeting，设定 Topic / Agenda  
2. **Moderator** 按 Round 调度 Participant 发言  
3. Participant 达成 **Consensus**（内部一致）  
4. **Confirmation**（可选）：Moderator 整理确认清单，Principal 批准或驳回  
5. 输出 **Minutes**、**Artifacts**、**Action Items**

`confirmation_mode: skip` 时可跳过第 4 步，完全交由 Meeting 自行得出结论。

---

## 项目状态

🚧 **Phase 0 → Phase 1** — 领域文档与 ADR 已完成；Monorepo 骨架已初始化（`apps/server` + `go test ./apps/server/...`）。

```
apps/server/cmd/roundtable/     # 服务入口（/health）
apps/server/internal/domain/    # 纯领域（Meeting、Event、Consensus）
apps/server/internal/engine/    # Meeting Engine 编排
apps/server/internal/scheduler/ # Moderator Fixed Order
apps/server/internal/adapter/   # storage、participant 等端口
apps/server/internal/platform/  # config、HTTP server（stdlib）
```

结构说明见 [ADR-0008](./docs/architecture/ADR-0008-project-structure.md)。

```bash
make test   # 运行测试
make run    # 启动 :8080 /health
```

---

## 文档

| 文档 | 说明 |
|------|------|
| [VISION.md](./docs/VISION.md) | 项目愿景 |
| [CONSTITUTION.md](./docs/CONSTITUTION.md) | 架构宪法（权威定义） |
| [PRINCIPLES.md](./docs/PRINCIPLES.md) | 六条设计原则 |
| [NAMING.md](./docs/NAMING.md) | 命名约定与 rationale |
| [COMMITS.md](./docs/COMMITS.md) | Git Commit 规范 |
| [domain/](./docs/domain/README.md) | 领域概念（Meeting、Principal、Moderator…） |
| [flow/](./docs/flow/context_diagram.md) | 状态机与上下文图 |
| [architecture/](./docs/architecture/README.md) | 架构决策记录（ADR） |

### 领域概念速览

| 概念 | 一句话 |
|------|--------|
| [Meeting](./docs/domain/meeting.md) | 围绕一个主题的完整结构化讨论 |
| [Principal](./docs/domain/principal.md) | 委托人，发起 Meeting 并拥有最终验收权 |
| [Moderator](./docs/domain/moderator.md) | 调度者，控场但不提供专业判断 |
| [Participant](./docs/domain/participant.md) | 领域专家，只响应邀请、不直连 |
| [Round](./docs/domain/round.md) | Meeting 内的一轮有序讨论 |
| [Consensus](./docs/domain/consensus.md) | Participant 之间是否达成一致 |
| [Confirmation](./docs/domain/confirmation.md) | Principal 确认关，审阅结论是否符合预期 |
| [Event](./docs/domain/event.md) | 领域事件，驱动状态变更与审计 |

### 已接受的 ADR

| ADR | 主题 |
|-----|------|
| [ADR-0002](./docs/architecture/ADR-0002-consensus-strategy.md) | Consensus 判定策略 |
| [ADR-0003](./docs/architecture/ADR-0003-event-model.md) | Event 模型与持久化 |
| [ADR-0004](./docs/architecture/ADR-0004-principal-confirmation.md) | Principal Confirmation 确认关 |
| [ADR-0005](./docs/architecture/ADR-0005-round-termination.md) | Round 终止条件 |
| [ADR-0007](./docs/architecture/ADR-0007-moderator-scheduling.md) | Moderator 调度策略 |

---

## 设计原则

1. **Everything is a Meeting** — 不是 Workflow，不是 Prompt，不是 Agent  
2. **Moderator controls the discussion** — Participant 不能抢话  
3. **Participants own expertise** — 各守 Role 边界  
4. **Consensus over Completion** — 团队一致，而非单体执行  
5. **Discussion is structured** — Round / Agenda / Minutes / Consensus / Confirmation  
6. **Principal owns the decision** — 最终决策权在委托人  

详见 [PRINCIPLES.md](./docs/PRINCIPLES.md)。

---

## 架构独立性

Meeting Engine 核心领域**不依赖**：

- **Runtime** — OpenClaw、LangGraph、AutoGen、CrewAI…
- **Model** — OpenAI、Anthropic、DeepSeek…
- **Transport** — Discord、Slack、Web UI…

上述均通过 Adapter 层接入。详见 [CONSTITUTION.md](./docs/CONSTITUTION.md)。

---

## 参与贡献

本项目处于早期架构阶段。贡献前请：

1. 阅读 [CONSTITUTION.md](./docs/CONSTITUTION.md) 与 [NAMING.md](./docs/NAMING.md)  
2. 遵循领域语言，不引入未经讨论的核心概念  
3. 架构变更请先写 ADR  
4. Commit 遵循 [COMMITS.md](./docs/COMMITS.md) 结构化格式（可配置 `git config commit.template .gitmessage`）

---

## License

待定。
