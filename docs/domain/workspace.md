# Meeting Workspace

Workspace 是 **Meeting 的文件产出区**——Participant / Moderator 通过文件工具读写 Markdown 产出，Discussion 的结果落盘于此。

> 参考 [OpenClaw Agent Workspace](https://docs.openclaw.ai/concepts/agent-workspace)，但 RoundTable 是 **Meeting-first**：每个 Meeting 一个 workspace，不是 Agent 的全局 home。

---

## 与 OpenClaw 的对应

| OpenClaw | RoundTable | 说明 |
|----------|------------|------|
| `~/.openclaw/workspace`（Agent home） | `data/workspaces/{meeting_id}/` | **按 Meeting 隔离** |
| `AGENTS.md`（运行规则） | [Profile](./profile.md) `participants/{id}/AGENTS.md` | 不进 workspace |
| `SOUL.md`（人格） | [Profile](./profile.md) `participants/{id}/SOUL.md` | Domain 仍持 Role/Expertise |
| `USER.md`（用户画像） | [Profile](./profile.md) `principals/{id}/USER.md` | 非 MEETING.md |
| `memory/YYYY-MM-DD.md`（日志） | [Knowledge](./knowledge.md) `memory/` + workspace `rounds/` | 长期 vs 本轮 |
| `MEMORY.md` 长期记忆 | [Knowledge](./knowledge.md) `data/knowledge/` | 默认按 owner 隔离 |
| `~/.openclaw/`（config、secrets） | `configs/` + `.env` | 与 workspace **分离** |
| Session transcripts | Event Store | Source of truth 是 Event，不是文件 |
| `ArtifactProduced.ref` | `artifacts/*.md` | 领域 Event 指向 workspace 相对路径 |

OpenClaw 原则我们保留的三条：

1. **Workspace 与 config/secrets 分离** — secrets 只在 `.env`
2. **Markdown 为可读产出** — Minutes、Artifacts 以 `.md` 为主
3. **相对路径以 workspace 为 cwd** — Participant 文件工具只解析 Meeting 目录内路径

---

## 目录结构（运行时）

根路径由 `configs/server.yaml` 的 `workspace.root` 配置（默认 `./data/workspaces`），**gitignore**。

```
data/workspaces/{meeting_id}/
├── MEETING.md          # Topic、Agenda、Principal brief（MeetingCreated 时 seed）
├── MINUTES.md          # 结构化纪要（RoundCompleted / MeetingFinished 时更新）
├── action-items.md     # Action Items 清单
├── artifacts/          # ArtifactProduced.ref 指向此处
│   ├── proposal.md
│   └── architecture.md
└── rounds/
    └── round-001.md    # 每轮 Moderator Summary（可选分文件）
```

### 文件职责

| 文件 | 写入者 | 读取者 | 说明 |
|------|--------|--------|------|
| `MEETING.md` | Engine（bootstrap） | Participant、Principal | 讨论上下文，类似 OpenClaw bootstrap inject |
| `MINUTES.md` | Engine（Event 策展） | Principal、Transport | 不是 chat history |
| `artifacts/*` | Participant / Moderator | Principal | `ArtifactProduced` 的物理文件 |
| `action-items.md` | Engine | Principal | `ActionItemAdded` 投影 |
| `rounds/*` | Engine | Participant（下轮上下文） | Round Summary 可选分文件 |

---

## 与领域模型

```
Meeting State（Event Sourcing）
    │ ArtifactProduced { ref: "artifacts/proposal.md" }
    ▼
Workspace（文件系统投影，可重建）
```

- **Event Store** = source of truth（何时、谁、产生了什么）
- **Workspace** = 给人和 Agent 读的 **materialized view**
- Meeting 结束后 workspace 只读；Completed 后可归档或私有 git backup（参考 OpenClaw）

---

## 代码位置

| 层 | 路径 |
|----|------|
| Port 定义 | `apps/server/internal/adapter/workspace/` |
| FS 实现 | `apps/server/internal/adapter/workspace/fs/` |
| 调用方 | `engine`（bootstrap、Event 投影）、`participant` adapter（读上下文、写 artifact） |

Domain **不得** import workspace adapter。

---

## 关联

- 架构决策：[ADR-0009](../architecture/ADR-0009-meeting-workspace.md)（Accepted）
- 产出物：[meeting.md](./meeting.md) — Minutes、Artifact
- Event：[event.md](./event.md) — `ArtifactProduced`
- 工程结构：[ADR-0008](../architecture/ADR-0008-project-structure.md)
