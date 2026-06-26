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

## 目录结构（运行时 v0.2）

根路径由 `configs/server.yaml` 的 `workspace.root` 配置（默认 `./data/workspaces`），**gitignore**。

```
data/workspaces/{meeting_id}/
├── MEETING.md              # 会议简报（主题、议程、参会者、状态、Token 汇总）
├── MINUTES.md              # 结构化纪要（Engine 随 Event 更新）
├── pre-meeting/
│   └── perspectives.md     # Round 0：各 Participant 独立视角（互不可见直至汇总）
├── rounds/
│   └── round-NNN.md        # 辩论轮 N≥1 的发言与 stance
├── free-dialogue/
│   └── after-round-001.md  # Round 1 完成后固定一次的 Q&A（可配置关闭）
├── moderator/
│   └── round-NNN-summary.md # 辩论轮结束后 Moderator 提炼摘要
├── usage/
│   ├── summary.md          # Token 用量：汇总表 + 每次 LLM 调用
│   └── tokens.jsonl        # 同上，JSON Lines
├── confirmation/
│   └── brief.md            # confirmation_mode: required 时
├── action-items.md         # （预留）Action Items 投影
└── artifacts/
    └── minutes.md          # Meeting 完成时的 Minutes 副本
```

### 文件职责

| 文件 / 目录 | 写入者 | 读取者 | 说明 |
|-------------|--------|--------|------|
| `MEETING.md` | Engine | Participant、Principal | 讨论上下文 bootstrap；含议程与 Token 总量 |
| `MINUTES.md` | Engine | Principal、Transport | 策展纪要，非 chat history |
| `pre-meeting/` | Engine（Round 0 完成） | Participant（后续轮次 prompt 注入） | Pre-meeting 不计入 `max_rounds` |
| `rounds/` | Engine（RoundCompleted） | Participant（下轮上下文） | 辩论轮 1+ |
| `free-dialogue/` | Engine（FreeDialogueCompleted） | Participant | Round 1 后互相 Q&A |
| `moderator/` | Engine（ModeratorSummarized） | Participant | 轮间提炼，非全文复制 |
| `usage/` | Engine（含 TokenUsage 的 Event） | Principal、运维 | LLM 成本可观测性 |
| `artifacts/*` | Participant / Moderator | Principal | `ArtifactProduced.ref` 物理文件 |
| `confirmation/` | Engine | Principal | Confirmation Brief |

---

## 与领域模型

```
Meeting State（Event Sourcing）
    │ RoundCompleted / ModeratorSummarized / FreeDialogueCompleted / …
    │ ParticipantResponded { token_usage } → usage/
    ▼
Workspace（文件系统投影，可重建）
```

- **Event Store** = source of truth（何时、谁、产生了什么、用了多少 token）
- **Workspace** = 给人和 Agent 读的 **materialized view**
- Meeting 结束后 workspace 只读；Completed 后可归档或私有 git backup

---

## 代码位置

| 层 | 路径 |
|----|------|
| Port 定义 | `apps/server/internal/adapter/workspace/` |
| FS 实现 | `apps/server/internal/adapter/workspace/fs/` |
| 投影 | `apps/server/internal/engine/run.go`、`token_usage.go`、`free_dialogue.go` |
| 调用方 | `participant` adapter（读 MEETING.md / 上下文）、`cmd/meet` |

Domain **不得** import workspace adapter。

---

## 关联

- 架构决策：[ADR-0009](../architecture/ADR-0009-meeting-workspace.md)
- 数据目录：[data/README.md](../../data/README.md)
- 轮次：[round.md](./round.md) — Round 0、自由对话
- 产出物：[meeting.md](./meeting.md) — Minutes、Artifact
- Event：[event.md](./event.md) — `ArtifactProduced`、`TokenUsage`
