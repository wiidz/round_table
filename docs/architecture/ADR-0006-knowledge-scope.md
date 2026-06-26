# ADR-0006: Knowledge 作用域与存储

**状态**: Accepted  
**日期**: 2026-06-26  
**关联**: [CONSTITUTION.md](../CONSTITUTION.md), [ADR-0003-event-model.md](./ADR-0003-event-model.md), [ADR-0009-meeting-workspace.md](./ADR-0009-meeting-workspace.md), [ADR-0010-agent-profiles.md](./ADR-0010-agent-profiles.md)

**参考**: [OpenClaw Agent Workspace — MEMORY.md / memory/](https://docs.openclaw.ai/concepts/agent-workspace)

---

## 背景

RoundTable 需要 **跨 Meeting 持久知识**（Knowledge）。ADR-0003 已定 Knowledge 使用**独立 Event 流**，Meeting 仅持 `KnowledgeRef`。

OpenClaw 用 `MEMORY.md`（策展长期记忆）+ `memory/YYYY-MM-DD.md`（日誌）。RoundTable 需确定：**谁拥有 Knowledge、文件放哪、如何共享**。

---

## 决策

### 1. 默认隔离，可选共享

| 作用域 | 路径 | 默认 | 说明 |
|--------|------|------|------|
| **Participant** | `data/knowledge/participants/{id}/` | ✅ 默认 | 每个专家私有记忆 |
| **Principal** | `data/knowledge/principals/{id}/` | ✅ 默认 | 委托人私有记忆 |
| **Shared** | `data/knowledge/shared/` | 可选 | 显式共享池，Meeting 通过 `KnowledgeRef` 引用 |

**不采用**全局单一 Knowledge 目录作为默认——避免 Meeting / 用户间泄漏。

### 2. 标准文件布局

```
{knowledge.root}/{scope}/{owner_id}/
├── MEMORY.md           # 策展长期记忆（OpenClaw 对应）
└── memory/
    └── YYYY-MM-DD.md   # 按日 append 日志
```

Shared 池无 `owner_id`：`data/knowledge/shared/`。

### 3. KnowledgeRef（Meeting 引用）

Meeting State 持有引用列表，不 embed 内容：

```go
// 概念形状（Domain 后续细化）
type KnowledgeRef struct {
    Scope   string // participants | principals | shared
    OwnerID string // empty when scope=shared
}
```

Engine 在 `Preparing` / `ParticipantInvited` 时按 ref 加载；Participant **默认只读** 自身 scope + Meeting 显式 refs（含 shared）。

### 4. Event 流

- Knowledge 变更产生 **独立 Knowledge Event**（ADR-0003 D-E06），不写入 Meeting Event 流
- 文件系统是 Knowledge 的 **materialized view**（与 Meeting workspace 投影类似）
- v0.1：文件 append/write + 骨架；完整 Event 流 Phase 2

### 5. 工程结构

```
data/knowledge/
├── participants/{id}/
├── principals/{id}/
└── shared/

apps/server/internal/adapter/knowledge/
├── port.go
└── fs/
```

模板：`data/_templates/knowledge/`（入库 seed）

### 6. 配置

```yaml
knowledge:
  root: ./data/knowledge
  templates: ./data/_templates/knowledge
  shared_enabled: true
```

环境变量：`ROUND_TABLE_KNOWLEDGE_ROOT`

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| 默认全局共享 Knowledge | 隐私与 Meeting 隔离差 |
| Knowledge 写入 Meeting workspace | 生命周期不同；ADR-0009 仅产出 |
| Knowledge 仅向量库、无 md | 违背 OpenClaw 可读性；Principal 难审阅 |
| Knowledge embed 进 Meeting Event | ADR-0003 已拒绝 |

---

## 后果

### 待实现

- [ ] Knowledge Event 流 + Apply
- [ ] Engine：按 `KnowledgeRef` 加载上下文
- [ ] Participant adapter：读 MEMORY + 近 N 日 memory/
- [ ] 共享池 ACL（v0.2：谁可写 shared）

---

## 决议项

| 编号 | 决议 |
|------|------|
| D-K01 | Knowledge 默认按 Participant / Principal **隔离**存储 |
| D-K02 | `shared/` 为显式可选共享池 |
| D-K03 | 文件布局：`MEMORY.md` + `memory/YYYY-MM-DD.md` |
| D-K04 | Adapter 在 `internal/adapter/knowledge` |
