# ADR-0011: Meeting Mode（裁决型 vs 研讨型）

**状态**: Accepted  
**日期**: 2026-06-26  
**关联**: [meeting.md](../domain/meeting.md), [ADR-0002-consensus-strategy.md](./ADR-0002-consensus-strategy.md), [ADR-0005-round-termination.md](./ADR-0005-round-termination.md)

---

## 背景

RoundTable 默认流程面向 **裁决型（decision）** 议题：是否批准、是否上线，Participant 用 `agree/object` 表态，Moderator 做 No Objection 共识判定。

许多真实议题是 **研讨型（deliberation）**：设计游戏职业与技能、探索技术方案——目标是 **合成方案草案**，而非二元 OK/KO。

需要一种方式支持两种语义，且 **不复制整条 Meeting 编排**。

---

## 决策

### 1. 统一编排 + 模式化策略（非两套 Engine）

Meeting 生命周期、Round 调度、Free Dialogue、Workspace 投影、Token/流式输出 **共用同一 Engine 主循环**。

`meeting_mode` 在 `MeetingCreated` 配置，默认 `decision`。差异通过 **Policy 挂载点** 注入：

| 挂载点 | `decision` | `deliberation` |
|--------|------------|----------------|
| 辩论轮 Prompt | `Phase: debate`，要求 agree/object | `Phase: deliberation`，贡献设计点 |
| 辩论轮 Stance | LLM 返回 agree/object/abstain | 强制 `none`（不投票） |
| 轮间 Moderator | 分歧 vs 缓解措施 | 方案要素 + 冲突 + 开放问题 |
| Round 后终止 | ConsensusStrategy 判定 | **跳过共识**，达 `max_rounds` 后合成 |
| 终态事件 | `ConsensusReached` | `SynthesisCompleted` |
| 主 Artifact | minutes + 批准结论 | `artifacts/design-draft.md` + `open-questions.md` |
| Confirmation Brief | 「是否接受结论」 | 「草案是否足够进入下一环节」 |

### 2. Meeting Mode 枚举（v0.1）

| 值 | 说明 |
|----|------|
| `decision` | **默认**。二元决策，Consensus + 可选 Confirmation |
| `deliberation` | 方案共建，轮次结束后 Moderator 合成草案 |

v0.2 可扩展 `review`、`retrospective` 等，仍走同一编排。

### 3. 研讨型终止（v0.1 → v0.2）

```
每辩论轮结束
  → Moderator 轮间摘要
  → round >= min_rounds_before_synthesis → 合成就绪检测（DeliberationReadinessChecked）
       ├─ ready && round < max → 提前 SynthesisCompleted（resolved_by=readiness）
       ├─ ready && round == max → SynthesisCompleted（resolved_by=synthesis）
       └─ !ready && round < max → 下一轮
  → round >= max_rounds_per_segment && !ready → SynthesisCompleted（resolved_by=max_rounds）
```

**不**对 deliberation 运行 `ConsensusStrategy.Evaluate`。

#### v0.2 合成就绪检测

| 项 | 决策 |
|----|------|
| 挂载点 | `ModeratorSummarized` 之后、`startRound` / `completeDeliberation` 之前 |
| 最早轮次 | `MeetingCreated.min_rounds_before_synthesis`（默认 **2**） |
| 判定 | LLM 主路径（`Phase: deliberation-readiness`）；`Model == nil` 或失败 → **not ready** |
| 事件 | 每轮一条 `DeliberationReadinessChecked`（`ready`, `rationale`, `gaps[]`） |
| Workspace | `moderator/round-N-readiness.md` |

`resolved_by` 四态：

| 值 | 条件 |
|----|------|
| `readiness` | `round < max` 且就绪 |
| `synthesis` | `round == max` 且就绪 |
| `max_rounds` | 达上限仍未就绪 |
| `principal` | Principal `SynthesisForced` 于 Turn boundary |

### 4. 方案合成（v0.1.1）

达 `max_rounds` 后，Moderator 将全量会议记录合成为 `design-draft`：

```
completeDeliberation
  → synthesizeDeliberationFinal
       ├─ Engine.Model 可用 → LLM 合成（Phase: deliberation-synthesis）
       │     读 Pre-meeting / 各轮 transcript / Moderator 摘要 / 自由对话
       │     输出 JSON：core_scheme[]、decisions[]、open_questions[]
       │     → assembleDesignDraft → artifacts/design-draft.md
       │     API 失败或 JSON 无效 → 回退规则合成
       └─ Model == nil（测试 stub）→ 规则合成 moderatorSynthesizeFinal
  → SynthesisCompleted（含 TokenUsage）
```

| 路径 | 用途 |
|------|------|
| **LLM 合成** | 生产默认；Moderator 读 `AGENTS.md`，流式输出 + token 记录 |
| **规则合成** | fallback；marker + 启发式，无 API 依赖（单测 / integration） |

