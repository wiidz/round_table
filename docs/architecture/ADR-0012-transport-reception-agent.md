# ADR-0012: Transport Reception Agent（自然语义接待层）

**状态**: Accepted  
**日期**: 2026-06-28  
**关联**: [CONSTITUTION.md](../CONSTITUTION.md), [ADR-0007-moderator-scheduling.md](./ADR-0007-moderator-scheduling.md), OpenClaw `get-reply` 分层（规则 → LLM）

---

## 背景

Principal 希望通过 **自然语言** 完成专家查询、会议状态、纪要/草案获取，以及（后续）开会与专家管理，而不依赖 `!rt` 前缀与固定关键词。

Constitution 规定：

- **Moderator**（Engine 内）是会中调度器，不负责 Transport 入站解析。
- **Transport** 是 Principal 与系统之间的适配层，可引入 LLM。
- Engine / 合成 pipeline 保持 **领域无关**，禁止为单一 workspace 增加 heuristic。

OpenClaw 采用 **Hybrid**：控制命令短路 + 默认 LLM agent。RoundTable 采用同构分层，但 mutating 操作必须经 **确认 session**，且 tool 仅调用已有 Go service。

---

## 决策

### 1. 分层（Discord Host Bot）

```
入站消息
  → 安全 / 运行中干预 / 确认关 / 精确 artifact 触发
  → !rt 前缀命令
  → meet / expert setup 多轮回复
  → 规则 fast path（新会议、开个会…）
  → Reception Agent（LLM + 只读 tools）   ← 本 ADR
  → MisplacedInputHint / 静默
```

Reception **不替代** Engine Moderator；**不**在 synthesis / deliberation 内增加 NL heuristic。

### 2. R1 范围（只读 tools）

| Tool | 说明 | 底层 |
|------|------|------|
| `list_participants` | 专家名录 | `ParticipantAdmin` / roster |
| `meeting_status` | 频道输入态 / 会议 ID | `InputPhase` |
| `get_artifact` | 纪要 / 草案 / 待决 / 结论 | `MeetRunner.fetchArtifact` |
| `clarify` | 无法执行时追问 | LLM 生成文案 |

R2+：`start_meeting`、`create_participant` 等 **写操作** + 统一 confirm session。

### 2b. R2 范围（写操作 + 确认）

| Tool | 说明 | 底层 |
|------|------|------|
| `create_participant` | 新建专家 | `ParticipantAdmin.executeCreate` |
| `update_participant` | 编辑专家 | `ParticipantAdmin.executeUpdate` |
| `delete_participant` | 删除专家 | `ParticipantAdmin.executeDelete` |
| `update_participant_profile` | 写/编辑 SOUL、AGENTS、TOOLS | Principal 描述方向 → LLM 草稿 → 确认 → `Profile.WriteParticipant`；也可直接粘贴 Markdown |
| `start_meeting` | 自然语言开会（非「开个会」fast path） | `MeetRunner.launch` |

LLM 解析参数 → 展示 preview → Principal 回复 **1/0** → 执行。需 Principal bind。

### 3. LLM 调用

- 使用现有 `model.Port`（OpenAI-compatible），与 Engine 相同 provider 配置。
- 单次 completion：system + user（含 roster 快照、当前 phase）→ **JSON** `{ "tool", "artifact", "message" }`。
- `transport.discord.reception_agent_enabled`（默认 `true`）；无 API key 时 Reception 不挂载。

### 4. 确认与鉴权

- R1 只读，无需 confirm。
- R2 起：mutating tool 返回 preview，写入 setup session，Principal 回复 `1/0` 后执行。
- 写操作需 Principal bind（与开会一致）；R1 只读对频道内已绑定 Principal 开放。

### 5. 与规则 fast path 的关系

保留 `TryBeginNaturalMeet`、`isMeetStartTrigger` 等 **零 token** 路径。Reception 仅处理 **未命中** 规则的普通语句。

---

## 后果

- Discord Transport 增加可选 LLM 依赖（Host 接待）；Engine 会议内 LLM 不变。
- 需监控 token 与 latency；后续可加「会中高频轮次跳过 Reception」策略。
- Web / 其他 Transport 可复用同一 Reception 模块（非 Discord 专用 domain）。

---

## 参考

- `apps/server/internal/adapter/transport/discord/reception*.go`
- OpenClaw: `src/auto-reply/reply/get-reply-inline-actions.ts`, `runEmbeddedAgent`
