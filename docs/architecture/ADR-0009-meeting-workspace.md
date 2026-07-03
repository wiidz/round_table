# ADR-0009: Meeting Workspace（文件产出区）

**状态**: Accepted  
**日期**: 2026-06-26  
**关联**: [CONSTITUTION.md](../CONSTITUTION.md), [ADR-0003-event-model.md](./ADR-0003-event-model.md), [ADR-0008-project-structure.md](./ADR-0008-project-structure.md)

**参考**: [OpenClaw Agent Workspace](https://docs.openclaw.ai/concepts/agent-workspace)

---

## 背景

RoundTable 的 Meeting 必须产出 **Minutes、Artifacts、Action Items**。领域层已有 `ArtifactProduced` Event 与 `ArtifactRef`，但缺少**文件读写的基础设施**与目录规范。

OpenClaw 用 **Workspace** 作为 Agent 的唯一 cwd：Markdown 文件承载上下文与产出，与 config/secrets 分离。RoundTable 需要类似能力，但必须 **Meeting-first**——每个 Meeting 独立 workspace，而非全局 Agent home。

---

## 决策

### 1. 引入 Meeting Workspace

- **运行时根目录**: `{workspace.root}/{meeting_id}/`（默认 `./data/workspaces/`）
- **不入库**: 与 OpenClaw 一样，workspace 是运行时数据；可选手动 git backup
- **与 config 分离**: `configs/` + `.env` 管服务配置；workspace 管讨论产出

### 2. 标准文件布局（v0.2，已实现）

```
{meeting_id}/
├── MEETING.md
├── MINUTES.md
├── pre-meeting/perspectives.md      # Round 0
├── rounds/round-NNN.md              # 辩论轮 1+
├── free-dialogue/after-round-001.md # Round 1 后 Q&A（可关闭）
├── moderator/round-NNN-summary.md
├── usage/summary.md                 # Token 统计
├── usage/tokens.jsonl
├── confirmation/brief.md            # 可选
└── artifacts/
```

不照搬 OpenClaw 的 `SOUL.md` / `AGENTS.md` / `USER.md` 进 **Meeting workspace**——身份与长期记忆分别在 Profile（ADR-0010）与 Knowledge（ADR-0006）。

### 3. OpenClaw → RoundTable 映射

| OpenClaw | RoundTable |
|----------|------------|
| Agent workspace（单例） | Per-Meeting workspace |
| Bootstrap files inject | `MEETING.md` + 上轮 `rounds/` 作为 Participant 上下文 |
| `memory/*.md` append | Event Store append；`rounds/` 为可选投影 |
| `MEMORY.md` 长期记忆 | Knowledge（ADR-0006，跨 Meeting） |
| File tools cwd = workspace | `workspace.Port` 限定 Meeting 目录 |
| Secrets in env | 不变 |

### 4. 工程结构补充（修订 ADR-0008）

```
round_table/
├── data/workspaces/              # gitignore，运行时 Meeting 产出
└── apps/server/internal/adapter/
    └── workspace/                # Port + fs 实现
        ├── port.go
        └── fs/
```

Monorepo 根目录 `data/` 与 `apps/server` 并列：workspace 是**跨 Meeting 的运行时数据区**，不是 Go 源码。

### 5. Workspace Port（adapter 层）

```go
type Port interface {
    EnsureMeeting(meetingID, topic string) error
    Read(meetingID, relPath string) ([]byte, error)
    Write(meetingID, relPath string, data []byte) error
    List(meetingID, relPath string) ([]string, error)
    Resolve(meetingID, relPath string) (string, error) // abs path, jail to meeting dir
}
```

- Engine 在 `MeetingCreated` 后 `EnsureMeeting`，写入 `MEETING.md`
- `ArtifactProduced.ref` 必须是 **meeting 内相对路径**（如 `artifacts/proposal.md`）
- Participant adapter 读 workspace 拼上下文；写 artifact 经 Moderator 许可后 `Write`

### 6. 配置

```yaml
# apps/server/configs/server.yaml
workspace:
  root: ./data/workspaces
```

环境变量覆盖: `ROUND_TABLE_WORKSPACE_ROOT`

### 7. 安全边界

- 相对路径解析 **jail** 在 `{meeting_id}/` 内（禁止 `../` 逃逸）
- workspace 内 **禁止** 写入 secrets
- v0.1 不做 sandbox；未来可参考 OpenClaw `agents.defaults.sandbox`

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| 产出直接写 SQLite BLOB | 不利于 Principal / Agent 读 md；违背 OpenClaw 可读性 |
| 全局单一 workspace | 多 Meeting 混淆；违反 Meeting 隔离 |
| 照搬 OpenClaw `SOUL.md` 等 | Agent 人格应在 Participant 模型，非文件 |
| Workspace 放进 `internal/domain` | 文件 IO 是 adapter 职责 |
| 用 Viper/JSON 存 Minutes | Minutes 是人读 md；结构化部分在 Event State |

---

## 后果

### 已实现（v0.2）

- [x] `adapter/workspace/fs` 实现 + path jail 测试
- [x] Engine：`MeetingCreated` → `EnsureMeeting` + `MEETING.md`
- [x] Engine：`RoundCompleted` / `ModeratorSummarized` / `FreeDialogueCompleted` → 投影
- [x] Engine：`ParticipantResponded` 等携带 `TokenUsage` → `usage/`
- [x] Participant adapter：读 `MEETING.md` + Discussion context 作 prompt
- [x] Transport：`GET /api/meetings/{id}/archive` 打包下载 workspace（zip）；Web 会议详情页触发
- [x] Transport：`DELETE /api/meetings/{id}` 删除会议记录与 workspace

### 正影响

- Discussion 有明确物理产出，Principal 可审阅 md
- 与 OpenClaw Runtime 对接时有清晰文件契约
- Event 仍可重建 workspace（投影可丢可重建）
- Token 用量落盘 `usage/`，便于成本审计

---

## 决议项

| 编号 | 决议 |
|------|------|
| D-WS01 | 每 Meeting 一个 workspace 目录，根路径可配置 |
| D-WS02 | Markdown 为主产出格式；`ArtifactProduced.ref` 为相对路径 |
| D-WS03 | Workspace 在 `adapter/workspace`；Domain 不 import |
| D-WS04 | 参考 OpenClaw 分离原则，不照搬 Agent bootstrap 文件集 |
