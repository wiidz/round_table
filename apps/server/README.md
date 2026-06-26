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

配置分层：`configs/server.yaml` → `.env` → 环境变量。

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

# 跑一场真实 LLM 会议（需 apps/server/.env 中配置 DEEPSEEK_API_KEY）
cp apps/server/.env.example apps/server/.env   # 填入 key
make meet TOPIC="REST API 是否应采用 GraphQL"
# 可选：MEET_FLAGS="-max-rounds 2 -participants architect:Architect:design,dev:Developer:backend"
```

会议产出在 `data/workspaces/{meeting_id}/`（rounds、minutes、artifacts）。

### Discord Transport（v0.2 切片）

Bot Token 写在 **`apps/server/.env`**（勿提交 git）：

```bash
DISCORD_BOT_TOKEN=your-bot-token
```

非敏感选项在 **`apps/server/configs/server.yaml`** → `transport.discord`（`allow_dm` / `allow_guild` / `guild_id`）。

在 [Discord Developer Portal](https://discord.com/developers/applications) 创建 Bot 并开启 **Message Content Intent**，邀请 Bot 进服务器后：

```bash
make run-discord
```

当前行为：文本指令绑定 Principal（每个服务器或私信会话一位）。

```
!rt help
!rt principal bind    # 绑定你自己为 Principal
!rt principal whoami  # 查看绑定
!rt principal unbind  # 解除绑定
```

绑定数据持久化在 `data/transport/discord-principal.json`（路径可在 server.yaml 配置）。

```
!rt meet 设计影舞者核心技能          # 默认 deliberation（见 server.yaml）
!rt meet -mode decision 是否上线    # 裁决型
```

仅已绑定的 Principal 可发起；每个 Discord 频道同时只允许一场会议。进度里程碑会推送到频道。

结构说明见 [ADR-0008](../../docs/architecture/ADR-0008-project-structure.md)。
