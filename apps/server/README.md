# apps/server

RoundTable **Meeting Engine** 后端（Go）。

Monorepo 内所有 Go 代码集中在此目录，避免与 `apps/web` 等客户端混放。

```
apps/server/
├── cmd/
│   ├── roundtable/     # HTTP/WS 服务入口
│   ├── meet/           # 本地跑一场会议（DeepSeek LLM）
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

结构说明见 [ADR-0008](../../docs/architecture/ADR-0008-project-structure.md)。
