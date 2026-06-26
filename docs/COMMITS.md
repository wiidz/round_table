# Git Commit 规范

RoundTable 采用**结构化 Commit Message**，记录的不只是「改了什么」，还有**为什么改、做了哪些设计决策**。

> 架构项目的 Commit 是设计史的一部分，应与 ADR、Domain 文档同等严肃。

---

## 两个维度

Commit 需要同时回答两个问题：

| 维度 | 问什么 | 写在哪 |
|------|--------|--------|
| **Module（工程模块）** | 改的是哪个代码库 / 工程？ | 标题 `scope` + 正文 **Module** |
| **Affected Components（领域组件）** | 影响了 Meeting Engine 的哪些概念？ | 正文 **Affected Components** |

示例：在 `server` 里实现 Confirmation 状态机 → Module 是 `server`，Affected Components 含 `Confirmation`、`Meeting`。

纯文档、纯 infra 改动若无领域语义变化，Affected Components 填 `None`。

---

## 格式

```text
<type>(<module>): <一句话描述>

Context

为什么需要这次修改？

当前背景是什么？

Decision

本次做出的设计决策是什么？

为什么选择这种方案？

Implementation

1.
2.
3.

Module

server

（跨模块时每行一个）

Affected Components

Meeting

Moderator

...

Files

- relative/path/file.go
  做了什么

References

ADR-000X（如有）

Breaking Changes

（没有可省略）
```

---

## 标题行

```
<type>(<module>): <一句话描述>
```

| 规则 | 说明 |
|------|------|
| **type** | 必填，见下表 |
| **module** | 必填，**工程模块**，见下表；跨模块时选**主模块**，其余写在正文 Module |
| **描述** | 中文或英文均可；祈使语气；≤ 72 字符；句末不加句号 |

### Type

| Type | 用途 |
|------|------|
| `feat` | 新功能、新领域能力 |
| `fix` | 缺陷修复 |
| `docs` | 仅文档变更 |
| `adr` | 新增或修订 ADR |
| `refactor` | 重构，不改变外部行为 |
| `test` | 测试 |
| `chore` | 构建、工具、杂项 |
| `deps` | 依赖变更 |

### Module（标题 scope）

**scope = 工程模块**，不是领域概念名。领域概念写在 **Affected Components**。

| Module | 说明 | 典型路径 |
|--------|------|----------|
| `server` | Go 后端：Meeting Engine、API、WebSocket、Scheduler | `cmd/` `internal/` `pkg/` |
| `web` | React 前端 | `web/` `apps/web/` |
| `android` | Android 客户端（规划） | `apps/android/` |
| `ios` | iOS 客户端（规划） | `apps/ios/` |
| `docs` | 项目文档 | `docs/` |
| `infra` | 部署与构建 | `Dockerfile` `Makefile` `docker-compose*` `.github/` |
| `repo` | 仓库级配置 | `README.md` `.cursorrules` `.gitmessage` `AGENTS.md` |

#### docs 子模块（可选二级 scope）

文档改动范围明确时，可用 `docs/<子模块>`：

| Scope | 说明 |
|-------|------|
| `docs/domain` | `docs/domain/` 领域概念 |
| `docs/architecture` | `docs/architecture/` ADR |
| `docs/flow` | `docs/flow/` 状态机、流程图 |
| `docs` | 上述以外的文档（VISION、CONSTITUTION、NAMING…） |

#### server 子模块（可选，实现阶段使用）

| Scope | 说明 |
|-------|------|
| `server/engine` | Meeting Engine 核心 |
| `server/scheduler` | Moderator 调度 |
| `server/adapter` | Runtime / Model / Transport 适配 |

---

## 正文各节

### Context（必填）

说明**动机与背景**。

### Decision（架构/设计变更时必填）

说明**设计决策与取舍**。机械性格式化改动可写「无设计决策」。

### Implementation（必填）

numbered 列表，概括主要步骤。

### Module（必填）

声明本次改动涉及的**工程模块**，每行一个。

- 单模块：与标题 scope 相同，写一行即可  
- 跨模块：标题用**主模块**，此处列全，例如 server + web 联动时两行都写  

```
Module

server
web
```

### Affected Components（必填）

受影响的**领域组件**（[NAMING.md](./NAMING.md) 语言），每行一个。无领域影响写 `None`：

```
Meeting
Principal
Moderator
Participant
Round
Consensus
Confirmation
Event
Minutes
Artifact
Knowledge
Scheduler
Engine
```

### Files（必填）

```text
- relative/path/file.go
  做了什么
```

### References / Breaking Changes

同前。

---

## 示例

### docs：新增 ADR

```text
adr(docs/architecture): 定义 Principal Confirmation 确认关

Context

Consensus 只表示 Participant 内部一致，无法表达 Principal 是否接受结论。

Decision

在 Consensus 与 Completed 之间插入 Confirmation 状态。
confirmation_mode: skip | required，默认 required。

Implementation

1. 起草 ADR-0004
2. 更新 state_machine 与 domain/confirmation.md

Module

docs

Affected Components

Meeting
Principal
Moderator
Confirmation
Event

Files

- docs/architecture/ADR-0004-principal-confirmation.md
  新增确认关架构决策

References

ADR-0004
```

### server：实现功能（未来）

```text
feat(server/engine): 实现 ConsensusReached Event 归约

Context

Meeting State 需由 Event fold 推导，Consensus 是首个必实现归约节点。

Decision

Apply 函数纯函数化；非法转换在 apply 层拒绝写入。

Implementation

1. 定义 ConsensusReached payload
2. 实现 Apply 至 MeetingState.Consensus
3. 单测覆盖合法/非法转换

Module

server

Affected Components

Meeting
Consensus
Event

Files

- internal/meeting/apply.go
  ConsensusReached 归约逻辑

- internal/meeting/apply_test.go
  归约单测

References

ADR-0002
ADR-0003
```

### infra：Docker

```text
chore(infra): 添加 server 多阶段构建 Dockerfile

Context

准备本地与 CI 统一构建环境。

Decision

采用 Go 1.25 多阶段构建；运行时 distroless 镜像。

Implementation

1. 新增 Dockerfile
2. docker-compose 增加 server 服务

Module

infra

Affected Components

None

Files

- Dockerfile
  server 多阶段构建

- docker-compose.yml
  本地编排
```

### 跨模块

```text
feat(server): Confirmation API 与 web 审阅页联调

Module

server
web

Affected Components

Confirmation
Principal
Meeting

Files

- internal/api/confirmation.go
  Confirmation 批准/驳回 HTTP 接口

- web/src/pages/Confirmation.tsx
  Principal 确认清单 UI
```

---

## 本地配置

```bash
git config commit.template .gitmessage
```

---

## 关联

- [CONSTITUTION.md](./CONSTITUTION.md)
- [NAMING.md](./NAMING.md)
- [architecture/README.md](./architecture/README.md)
