# ADR-0005: Round 终止条件

**状态**: Accepted  
**日期**: 2026-06-26  
**关联**: [round.md](../domain/round.md), [ADR-0002-consensus-strategy.md](./ADR-0002-consensus-strategy.md), [ADR-0004-principal-confirmation.md](./ADR-0004-principal-confirmation.md)

---

## 背景

Round 是 Meeting 内的基本讨论单元。Domain 文档（D-R01 ~ D-R06）与 ADR-0002（僵局 Moderator Decision）、ADR-0004（`rounds_per_cycle`）均依赖明确的**终止条件**与**默认上限**。

缺少 ADR-0005 将导致 Scheduler 与 Engine 无法实现「何时结束本轮、何时结束讨论段」。

---

## 决策

### 1. 两层计数：Round 与 Running Segment

```
Meeting (Running)
  └── Running Segment（一次连续讨论段）
        └── Round 1 .. Round N
```

| 概念 | 说明 |
|------|------|
| **Round** | 一轮完整发言循环 + Summary + Consensus Check |
| **Running Segment** | 从进入 `Running` 到 `Consensus` 或 `Paused` / `Aborted`；每次 `ConfirmationRejected` 回到 `Running` 开启**新 Segment** |

每个 Running Segment 内 Round 编号从 1 重新计数。

### 2. 默认参数（MeetingCreated 可配置）

| 参数 | 默认 | 说明 |
|------|------|------|
| `max_rounds_per_segment` | **5** | 单个 Running Segment 内最多 Round 数 |
| `max_confirmation_cycles` | **3** | 见 ADR-0004 |

`rounds_per_cycle`（ADR-0004 用语）与 `max_rounds_per_segment` **同义**，统一使用 `max_rounds_per_segment`。

### 3. 单 Round 内发言规则（D-R02）

v0.1：**每个 Participant 每 Round 发言 exactly 一次**。

```
RoundStarted(order=[A,B,C])
  → A responds → B responds → C responds
  → RoundCompleted(summary)
  → Consensus Check
```

不支持 Round 内 rebuttal（反驳再发言）。需 rebuttal 时进入 **Next Round**。

### 4. Round 正常终止（D-R03）

每 Round 结束后按优先级判断：

```
1. Consensus Check 通过 → ConsensusReached，结束 Round 循环
2. 未通过 && round_number < max_rounds_per_segment → RoundStarted(round+1)
3. 未通过 && round_number >= max_rounds_per_segment → 僵局处理（§5）
```

Principal **Force Consensus**、**Veto**、**Pause**、**Abort** 可在任意 Round 间隙触发（见 §6），不受 Round 计数约束。

### 5. 僵局：达 Segment Round 上限（D-R03、D-R06）

`max_rounds_per_segment` 内未 Consensus 时：

1. Moderator 执行 **Moderator Decision** 兜底（ADR-0002）
2. 产生 `ConsensusReached`，payload `resolved_by: moderator`
3. Minutes 标记该 Consensus 为 **Moderator-resolved**
4. 进入 Confirmation 或 Completed（依 `confirmation_mode`）

**不**使用 Partial Consensus 状态；**不**无限延长 Segment。

### 6. Principal 与 Round 边界（D-R04）

v0.1 Principal **不可**在 Round 中途插入发言（不改变 Order）。

Principal 可在 **Participant 发言间隙**（Turn boundary）执行：

| 操作 | 效果 |
|------|------|
| `Pause` | 当前 Speaker 完成后暂停，不启动新 Round |
| `Force Consensus` | 立即 ConsensusReached，结束 Round 循环 |
| `Veto` | 仅在 Consensus 状态（Round 循环已结束） |
| `Abort` | MeetingFinished，outcome: aborted |

Round 内 reorder / 插话推迟至 v0.2。

### 7. Round 与 Agenda（D-R01）

v0.1：**不**建立 Round 与 Agenda Item 的一对一映射。

- 整个 Meeting 共享一个讨论上下文
- Agenda 作为 Meeting 级目标列表，Moderator Summary 可引用进度，但不驱动 Round 边界
- Agenda Item 级 Round 划分推迟 v0.2

### 8. Summary 与 Minutes（D-R05）

| 内容 | 写入 Minutes |
|------|--------------|
| 每 Round 的 Moderator Summary | ✅ 结构化字段 `rounds[].summary` |
| Participant 发言全文 | ✅ 来自 Event，可按 Round 索引 |
| 每发言即时摘要 | ❌ v0.1 不做 |

Minutes 由 Event 策展生成，Round Summary 是 Minutes 的章节之一，不是简单拼接 chat。

### 9. 与 Confirmation 驳回的联动

`ConfirmationRejected` → Meeting 回到 `Running` → **新 Running Segment**：

- `round_number` 重置为 1
- `max_rounds_per_segment` 仍适用（默认仍为 5，Meeting 级可单独调低）
- Principal Feedback 注入 Segment 上下文（ADR-0004）
- `confirmation_cycle` 递增，受 `max_confirmation_cycles` 约束

---

## 终止条件总览

```
                    ┌─────────────────┐
                    │  Round 进行中    │
                    └────────┬────────┘
                             │ RoundCompleted
                             ▼
                    ┌─────────────────┐
              ┌────│  Consensus Check │────┐
              │    └─────────────────┘    │
           通过│                      未通过│
              ▼                           ▼
      ConsensusReached          round < max?
      （结束 Round 循环）              │
                              ┌───────┴───────┐
                             是              否
                              ▼               ▼
                        Next Round    Moderator Decision
                                        → ConsensusReached
```

---

## 拒绝的选项

| 选项 | 原因 |
|------|------|
| 无 Round 上限 | 易无限循环，LLM 成本高 |
| Round 内多轮 rebuttal | v0.1 复杂度高；Next Round 可表达反对 |
| Partial Consensus 状态 | 与 ADR-0002 Moderator Decision 重复 |
| Round 1:1 Agenda Item | v0.1 Meeting 模型未稳定，过度结构化 |
| Principal Round 内插话 | 破坏 Order 模型；v0.1 用 Turn boundary 操作替代 |
| 全局 `max_rounds_total` 硬 cap | v0.1 由 Segment × Confirmation cycle 间接约束足够 |

---

## 后果

### 正面

- ADR-0002 僵局路径可落地
- ADR-0004 `rounds_per_cycle` 有明确定义
- Scheduler 实现边界清晰

### 负面

- 5 Round 可能对复杂议题偏少——Meeting 创建时可调
- 无 rebuttal 可能降低讨论深度——靠多 Round 弥补

### 待实现

- [ ] `max_rounds_per_segment` 写入 `MeetingCreated` payload
- [ ] Running Segment 计数与 round_number 重置逻辑
- [ ] 达上限 Moderator Decision 流程
- [ ] Minutes `rounds[].summary` 结构

---

## 决议项对照

| 编号 | 决议 |
|------|------|
| D-R01 | v0.1 不映射 Agenda Item；Agenda 为 Meeting 级上下文 |
| D-R02 | 每 Participant 每 Round 发言一次 |
| D-R03 | 优先 Consensus；达 `max_rounds_per_segment` 则 Moderator Decision |
| D-R04 | v0.1 Principal 仅 Turn boundary 操作，不插话 |
| D-R05 | Round Summary 结构化写入 Minutes |
| D-R06 | 达上限用 Moderator Decision，非 Partial Consensus |
