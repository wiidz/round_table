# ADR-0002: Consensus 判定策略

**状态**: Accepted  
**日期**: 2026-06-26  
**关联**: [consensus.md](../domain/consensus.md), [moderator.md](../domain/moderator.md), [CONSTITUTION.md](../CONSTITUTION.md)

---

## 背景

Consensus 是 Meeting 级属性，表示集体是否在 Agenda 上达成 sufficient agreement。  
Domain 文档（D-C01 ~ D-C05）要求明确：默认策略、可配置性、少数意见处理、Principal 权限、与 Completed 的边界。

---

## 决策

### 1. 默认策略：No Objection + Moderator Decision 兜底

v0.1 默认采用 **两层判定**：

```
每 Round Summary 后
  → No Objection Check（无人明确反对 → ConsensusReached）
  → 若有人反对 → 进入 Next Round
  → 达到 Round 上限仍无 Consensus → Moderator Decision（僵局兜底，见 [ADR-0005](./ADR-0005-round-termination.md)，默认 `max_rounds_per_segment=5`）
```

**No Objection** 作为默认主策略，理由：

- 符合会议隐喻：无异议即通过，比 100% 显式赞成更自然
- 比 Supermajority 更易在 LLM Participant 场景下自动判定
- 保留反对权，比纯 Moderator 独裁更贴近「协作决策」

**Moderator Decision** 仅作为僵局兜底，不是常规路径。

### 2. 策略可插拔，Meeting 级可配置

Domain 层定义 `ConsensusStrategy` 接口：

```go
type ConsensusStrategy interface {
    Evaluate(ctx ConsensusContext) (ConsensusResult, error)
}
```

| 配置层级 | 说明 |
|----------|------|
| **系统默认** | `NoObjectionStrategy` |
| **Meeting 级** | 创建时可覆盖，如 `UnanimousStrategy`、`SupermajorityStrategy(0.8)` |
| **Agenda Item 级** | v0.1 不支持，后续扩展 |

v0.1 实现三种策略即可：

| 策略 | 行为 |
|------|------|
| `NoObjection` | 无 Participant 标记 `object` → 通过 |
| `Unanimous` | 所有 Participant 标记 `agree` → 通过 |
| `Supermajority` | `agree` 比例 ≥ 阈值 → 通过 |

`Vote` 与加权投票推迟至 v0.2（需显式 ballot 模型）。

### 3. Participant 表态模型

每 Round 结束时，Moderator 向各 Participant 收集结构化表态（非自由文本）：

| 值 | 含义 |
|----|------|
| `agree` | 明确同意 |
| `object` | 明确反对（须附理由） |
| `abstain` | 弃权，不计入反对 |

No Objection 判定：`object` 计数为 0 即通过。  
Supermajority 判定：`agree / (agree + object)` ≥ 阈值（abstain 不计入分母）。

### 4. 部分一致：记录 Dissenting Opinion

当 Supermajority 或 Moderator Decision 产生「通过但有反对」的结果时：

- Consensus **仍然成立**
- 反对者的 `object` 理由写入 Minutes 的 **Dissenting Opinions** 段
- `ConsensusReached` Event payload 包含 `dissent: []DissentingOpinion`

Minority 不被静默丢弃，但不阻塞决策。

### 5. Principal 权限

Principal 位于 Moderator 之上（见 [context_diagram.md](../flow/context_diagram.md)、[principal.md](../domain/principal.md)）。

| 权限 | v0.1 |
|------|------|
| **Veto** | Principal 可在 `ConsensusReached` 后、Completed 前否决，Meeting 回到 `Running` |
| **Force Consensus** | Principal 可强制宣布 Consensus，跳过后续 Round |
| **Override Strategy** | Principal 可临时切换 Meeting 的 ConsensusStrategy |

Principal 操作产生 Event：`ConsensusVetoed`、`ConsensusForced`（见 ADR-0003）。

### 6. Consensus 与 Completed 的边界

> **已由 [ADR-0004](./ADR-0004-principal-confirmation.md) 修订。** 本节 `completion_mode` 不再使用。

```
Running → ConsensusReached → [Confirmation 或 skip] → MeetingFinished → Completed
```

| 模式 | 行为 |
|------|------|
| **`confirmation_mode: skip`** | `ConsensusReached` 后直接 `MeetingFinished` |
| **`confirmation_mode: required`** | 进入 Confirmation，Principal 批准后才 `MeetingFinished` |

详见 ADR-0004。

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| 默认 Unanimous | LLM Participant 难以可靠表达「明确同意」，易陷入无限 Round |
| 默认纯 Moderator Decision | 违背 Consensus over Completion，Moderator 不应替代集体 |
| 默认 Vote | v0.1 缺少 ballot 与选项模型，过度设计 |
| Agenda Item 级策略 | v0.1 范围过大，Meeting 级足够 |

---

## 后果

### 正面

- 默认路径简单、可自动运行
- 策略可插拔，后续扩展 Vote 不破坏 Domain
- Dissenting Opinion 保留少数声音
- Principal 保留最终控制权

### 负面

- No Objection 可能「沉默通过」——依赖 Participant 必须显式 `object`，需在 Prompt 层强调
- Moderator Decision 兜底仍可能独断——应写入 Minutes 并标记 `resolved_by: moderator`

### 待实现

- [ ] `ConsensusStrategy` 接口与三种实现
- [ ] Participant 表态收集流程（Moderator 职责）
- [ ] `ConsensusReached` Event payload 定义
- [ ] Minutes 中 Dissenting Opinions 段结构

---

## 决议项对照

| 编号 | 决议 |
|------|------|
| D-C01 | 默认 No Objection + Moderator Decision 兜底 |
| D-C02 | Meeting 级可配置，系统默认 No Objection |
| D-C03 | 部分一致算 Consensus，少数意见写入 Dissenting Opinions |
| D-C04 | Principal 拥有 Veto 与 Force 权限 |
| D-C05 | 由 ADR-0004 `confirmation_mode` 取代 completion_mode |
