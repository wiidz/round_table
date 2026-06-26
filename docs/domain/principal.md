# Principal

Principal（委托人）是 Meeting 的**发起者与决策者**——提出议题、拥有最终验收权，但不参与专家讨论。

RoundTable 全栈统一使用 **Principal** 这一术语，从 Domain 到 Transport / Auth / UI。

---

## 为什么选 Principal

| 术语 | 问题 |
|------|------|
| **登录账号称谓** | 技术/UI 用语，不表达会议中的角色与决策归属；易与「系统账号」「终端账号」混淆 |
| **Client** | 暗示甲乙方关系，不适合内部决策场景 |
| **Owner** | 与代码所有权、资源 Owner 混淆 |
| **Admin** | 暗示系统管理权限，非业务角色 |

**Principal** 强调**决策归属**——Meeting 因 Principal 的问题而召开，结论须符合 Principal 的预期。与 Moderator（司仪）、Participant（专家）形成清晰三角。

### 曾考虑的备选

| 名称 | 优点 | 未采用原因 |
|------|------|------------|
| **Convener** | 会议隐喻直观（召集人） | 偏「发起动作」，决策归属语义弱于 Principal |
| **Sponsor** | 企业会议常见 | 暗示出资方，范围偏窄 |
| **Requester** | 提出需求 | 过于被动，无验收权语义 |
| **Stakeholder** | 利益相关 | 过于宽泛，不可操作 |

---

## 已定义

### 定义

Principal 是因某个 **Topic** 而委托 Meeting 的人。Principal 不提供领域专业意见（那不是 Participant 的事），也不调度讨论（那是 Moderator 的事）。Principal 的职责是：**定方向、做验收、保最终控制权**。

### 三角角色

| 角色 | 隐喻 | 核心问题 |
|------|------|----------|
| **Principal** | 委托人 | 我要解决什么问题？结论是否符合我的预期？ |
| **Moderator** | 司仪 | 谁先说、何时总结、何时进入下一环节？ |
| **Participant** | 与会专家 | 从专业角度怎么看？ |

```
Principal
    │  发起 Topic / Agenda，审阅 Confirmation
    ▼
Moderator
    │  调度发言，整理 Brief，检测 Consensus
    ▼
Participant × N
       提供专业意见，不参与调度
```

### 属性

| 属性 | 说明 |
|------|------|
| **Identity** | Principal 的唯一标识（Domain 内；外部账号如 OAuth、Discord ID 在 Adapter 层绑定为 Principal Identity） |
| **DisplayName** | 显示名称（可选，Transport 层提供） |
| **Preferences** | 默认 Meeting 配置（如 `confirmation_mode`） |

Principal 不是 Meeting 内的集合——**一个 Meeting 有且仅有一个 Principal**（创建者）。

### 权限

Principal 位于 Moderator 之上，拥有**最终控制权**：

| 类别 | 权限 | 触发 Event |
|------|------|------------|
| **发起** | 创建 Meeting，设定 Topic / Agenda | `MeetingCreated` |
| **Confirmation** | 批准 / 驳回 Confirmation Brief | `ConfirmationApproved` / `ConfirmationRejected` |
| **Confirmation** | 跳过确认关、达上限强制批准 | `ConfirmationSkipped` / `ConfirmationForced` |
| **Consensus** | 否决 Participant 共识，要求继续讨论 | `ConsensusVetoed` |
| **Consensus** | 强制宣布 Consensus | `ConsensusForced` |
| **生命周期** | 暂停 / 恢复 / 终止 Meeting | `MeetingPaused` / `MeetingResumed` / `MeetingFinished` |
| **配置** | 覆盖 ConsensusStrategy | 配置变更 Event（v0.2） |

Principal **不能**：

- 直接与 Participant 对话（须经 Moderator）
- 自封为 Participant 发言（角色边界）
- 替代 Moderator 调度发言顺序

### 通信模型

```
Principal → Moderator → Participant → Moderator → … → Principal（Confirmation）
```

Participant 与 Principal 之间**无直连通道**。Principal 的 Feedback 由 Moderator 翻译为 Participant 可执行的上下文。

### 与 Confirmation 的关系

Confirmation 是 Principal 的专属环节——Moderator 整理清单，**Principal 审阅**。详见 [confirmation.md](./confirmation.md)。

| 概念 | 谁参与 |
|------|--------|
| Consensus | Participant 之间 |
| Confirmation | Principal + Moderator |

---

## 命名约定

| 层 | 术语 | 说明 |
|----|------|------|
| **Domain** | Principal | Meeting Engine 核心概念 |
| **Event Actor** | `principal` | Event Envelope 的 `Actor` 字段值 |
| **Transport / Auth** | Principal | 登录身份、OAuth、Discord ID 等均绑定为 Principal Identity |
| **UI** | 可显示「你」或 Principal 名称 | 面向 Principal 的交互界面 |

---

## 待决策

| 编号 | 问题 | 选项 / 备注 |
|------|------|-------------|
| D-PR01 | 是否允许多 Principal 共同委托 | v0.1 单一 Principal；v0.2 可考虑 Co-principal |
| D-PR02 | Principal 缺席时的默认行为 | Meeting 无法进入 Confirmation / 自动 skip |
| D-PR03 | Principal 可否中途转让 | 如转交同事继续验收 |
| D-PR04 | 匿名 Principal | 是否需要 Domain 级匿名（通常 Transport 层处理） |

---

## 关联

- 父索引：[README.md](./README.md)
- 验收环节：[confirmation.md](./confirmation.md)
- 调度者：[moderator.md](./moderator.md)
- 权威原则：[CONSTITUTION.md](../CONSTITUTION.md) § Core Concepts — Principal
- 事件 Actor：[event.md](./event.md)
