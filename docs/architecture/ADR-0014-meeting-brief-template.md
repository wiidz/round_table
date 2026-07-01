# ADR-0014: Meeting Brief Template（会议简报模板）

**状态**: Accepted  
**日期**: 2026-06-30  
**关联**: [ADR-0009-meeting-workspace.md](./ADR-0009-meeting-workspace.md), [ADR-0010-agent-profiles.md](./ADR-0010-agent-profiles.md), [ADR-0011-meeting-mode.md](./ADR-0011-meeting-mode.md), [workspace.md](../domain/workspace.md), [principal.md](../domain/principal.md)

---

## 背景

RoundTable 中 **Principal 委托 Meeting 的意图** 通过 **Meeting Brief（会议简报）** 表达：主题、目标、议程、讨论范围与完成标准。Discord Transport 已实现 Brief 三步向导（Goal → Agenda → Scope），写入 `MeetingCreated` 的 `Goal` / `Agenda`，并投影到 workspace 的 `MEETING.md`「会议简报 · Meeting Brief」段。

当前痛点：

1. **不可复用** — 每次开会从零填写；相似议题重复劳动。
2. **Brief 与合成强绑定** — 跳过 Brief 时 `Agenda` 为空，研讨型合成退化为 flat（见 [meeting-modes-deliberation-decision.md](../flow/meeting-modes-deliberation-decision.md) §P0）。
3. **与 Principal Profile 混淆** — `profiles/principals/{id}/USER.md` 描述「人」的偏好，不是「这次会」的任务书；二者职责不同，UI 上并列展示 Principal 档案易误导。
4. **Web 无 parity** — Brief 向导仅在 Discord；Web 侧缺少等价的可编辑、可复用入口。

需要一层 **跨 Meeting 复用、与单次 workspace 产出分离** 的 Brief 存储，且不引入新的 Domain 聚合根。

---

## 决策

### 1. 新增 Brief Template 层（非 Domain 概念）

**Meeting Brief Template** 是 Adapter 层的 **输入素材**，与 Profile（身份）、Knowledge（长期记忆）、Workspace（单次产出）并列，**不**成为 Constitution 级核心概念。

Meeting 仍通过 `MeetingCreated` 携带 `Topic`、`Goal`、`Agenda` 等字段；Template 仅在创建前被 Transport / Web 读取并填充 payload。

```
Principal 选模板 / 克隆历史 Brief
    → Transport 填充 meetLaunchConfig（或 Web 等价结构）
    → MeetingCreated Event
    → Engine State + workspace MEETING.md 投影
```

### 2. 物理路径

```
data/
├── _templates/briefs/              # ✅ 入库 — 项目内置模板
│   └── {template_id}/
│       └── BRIEF.yaml
└── briefs/                         # gitignore — 运行时 Principal 自建模板
    └── {template_id}/
        └── BRIEF.yaml
```

| 根 | 用途 |
|----|------|
| `_templates/briefs/` | 官方 / 场景 seed（类似 `_templates/profiles/`） |
| `briefs/` | Principal 保存的个人模板（可选 `owner` 字段或目录前缀 `{principal_id}/`） |

配置（v0.1 草案）：

```yaml
brief:
  root: ./data/briefs
  templates: ./data/_templates/briefs
```

环境变量：`ROUND_TABLE_BRIEF_ROOT`、`ROUND_TABLE_BRIEF_TEMPLATES`

Adapter 位置：`apps/server/internal/adapter/brief/`（`port.go` + `fs/`）

### 3. BRIEF.yaml  schema（v0.1）

与 Discord `meetBrief` + `meetLaunchConfig` 对齐，字段均为**可选默认值**；创建 Meeting 时 Principal 可覆盖。

```yaml
# BRIEF.yaml — Meeting Brief Template v0.1
meta:
  title: "游戏平衡评审"           # 模板展示名
  description: "裁决型，多轮 object 辩论"
  # owner: "discord:123"       # 可选，个人模板归属

topic: ""                         # 留空则创建时必填

brief:
  goal: "围绕指定 Topic 形成可执行共识"
  agenda:
    - "核心机制是否成立"
    - "数值与体验风险"
  in_scope: ""
  out_of_scope: "实施排期与人力"
  done_criteria: "每议题至少 1 条可执行结论"

meeting:
  mode: decision                  # decision | deliberation
  max_rounds: 3
  min_rounds_before_synthesis: 2  # deliberation only
  confirmation_mode: required       # required | skip
  free_dialogue_max_questions: 1
  participant_ids: []             # 空 = 全 roster
```

**映射到 Domain**（与现有 Discord 路径一致）：

| BRIEF.yaml | MeetingCreated / State |
|------------|------------------------|
| `topic` | `Topic` |
| `brief.goal` + scope 段 | `Goal`（经 `formatBriefForEngineGoal` 合并） |
| `brief.agenda[]` | `Agenda[]` |
| `meeting.mode` | `MeetingMode` |
| `meeting.*` 其余 | 对应 payload 字段 |

模板 **不含** Participant Profile（SOUL/AGENTS）内容；仅可指定 `participant_ids` 引用 roster。

### 4. 两种复用方式（均纳入本 ADR）

| 方式 | 来源 | 行为 |
|------|------|------|
| **A. 模板库** | `BRIEF.yaml` | 列表 → 选模板 → 向导预填 → 微调 → CreateMeeting |
| **B. 历史克隆** | 已完成 workspace 的 `MEETING.md` Brief 段 | 解析「会议主题 / 目标 / 议程 / 范围」→ 生成草稿 `meetLaunchConfig` 或新 `BRIEF.yaml` |

