# data/

RoundTable **运行时数据根目录**（运行时内容 gitignore，模板与场景在 `_templates/` 入库）。

三层存储（参考 OpenClaw，Meeting-first 拆分）：

```
data/
├── _templates/                    # ✅ 入库 — 首次 Ensure 时 seed
│   ├── workspaces/                # MEETING.md 等 bootstrap 模板
│   ├── profiles/                  # SOUL / AGENTS 模板
│   ├── knowledge/                 # memory 模板
│   └── scenarios/                 # 端到端测试场景（profile + README）
│       └── 3-round-debate/        # make meet-3round
├── workspaces/{meeting_id}/       # Meeting 产出（ADR-0009）
├── profiles/                      # Agent 身份（ADR-0010）
│   ├── participants/{id}/
│   ├── principals/{id}/
│   └── moderator/
└── knowledge/                     # 长期记忆（ADR-0006）
    ├── participants/{id}/
    ├── principals/{id}/
    └── shared/                    # 可选共享池
```

| 层 | ADR | 生命周期 |
|----|-----|----------|
| workspaces | ADR-0009 | 单次 Meeting |
| profiles | ADR-0010 | 长期，跨 Meeting |
| knowledge | ADR-0006 | 长期，默认按 owner 隔离 |

---

## Meeting Workspace 布局（v0.2 已实现）

根路径：`data/workspaces/{meeting_id}/`（`workspace.root`，默认 `./data/workspaces`，**gitignore**）。

```
{meeting_id}/
├── MEETING.md                     # 会议简报：主题、议程、参会者、Token 汇总
├── MINUTES.md                     # 结构化纪要（随 Round / 共识更新）
├── pre-meeting/
│   └── perspectives.md            # Round 0：各 Participant 独立视角
├── rounds/
│   ├── round-001.md               # 辩论轮发言（Round 1+）
│   └── round-00N.md
├── free-dialogue/
│   └── after-round-001.md         # Round 1 后 Q&A（可配置关闭）
├── moderator/
│   └── round-00N-summary.md       # 每轮辩论后 Moderator 提炼摘要
├── usage/
│   ├── summary.md                 # Token 用量汇总 + 每次 LLM 调用明细
│   └── tokens.jsonl               # 机器可读日志
├── confirmation/                  # confirmation_mode: required 时
│   └── brief.md
└── artifacts/
    └── minutes.md                 # 最终 Minutes 副本（Completed 时）
```

完整说明见 [docs/domain/workspace.md](../docs/domain/workspace.md)。

---

## 测试场景

| 场景 | 命令 | 说明 |
|------|------|------|
| 3 轮辩论 | `make meet-3round` | skeptic + pragmatist，含 Pre-meeting、Round 1 自由对话、Token 统计 |

场景定义：`data/_templates/scenarios/3-round-debate/`（README + Participant Profile）。

关闭 Round 1 后自由对话：`-max-free-dialogue-questions 0` 或 `server.yaml` 中 `free_dialogue_max_questions: 0`。

---

## 配置

`apps/server/configs/server.yaml`：

| 键 | 说明 | 默认 |
|----|------|------|
| `workspace.root` | Workspace 根目录 | `./data/workspaces` |
| `profile.root` / `profile.templates` | Participant 身份 | `./data/profiles` |
| `knowledge.root` | 长期记忆 | `./data/knowledge` |
| `meeting.max_rounds_per_segment` | 辩论轮上限（**不含** Round 0） | `5` |
| `meeting.free_dialogue_max_questions` | Round 1 后每人提问轮数；`0` 关闭 | `1` |

环境变量：`ROUND_TABLE_WORKSPACE_ROOT`、`ROUND_TABLE_FREE_DIALOGUE_MAX_QUESTIONS` 等。

本地跑会需 `apps/server/.env` 配置 `DEEPSEEK_API_KEY`（见 `.env.example`）。

---

## 代码位置

Adapter：`apps/server/internal/adapter/{workspace,profile,knowledge}`。

Engine 投影：`apps/server/internal/engine/run.go`（Event → 文件）、`token_usage.go`、`free_dialogue.go`。

CLI：`apps/server/cmd/meet` — `make meet` / `make meet-3round`。
