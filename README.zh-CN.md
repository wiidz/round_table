[English](README.md) | **中文**

# RoundTable

> **Build AI Teams, not AI Agents.**

*One problem. Many minds. One decision.*

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://github.com/wiidz/round_table/actions/workflows/ci.yml/badge.svg)](https://github.com/wiidz/round_table/actions/workflows/ci.yml)

RoundTable 是一个 **Multi-Agent Meeting Engine（多智能体会议引擎）**——用结构化会议的方式，让多个 AI 专家讨论、辩论并达成共识，而不是堆叠一个更强的单体 Agent。

---

## 快速启动

两种方式任选其一。**只想先看看界面？** 两种方式都可以先 `make seed-demo`，无需 API Key 即可在 Web 里浏览一场示例会议。

### 方式 A · 本机开发（Go + Node）

**需要：** Go 1.25+、Node.js 22+、Make（见 `apps/server/go.mod`）

```bash
git clone https://github.com/wiidz/round_table.git
cd round_table

cp deploy/.env.example deploy/.env   # 仅浏览演示可暂不填 Key

make seed-demo      # 导入示例会议（无需 API Key）
make server-dev     # 终端 1：API → http://localhost:7777
make web-dev        # 终端 2：Web → http://localhost:5173
```

打开 Web → **会议** → `mtg-demo-001`，可查看概览、文档、流程与回放。

**在 Discord 里真实跑会**（需配置 `DEEPSEEK_API_KEY` + `DISCORD_BOT_TOKEN`）：

```bash
make run-discord    # 另开终端；大陆访问 Discord 见 Makefile 内代理说明
```

更细的步骤见 [docs/getting-started.md](./docs/getting-started.md)。

### 方式 B · Docker 一键部署（推荐上服务器）

**需要：** Docker 24+、Docker Compose v2（生产环境建议 Linux；Mac Docker Desktop 的 host 网络行为不同）

```bash
git clone https://github.com/wiidz/round_table.git
cd round_table

cp deploy/.env.example deploy/.env   # 填入 DEEPSEEK_API_KEY、DISCORD_BOT_TOKEN
sh deploy/init-data-dirs.sh

make docker-up      # 构建并启动 Web + API + Discord Supervisor
```

默认访问 <http://127.0.0.1:7777>。日志：`make docker-logs` / `make docker-logs-discord`。

详见 [deploy/README.md](./deploy/README.md)。

---

## 核心名词（先读这个）

| 名词 | 是谁 | 做什么 |
|------|------|--------|
| **会议（Meeting）** | 整场讨论的单位 | 围绕一个 **议题（Topic）** 从会前准备到结案，产出纪要、交付物与 Token 统计；一切状态由 **事件（Event）** 驱动并可审计 |
| **委托人（Principal）** | 人类决策者 | 发起会议、设定议题与边界，拥有最终 **验收权**；确认关里可批准或驳回方案 |
| **主持人（Moderator）** | 调度型 Agent | **控场**——安排发言顺序、轮次总结、检测是否可进入共识/合成；提供专业判断的是专家，不是主持人 |
| **专家（Participant）** | 领域 Agent | 各守角色与专长（如策划、研发、运营），**只在被邀请时发言**，彼此不直连 |
| **会议流程** | Engine 标准路径 | 会前准备（Round 0）→ 辩论/研讨轮次（Round 1+）→ 可选自由问答 → 主持人总结 → 共识/方案合成 → 可选 **委托人确认** → 结案产出 |

研讨型（deliberation）侧重方案草案；裁决型（decision）侧重可执行共识。驳回后会追加研讨并再次呈报确认。

```
委托人定议题 → 主持人调度 → 专家多轮发言 → 内部共识 → [委托人确认] → 纪要 / 交付物
                              ↑                    │
                              └──── 驳回后追加研讨 ─┘
```

领域细节：[docs/domain/](./docs/domain/README.md) · 架构宪法：[CONSTITUTION.md](./docs/CONSTITUTION.md)

---

## 当前支持与后续扩展

