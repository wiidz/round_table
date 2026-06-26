# ADR-0008: Go 工程结构（Monorepo）

**状态**: Accepted  
**日期**: 2026-06-26（2026-06-26 修订：Monorepo / `apps/server`）  
**关联**: [CONSTITUTION.md](../CONSTITUTION.md), [ADR-0003-event-model.md](./ADR-0003-event-model.md)

**参考项目**: `crm_api`（Gin CRM 双端口 API）

---

## 背景

RoundTable 进入 Phase 0 代码化，需要确定 Go 仓库目录结构。  
团队已有 `crm_api` 作为 Go 项目参考，但其分层与 RoundTable 的 Meeting-first、stdlib-first 原则并不完全兼容。

仓库采用 **Monorepo**：Go 后端、Web、Mobile 均为 `apps/` 下的独立应用，避免根目录 Go 代码与前端混放。

---

## 决策

### 1. Monorepo 布局

```
round_table/
├── go.work                  # Go workspace，引用 apps/server
├── apps/
│   ├── server/              # 全部 Go 代码
│   │   ├── go.mod
│   │   ├── cmd/
│   │   ├── configs/
│   │   └── internal/
│   ├── web/                 # React（Phase 6）
│   ├── android/
│   └── ios/
├── pkg/                     # 将来跨 app 共享的 Go 库（v0.1 留空）
├── docs/
├── scripts/
└── Makefile                 # 委托 apps/server
```

**Module 路径**: `round_table/apps/server`  
**拒绝**: 根目录 `cmd/` + `internal/` — 与 Monorepo 语义冲突，Go 代码易与 `apps/web` 边界模糊。

### 2. 借鉴 crm_api 的部分

| 做法 | round_table 对应 |
|------|------------------|
| 薄 `cmd/` 多入口 | `apps/server/cmd/roundtable`、`cmd/migrate` |
| `internal/` 与入口分离 | 全部业务在 `apps/server/internal` |
| `configs/*.yaml` | `apps/server/configs/server.yaml` |
| `Makefile`（run/build/test/migrate） | 根 Makefile，路径指向 server |
| 优雅关闭 | `platform/server` signal + Shutdown |
| 第三方集成独立包 | `internal/adapter/{model,runtime,transport}` |

### 3. 不照搬 crm_api 的部分

| crm_api 做法 | 不采用原因 |
|--------------|------------|
| 全局 `repos.Xxx.Repo` | 违反无全局状态；难测试 |
| GORM entity = 领域模型 | DB 泄漏进 Domain；Event Sourcing 不用 ORM 建模 Meeting |
| Gin + 重依赖栈 | Constitution：stdlib 优先 |
| 包级 service 无 interface | Engine/Scheduler 需 mock |
| `internal/domain/console/todo` CRUD 形状 | RoundTable 是 Event 驱动状态机，不是 CRUD |
| 巨型单文件 router | Transport 按域拆分 Register |

### 4. apps/server 目录结构

```
apps/server/
├── cmd/
│   ├── roundtable/          # HTTP/WS 服务入口
│   └── migrate/             # SQLite 迁移 CLI
├── configs/                 # 运行时配置（无 secret 入库）
├── internal/
│   ├── domain/              # 纯领域：零框架、零 DB import
│   │   ├── meeting/         # MeetingState、Apply、Status
│   │   ├── event/           # Event 类型、Envelope
│   │   └── consensus/       # ConsensusStrategy
│   ├── engine/              # Meeting Engine 编排（调 scheduler + store）
│   ├── scheduler/           # Moderator 调度（ADR-0007 Fixed Order）
│   ├── adapter/             # 端口实现
│   │   ├── storage/         # EventStore（memory → sqlite）
│   │   ├── participant/     # ParticipantPort（stub → model）
│   │   ├── model/           # LLM Model Adapter（v0.2）
│   │   ├── runtime/         # Agent Runtime Adapter（v0.2）
│   │   └── transport/       # HTTP/WS handlers
│   └── platform/            # 配置、进程、日志
└── go.mod
```

### 5. 依赖方向

```
cmd → platform → engine → scheduler → domain
                    ↓
                 adapter → domain
transport → engine（仅 DTO 转换，无业务）
```

**禁止**：`domain` import `adapter`、`platform`、`transport`。

### 6. 与 CONSTITUTION 对齐

| Constitution | 结构体现 |
|--------------|----------|
| Meeting First | `apps/server/internal/domain/meeting` 为核心 |
| Runtime/Model/Transport 独立 | `internal/adapter/*` |
| stdlib 优先 | v0.1 HTTP 用 `net/http`，存储先 memory 再 sqlite |
| 接口优于实现 | `storage.Store`、`participant.Port`、`consensus.Strategy` |
| Monorepo | `apps/server` + `apps/web` + … |

### 7. apps/ 与 Transport

crm 用 client/console 双端口。RoundTable 用 **Transport adapter** 抽象：

- v0.1：单 `cmd/roundtable` + REST
- 未来：`apps/web`、mobile 均为 Transport 客户端，不增加 server 副本

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| 根目录 `cmd/` + `internal/` | Monorepo 下 Go 应收敛到 `apps/server` |
| 完全复制 crm `internal/domain/console/*` | CRM CRUD 形态不适合 Event Sourcing |
| 顶层无 `internal/domain`，直接用 GORM model | 领域不纯 |
| `internal/pkg/` 代替顶层 `pkg/` | 跨 app 复用需顶层 `pkg/` + 独立 go.mod |
| 单文件 `main.go` 包含 Engine | 无法测试 |

---

## 后果

### 待实现（Phase 1）

- [x] `domain/meeting` Apply 全 Event 类型
- [ ] `adapter/storage/sqlite`
- [ ] `engine` 主循环
- [ ] `transport/http` Meeting API

---

## 决议项

| 编号 | 决议 |
|------|------|
| D-ST01 | Monorepo：`apps/server` 承载全部 Go；`go.work` 在仓库根 |
| D-ST02 | 借鉴 crm Makefile/configs/多 cmd，不借鉴 repos 全局注册表 |
| D-ST03 | Domain 层禁止 import adapter/platform/transport |