`decisions` 仅含已收敛共识；含「留待讨论 / 待确认 / 未表态」的条目归入 `open_questions`。

#### 合成质量原则（禁止案例特化）

| 层级 | 允许 | 禁止 |
|------|------|------|
| **LLM** | Moderator prompt / schema 定义字段语义；读 `AGENTS.md` | 在 prompt 里写死某场景 Topic 的字段示例 |
| **代码后处理** | JSON 修复；`decisions` 与 `core_scheme` 文本去重（关键词重叠 / near-duplicate）；通用「待决」短语拆分 | 按 workspace 追加 domain marker；从 free-dialogue 抽句补已决；已决条数 quota |
| **规则 fallback** | 通用研讨用语 marker（`decisionLineMarkers` 等，见 `deliberation_synthesis.go`） | 为 `game-class-design` 等模板单独分支 |

**已决可为空**：若 LLM 已将共识写入 `core_scheme`，Executive Summary 中「已决要点」为空或占位文案是合法结果，不得用规则硬凑。

质量不足时的优先顺序：改 Moderator prompt / `AGENTS.md` → ADR 设计 → 用户确认后再考虑新的**通用**结构处理（非 marker 堆叠）。

### 5. SynthesisCompleted 事件

新增 `SynthesisCompleted`（与 `ConsensusReached` 并列），payload：

- `summary` — 方案草案正文
- `open_questions[]` — 未决事项
- `resolved_by` — `readiness` | `synthesis` | `max_rounds` | `principal`（Principal 强制合成）

状态机：`Running → StatusConsensus`（与 ConsensusReached 相同后续路径），便于复用 Confirmation / MeetingFinished。

`ConsensusState` 在 deliberation 下 `strategy=deliberation`，`resolved_by` 取自 payload。

### 6. Goal 默认值

| Mode | 默认 Goal |
|------|-----------|
| `decision` | 围绕 Topic 达成可执行共识 |
| `deliberation` | 围绕 Topic 形成可评审的方案草案 |

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| 独立 DeliberationEngine | Round/Free Dialogue/Workspace 大量重复 |
| deliberation 仍用 agree/object | 语义错位，LLM 易误当投票 |
| 新 Status `Synthesized` | v0.1 复用 `Consensus` 门闩，减少状态机分叉 |

---

## 实现范围（v0.1）

- [x] `MeetingCreated.meeting_mode`
- [x] Engine deliberation 分支 + Workspace 产物
- [x] LLM 合成主路径 + 规则 fallback（`deliberation_synthesis_llm.go`）
- [x] Workspace：`artifacts/design-draft.md`、`artifacts/open-questions.md`
- [x] `meet -mode deliberation` + `game-class-design` 场景模板
- [x] `SynthesisCompleted` 携带 `TokenUsage`
- [x] v0.2：合成就绪检测（`DeliberationReadinessChecked`、`min_rounds_before_synthesis`）
- [x] v0.2：Agenda 子项驱动合成结构

---

## v0.2 补充：Agenda 子项驱动合成（已实现）

当 `MeetingCreated.agenda[]` 非空时：

| 层级 | 行为 |
|------|------|
| **合成 prompt** | 列出 agenda_id + Title；要求 LLM 按子项输出 |
| **LLM JSON** | `sections[{agenda_id, summary, decisions, open_questions}]` + `cross_cutting` |
| **design-draft** | Executive Summary 按 Agenda 分节（替代单一「核心方案」） |
| **readiness** | 判定 gaps 时参考各 agenda 覆盖度 |
| **无 Agenda** | 保持 v0.1.1 扁平 `core_scheme` 路径（向后兼容） |
| **规则 fallback** | 有 Agenda 时按标题/token 重叠分节；无 Agenda 时扁平结构 |

CLI：`-agenda "id:Title,id2:Title2"`；`make meet-game-class` 带默认四议程。

---

## v0.2 补充：Principal Force Synthesis（已实现）

研讨型 Running 阶段，Principal 可在 **Participant 发言间隙**（Turn boundary）强制结束辩论并合成：

```
Turn boundary (CurrentRound >= 1)
  → principal.Port.RunningAction
  → force_synthesis → SynthesisForced → completeDeliberation(resolved_by=principal)
```

| 模式 | Principal 强制 | Event | resolved_by |
|------|----------------|-------|-------------|
| `decision` | Force Consensus | `ConsensusForced` | `principal` |
| `deliberation` | Force Synthesis | `SynthesisForced` + `SynthesisCompleted` | `principal` |

裁决型 `ConsensusForced` 与研讨型 `SynthesisForced` 互斥（apply 层校验）。

---

## 关联

- [consensus.md](../domain/consensus.md) — 仅 `decision` 模式适用
- [confirmation.md](../domain/confirmation.md) — Brief 模板按 mode 分支
