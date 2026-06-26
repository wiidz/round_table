# apps/server

RoundTable **Meeting Engine** 后端（Go）。

Monorepo 内所有 Go 代码集中在此目录，避免与 `apps/web` 等客户端混放。

```
apps/server/
├── cmd/
│   ├── roundtable/     # HTTP/WS 服务入口
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

仓库根目录 `go.work` 引用本 module。开发命令：

```bash
make test    # 从仓库根运行
make run
```

结构说明见 [ADR-0008](../../docs/architecture/ADR-0008-project-structure.md)。
