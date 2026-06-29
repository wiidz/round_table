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

**单次 Meeting 内（Running）**：

```
Pre-meeting (R0) → Debate (R1…) → [R1 后 Free Dialogue] → Moderator 总结 → Consensus
```

1. **Principal** 创建 Meeting，设定 Topic / Agenda  
2. **Moderator** 调度 Pre-meeting、辩论 Round、Round 1 后自由对话、轮间总结  
3. Participant 达成 **Consensus**（内部一致）  
4. **Confirmation**（可选）：Moderator 整理确认清单，Principal 批准或驳回  
5. 输出 **Minutes**、**Artifacts**、**Action Items**、**usage/**（Token 统计）

`confirmation_mode: skip` 时可跳过第 4 步。`free_dialogue_max_questions: 0` 可跳过 Round 1 后 Q&A。

本地端到端：`make meet-3round`（需 `DEEPSEEK_API_KEY`）。详见 [data/README.md](./data/README.md)。

---

## 项目状态

🚧 **Phase 1** — Meeting Engine 可本地端到端跑会（DeepSeek + `cmd/meet`）。  
🚧 **Phase 1.5** — **Discord Transport** 可完整跑会（Principal 绑定 → 预设菜单 → 确认关 → 交付物）。

**Engine（CLI）**已实现：Event Sourcing 主循环、Pre-meeting（Round 0）、多轮辩论、Round 1 自由对话（含 Principal turn boundary 代问）、Moderator 轮间摘要、Consensus / Confirmation（含 ItemNotes、上限三选一）、deliberation 合成、Workspace 投影、Token 用量统计。

**Discord**已实现：自然语言/`!rt` 发起、数字预设菜单、确认关交互、运行期干预、Principal 自由问答 `提问`、结束 artifact 推送与按需拉取、多 Bot 发言、中文 i18n、发送重试与重连提示。详见 [docs/adapters/discord-transport.md](./docs/adapters/discord-transport.md)。

```
apps/server/cmd/meet/            # 本地 CLI 跑会
apps/server/cmd/discord/         # Discord Transport bot
apps/server/cmd/roundtable/      # HTTP 服务入口（/health）
apps/server/internal/domain/     # 纯领域（Meeting、Event、Consensus）
apps/server/internal/engine/     # Meeting Engine 编排
apps/server/internal/scheduler/  # Moderator Fixed Order
apps/server/internal/adapter/    # model、participant、workspace、transport…
apps/server/internal/platform/   # config、bootstrap
```

```bash
make test          # 运行测试
make run           # 启动 :7777 /health
make server-dev    # 热重载 API（另开终端）
make web-dev       # Vite 前端 :5173（勿用 6666，Chrome 会拦截）
make run-discord   # 启动 Discord bot（需 DISCORD_BOT_TOKEN）
make stop-discord  # 清理孤儿 Discord 进程（Ctrl+C 后若 Bot 仍在线）
make meet-3round   # 三轮辩论场景（DeepSeek）
make meet TOPIC="…" MEET_FLAGS='-max-rounds 2 -participants "a:Role:x,b:Role:y"'
```

**Ubuntu 服务器 Docker 部署**（单容器 Web + API + Discord Supervisor）见 [deploy/README.md](./deploy/README.md)。

```bash
cp deploy/.env.example deploy/.env   # 填入 DEEPSEEK_API_KEY 等
sh deploy/init-data-dirs.sh
make docker-up                       # host 网络，默认 http://127.0.0.1:7777
make docker-logs
make docker-logs-discord             # data/logs/discord-transport.log
```

结构说明见 [ADR-0008](./docs/architecture/ADR-0008-project-structure.md)。数据目录见 [data/README.md](./data/README.md)。

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
| [adapters/discord-transport.md](./docs/adapters/discord-transport.md) | Discord 指令与频道行为 |
| [roadmap.md](./docs/roadmap.md) | 产品路线图（P0–P4） |
| [deploy/README.md](./deploy/README.md) | Ubuntu Docker 部署 |
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
| [ADR-0006](./docs/architecture/ADR-0006-knowledge-scope.md) | Knowledge 作用域（默认隔离） |
| [ADR-0009](./docs/architecture/ADR-0009-meeting-workspace.md) | Meeting Workspace 文件产出 |
| [ADR-0010](./docs/architecture/ADR-0010-agent-profiles.md) | Agent Profile（SOUL/USER） |
| [ADR-0011](./docs/architecture/ADR-0011-meeting-mode.md) | Meeting Mode（decision / deliberation） |

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