| 能力 | 现状 |
|------|------|
| **大模型** | ✅ [DeepSeek](https://platform.deepseek.com/)（`DEEPSEEK_API_KEY`） |
| **对外通道（Transport）** | ✅ **Discord**（Bot 跑会、确认关、自由问答） |
| **工作台** | ✅ Web UI（浏览会议、文档、流程、圆桌回放） |
| **后续** | 🔜 更多模型供应商、更多 Transport（Slack、企业 IM 等）——Engine 核心与 Adapter 解耦，见下方「架构独立性」 |

CLI 本地跑会：`apps/server/cmd/meet`（开发调试用，非面向最终用户的主路径）。

---

## 它是什么 / 不是什么

| | |
|---|---|
| ❌ 另一个 AI Agent | ✅ 协调多个专家的 Meeting Engine |
| ❌ Agent Runtime（LangGraph / AutoGen / CrewAI） | ✅ 独立于 Runtime 的领域引擎 |
| ❌ Workflow Engine | ✅ 结构化讨论，不是 DAG 任务流 |
| ❌ 聊天机器人 | ✅ Chat 只是界面，Discussion 才是架构 |

---

## 项目状态

🚧 **Phase 1** — Engine 可本地端到端跑会。  
🚧 **Phase 1.5** — Discord Transport 可完整跑会（委托人绑定 → 简报向导 → 预设菜单 → 确认关 → 交付物）。

Engine 已实现：Event Sourcing、Pre-meeting、多轮辩论/研讨、自由对话、主持人摘要、共识/合成、Confirmation（含驳回与轮次上限）、Workspace 投影、Token 统计。  
Discord 已实现：自然语言/`!rt` 发起、三步简报向导、确认关、运行期干预、Executive Recap、多 Bot、中文 i18n 等。详见 [docs/adapters/discord-transport.md](./docs/adapters/discord-transport.md)。

```
apps/server/cmd/discord/         # Discord Transport
apps/server/cmd/roundtable/      # HTTP API
apps/server/internal/engine/     # Meeting Engine
apps/web/                        # React 工作台
```

路线图：[docs/roadmap.md](./docs/roadmap.md)

---

## 文档

| 文档 | 说明 |
|------|------|
| [getting-started.md](./docs/getting-started.md) | 分步上手（演示数据、Discord、开发细节） |
| [CONTRIBUTING.md](./CONTRIBUTING.md) | 贡献指南 |
| [SECURITY.md](./SECURITY.md) | 安全报告 |
| [CONSTITUTION.md](./docs/CONSTITUTION.md) | 架构宪法 |
| [domain/](./docs/domain/README.md) | 领域模型详解 |
| [deploy/README.md](./deploy/README.md) | Docker 部署 |
| [apps/web/README.md](./apps/web/README.md) | Web UI 说明 |
| [architecture/](./docs/architecture/README.md) | ADR 索引 |

---

## 开发与测试（贡献者）

日常命令（完整列表见根目录 `Makefile`）：

| 命令 | 用途 |
|------|------|
| `make seed-demo` | 导入演示会议与档案 |
| `make server-dev` | API 热重载 |
| `make web-dev` | 前端开发服务器 |
| `make run-discord` | 本地 Discord Bot |
| `make docker-up` | Docker compose 启动 |
| `make sync-data-pull` | 从部署机拉取 `data/`（见 `deploy/sync-data.sh`） |

**测试与 CI：**

```bash
make test                    # Go 单元/集成测试
cd apps/web && npm test      # 前端 Vitest
cd apps/web && npm run build # 前端生产构建
```

CI 配置：`.github/workflows/ci.yml`。贡献前请阅读 [CONTRIBUTING.md](./CONTRIBUTING.md) 与 [COMMITS.md](./docs/COMMITS.md)。

---

## 设计原则

1. **Everything is a Meeting** — 不是 Workflow，不是 Prompt，不是 Agent  
2. **Moderator controls the discussion** — 专家不能抢话  
3. **Participants own expertise** — 各守 Role 边界  
4. **Consensus over Completion** — 团队一致，而非单体执行  
5. **Discussion is structured** — Round / Agenda / Minutes / Consensus / Confirmation  
6. **Principal owns the decision** — 最终决策权在委托人  

详见 [PRINCIPLES.md](./docs/PRINCIPLES.md)。

---

## 架构独立性

Meeting Engine **核心领域不依赖**具体 Runtime、Model 或 Transport，均通过 Adapter 接入。详见 [CONSTITUTION.md](./docs/CONSTITUTION.md)。

---

## License

[Apache 2.0](LICENSE)
