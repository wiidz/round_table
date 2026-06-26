# ADR-0010: Agent Profile（身份与运行手册）

**状态**: Accepted  
**日期**: 2026-06-26  
**关联**: [ADR-0009-meeting-workspace.md](./ADR-0009-meeting-workspace.md), [ADR-0006-knowledge-scope.md](./ADR-0006-knowledge-scope.md)

**参考**: [OpenClaw Workspace file map](https://docs.openclaw.ai/concepts/agent-workspace)

---

## 背景

OpenClaw 用 `SOUL.md`、`AGENTS.md`、`USER.md` 等文件定义 Agent 身份与行为。ADR-0009 明确这些**不属于 Meeting workspace**，但未规定物理位置。

RoundTable 需要 **跨 Meeting 复用的 Participant / Principal 身份层**，与讨论产出（workspace）和长期记忆（knowledge）分离。

---

## 决策

### 1. Profile 根目录

`data/profiles/`（gitignore 运行时实例；`data/_templates/profiles/` 入库模板）

### 2. 目录布局

```
data/profiles/
├── participants/{participant_id}/
│   ├── SOUL.md       # 人格、语气、边界（OpenClaw SOUL.md）
│   ├── AGENTS.md     # Meeting 内行为规则（OpenClaw AGENTS.md）
│   └── TOOLS.md      # 工具与环境约定（OpenClaw TOOLS.md）
├── principals/{principal_id}/
│   └── USER.md       # Principal 偏好（OpenClaw USER.md）
└── moderator/
    └── AGENTS.md     # v0.2 LLM Moderator；v0.1 可选
```

**不放** `IDENTITY.md` 为必需文件——名称/emoji 进注册元数据；可选后续扩展。

### 3. OpenClaw 映射

| OpenClaw | Profile 路径 |
|----------|--------------|
| SOUL.md | `profiles/participants/{id}/SOUL.md` |
| AGENTS.md | `profiles/participants/{id}/AGENTS.md` |
| TOOLS.md | `profiles/participants/{id}/TOOLS.md` |
| USER.md | `profiles/principals/{id}/USER.md` |
| Moderator 规则 | `profiles/moderator/AGENTS.md` + Domain ADR |
| MEMORY.md / memory/ | **Knowledge**（ADR-0006），非 Profile |
| Meeting 产出 | **Workspace**（ADR-0009） |

### 4. 加载时机

| 事件 | 加载 |
|------|------|
| `ParticipantInvited` | SOUL + AGENTS + TOOLS + Knowledge refs |
| Meeting `Preparing` | Principal USER.md |
| Moderator LLM 模式 | moderator/AGENTS.md |

Domain State 仍持 `Role` / `Expertise` / `Goal`；Profile 文件是 **Adapter 层注入 material**，不替代 Domain 字段。

### 5. Adapter

```
apps/server/internal/adapter/profile/
├── port.go
└── fs/
```

`EnsureParticipant(id)` 从 `_templates` seed 缺失文件，不覆盖已有内容。

### 6. 配置

```yaml
profile:
  root: ./data/profiles
  templates: ./data/_templates/profiles
```

环境变量：`ROUND_TABLE_PROFILE_ROOT`

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| SOUL.md 放在 Meeting workspace | 每次 Meeting 复制，人格不一致 |
| SOUL 仅 DB 字段 | 不利于 OpenClaw Runtime 对接与人读编辑 |
| 单全局 profile | 多 Participant 无法隔离 |

---

## 后果

### 待实现

- [ ] Participant 注册 API → `EnsureParticipant`
- [ ] Engine：Invite 时加载 profile 拼上下文
- [ ] Transport：Principal 编辑 USER.md

---

## 决议项

| 编号 | 决议 |
|------|------|
| D-PF01 | Profile 与 Workspace、Knowledge **三层分离** |
| D-PF02 | Participant 标准文件：SOUL、AGENTS、TOOLS |
| D-PF03 | Principal 标准文件：USER |
| D-PF04 | 模板在 `data/_templates/profiles/`，Ensure 时 seed |