方式 B **不**另建存储：从 `data/workspaces/{meeting_id}/MEETING.md` 只读提取；Engine 已有 `parseBriefGoalFields` 可复用于逆向。

方式 A 为 **主路径**；方式 B 为 **快捷路径**（「上次开得不错，再来一版」）。

### 5. 与 Principal Profile 的分工（澄清）

| 层 | 文件 | 谁写 | 回答的问题 | 注入对象 |
|----|------|------|------------|----------|
| **Principal Profile** | `profiles/principals/{id}/USER.md` | 可选，极少改动 | 我是谁、怎么跟我协作（语言、验收习惯） | Moderator / Reception（ADR-0010；**待实现** Engine 注入） |
| **Meeting Brief Template** | `briefs/{id}/BRIEF.yaml` | 每次开会前选/改 | 这次要解决什么、议程与边界 | `MeetingCreated` → 全体 Participant 上下文 |
| **Meeting Brief 实例** | workspace `MEETING.md` | Engine 投影 | 本次会议已生效的简报 | Participant prompt bootstrap |

Principal **不需要** SOUL.md / AGENTS.md / TOOLS.md — Principal 不是 Agent。

Principal **需要写的 prompt 级内容**：

1. **高频**：Meeting Brief（目标、议程、范围）— 用 Template 降低重复劳动。
2. **低频**：`USER.md` 偏好段 — 仅当希望 Moderator 长期按固定方式服务时填写；未注入前可忽略。

### 6. Transport / Web 集成点

| 入口 | v0.1 行为 |
|------|-----------|
| Discord `!rt meet` / 自然语言开会 | 向导开始前可选「从模板加载」；步骤间可保存为个人模板 |
| Discord Reception | `start_meeting` 对齐 Brief 向导（现有 backlog） |
| Web | Brief 模板列表 + 创建 Meeting 表单（与 Discord parity） |
| `cmd/meet` / 场景 | 可读 `--brief-template={id}` 填充 CLI 参数 |

Domain / Engine **不** import `brief` adapter；仅 Transport 与 HTTP API 读取模板并组装 `MeetingCreated`。

### 7. API（HTTP，v0.1 草案）

```
GET  /api/brief-templates              # 列表（内置 + 个人）
GET  /api/brief-templates/{id}         # 详情（解析后 JSON）
POST /api/brief-templates              # Principal 保存个人模板
POST /api/meetings/clone-brief         # body: { meeting_id } → 返回 meetLaunchConfig 草稿
```

权限：个人模板按 Principal Identity 隔离；内置模板只读。

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| Brief 放进 Principal `USER.md` | 混淆「人」与「会」；一次 Meeting 一 Brief，不应绑死在 Profile |
| Brief 放进 Knowledge / MEMORY | Brief 是 intentional 输入，不是讨论沉淀的长期事实 |
| Brief 仅存在 workspace | workspace 按 Meeting 隔离且 Completed 后只读；无法跨 Meeting 复用 |
| 新增 Domain 聚合根 `Brief` | 过度建模；`MeetingCreated` 已承载运行时 Brief |
| 纯 Markdown 无 schema | 难以稳定映射 `meeting_mode`、轮次等配置；YAML + 可选 human README |
| 在 Engine 内硬编码模板列表 | 违反 Generic Solutions；模板应文件化、可扩展 |

---

## 后果

### 正面

- 降低重复填写成本，提高 Agenda 完整率，改善 deliberation 合成质量。
- 产品叙事清晰：**Brief 模板** = 开会意图；**Principal 档案** = 个人偏好（次要）。
- 与现有 `meetBrief` / `MEETING.md` 投影一致，Engine 改动最小。

### 负面

- 多一层存储与 API；需维护 YAML schema 版本。
- 历史克隆依赖 `MEETING.md` 格式稳定；格式变更需同步解析器。

### 待实现

- [ ] `adapter/brief` FS store + `List` / `Read` / `Write`
- [ ] `_templates/briefs/` 至少 1 个 seed 模板
- [ ] Discord 向导：选模板 / 存为模板
- [ ] Web：Brief 模板页 + 创建 Meeting 预填
- [ ] `POST /api/meetings/clone-brief` + `MEETING.md` 逆向解析
- [ ] `data/README.md` 补充 brief 层说明
- [ ] ADR-0010 待办：Principal `USER.md` Engine 注入（与 Brief 独立）

---

## 决议项

| 编号 | 决议 |
|------|------|
| D-BR01 | Meeting Brief Template 为 **Adapter 输入层**，非 Domain 聚合根 |
| D-BR02 | 标准文件 **`BRIEF.yaml`**，字段对齐 `meetBrief` + `meetLaunchConfig` |
| D-BR03 | 内置模板 `_templates/briefs/`；个人模板 `data/briefs/` |
| D-BR04 | 复用：**模板库（主）** + **历史 Meeting 克隆（辅）** |
| D-BR05 | 与 Principal Profile（`USER.md`）**职责分离**，不在 Profile 内嵌 Brief |

---

## 关联

- [profile.md](../domain/profile.md) — Profile 三层与 Principal USER.md
- [workspace.md](../domain/workspace.md) — `MEETING.md` Brief 段
- [confirmation.md](../domain/confirmation.md) — Confirmation Brief（终局验收，非本 ADR）
- 代码参考：`apps/server/internal/adapter/transport/discord/meet_brief.go`、`engine/meeting_doc.go`
