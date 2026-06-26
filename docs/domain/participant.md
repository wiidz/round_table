# Participant

Participant 是 Meeting 中的**领域专家**，只提供专业观点，不参与调度。

---

## 已定义

### 定义

Participant 代表某一领域的专家角色（如 Designer、Programmer、Architect）。只在 Moderator 邀请时响应，永不自调度，永不与其他 Participant 直接通信。

### 属性

| 层 | 属性 | 说明 |
|----|------|------|
| **身份** | Role | 角色名称与职责边界 |
| **身份** | Expertise | 专业领域描述 |
| **身份** | Goal | 当前 Meeting / Round 中的目标 |
| **观点** | Opinion | 当前立场，可随 Round 演化 |
| **运行时** | Status | 生命周期状态（见下） |
| **适配** | Capabilities | 能力声明（如可生成代码、可审阅设计） |
| **适配** | Model | 使用的模型（Model Adapter 层，Domain 仅持有抽象引用） |
| **适配** | Prompt | 角色 system prompt（实现细节，不泄漏进 Domain 逻辑） |

### 运行时状态

```
Idle → Waiting → Thinking → Speaking → Waiting → Done
```

详见 [participant flow](../flow/participant.md)。

### 设计约束

- Participant 不是 Sub Agent
- 不与 Principal 直接通信（见 [principal.md](./principal.md)）
- 各 Participant 守 Role 边界（Designer 不写代码，Programmer 不决定世界观）
- Memory / 持久知识在 Domain 层统一称为 **Knowledge**（见 [CONSTITUTION.md](../CONSTITUTION.md)）
- Participant 持有的 Knowledge 引用方式待决策（见下）

---

## 待决策

| 编号 | 问题 | 选项 / 备注 |
|------|------|-------------|
| D-P01 | Knowledge 作用域 | 全局共享 / 每 Participant 私有 / Meeting 级快照 |
| D-P02 | Opinion 是否持久化 | 仅当前 Round 有效 vs 写入 Event 供 Minutes 引用 |
| D-P03 | 同一 Role 是否允许多实例 | 如两个 Programmer 参与同一 Meeting |
| D-P04 | Capabilities 与 Role 关系 | Role 隐含 Capabilities vs 显式声明 |
| D-P05 | Model 是否 per-Participant | 不同专家用不同模型 vs Meeting 级统一 |
| D-P06 | Prompt 存放位置 | Domain 实体 vs Runtime Adapter 配置 |
| D-P07 | Done 是否可重新激活 | Round 内 Done 后是否可在下一轮回到 Waiting |

---

## 关联

- 父索引：[README.md](./README.md)
- 权威定义：[CONSTITUTION.md](../CONSTITUTION.md) § Core Concepts — Participant, Opinion
- 通信约束：[moderator.md](./moderator.md)
- 事件：[event.md](./event.md) — ParticipantInvited, ParticipantResponded
