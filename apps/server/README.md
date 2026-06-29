# apps/server

RoundTable **Meeting Engine** 后端（Go）。

Monorepo 内所有 Go 代码集中在此目录，避免与 `apps/web` 等客户端混放。

```
apps/server/
├── cmd/
│   ├── roundtable/     # HTTP/WS 服务入口
│   ├── meet/           # 本地跑一场会议（DeepSeek LLM）
│   ├── discord/        # Discord Transport bot（收发消息）
│   └── migrate/        # SQLite 迁移 CLI
├── configs/            # 运行时配置（无 secret 入库）
├── internal/
│   ├── domain/         # 纯领域（零框架、零 DB）
│   ├── engine/         # Meeting Engine 编排
│   ├── scheduler/      # Moderator 调度
│   ├── adapter/        # storage、participant、transport 等端口
│   └── platform/       # config、HTTP server
└── go.mod
```

仓库根目录 `go.work` 引用本 module。

### 本地开发

Go module 代理（国内网络建议先设置）：

```bash
export GOPROXY=https://goproxy.cn,direct
```

配置分层：`configs/server.yaml` → **`deploy/.env`** → 环境变量。

运行时数据三层（见 [data/README.md](../../data/README.md)）：

| 层 | 路径 | ADR |
|----|------|-----|
| Workspace | `data/workspaces/` | ADR-0009 |
| Profile | `data/profiles/` | ADR-0010 |
| Knowledge | `data/knowledge/` | ADR-0006 |

开发命令（仓库根目录；`make` 已内置 `GOPROXY`）：

```bash
make test
make run     # :7777 /health
make tidy

# 跑一场真实 LLM 会议（需 deploy/.env 中配置 DEEPSEEK_API_KEY）
cp deploy/.env.example deploy/.env   # 填入 key
make meet TOPIC="REST API 是否应采用 GraphQL"
# 可选：MEET_FLAGS="-max-rounds 2 -participants architect:Architect:design,dev:Developer:backend"
```

会议产出在 `data/workspaces/{meeting_id}/`（rounds、minutes、artifacts）。

集成测试（Engine 端到端，stub LLM/Principal）：

```bash
go test ./apps/server/internal/engine/... -timeout 5m
```

### Discord Transport（v0.2）

Bot Token 写在 **`deploy/.env`**（勿提交 git）：

```bash
DISCORD_BOT_TOKEN=your-bot-token
# 可选：各 Participant 独立 Bot
DISCORD_BOT_TOKEN_DESIGNER=...
```

非敏感选项在 **`apps/server/configs/server.yaml`** → `transport.discord`（`locale`、`meet_mode`、预设默认值、`max_confirmation_cycles`、`participant_bots` 等）。

```bash
make run-discord    # 独立启动
make server-dev     # HTTP 热重载 + 自动拉起 Discord（源码变更时 rebuild 子进程 binary）
make stop-discord   # 清理孤儿进程
```

**完整指令与行为**见 [docs/adapters/discord-transport.md](../../docs/adapters/discord-transport.md)。摘要：

| 阶段 | 能力 |
|------|------|
| 身份 | `!rt principal bind/whoami/unbind` |
| 发起 | `新会议` → 主题 → 参会阵容 → **简报三步** → 预设 **1–6** / **J1–J5** / 自定义 |
| 确认关 | 批准/驳回、ItemNotes（`2: 意见`）、触顶三选一 |
| 运行中 | 暂停/恢复/终止、立即合成/强制共识 |
| 自由问答 | `提问 [participant] …`（Round 1 后） |
| 结束 | 各交付物短节选 + **最后一条**附 `获取纪要/草案/待决/结论` 提示 |

绑定数据：`data/transport/discord-principal.json`。每频道同时一场会；进度与 Brief 由主 Bot 推送，Participant 发言由对应 Bot 发出。

**Docker 部署**：见 [deploy/README.md](../../deploy/README.md)。

结构说明见 [ADR-0008](../../docs/architecture/ADR-0008-project-structure.md)。
