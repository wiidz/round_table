# Meeting State Machine

Meeting 生命周期状态及转换。

---

## 状态

```
Created → Preparing → Running → Paused → Consensus → Confirmation → Completed → Archived
                              ↑              │            │
                              └──────────────┴────────────┘
                                    Rejected / Vetoed
```

| 状态 | 说明 |
|------|------|
| **Created** | Meeting 已创建，尚未开始 |
| **Preparing** | 分配 Participant、加载 Knowledge |
| **Running** | 讨论进行中（Round 循环） |
| **Paused** | 暂停，不发起新 Round |
| **Consensus** | Participant 已达成一致，等待进入 Confirmation 或结束 |
| **Confirmation** | Principal 审阅 Confirmation Brief（`confirmation_mode: required`） |
| **Completed** | 最终结论与 Artifacts 已输出 |
| **Archived** | 只读归档 |

---

## 关键转换

| 从 | 到 | 触发 |
|----|-----|------|
| Running | Consensus | `ConsensusReached` |
| Consensus | Confirmation | `confirmation_mode: required` |
| Consensus | Completed | `confirmation_mode: skip` → `MeetingFinished` |
| Confirmation | Completed | Principal `ConfirmationApproved` → `MeetingFinished` |
| Confirmation | Running | Principal `ConfirmationRejected`（注入 Feedback，继续讨论） |
| Consensus | Running | Principal `ConsensusVetoed` |
| Running | Paused | `MeetingPaused` |
| Paused | Running | `MeetingResumed` |
| Completed | Archived | 归档操作 |

---

## Confirmation 循环

```
Consensus → Confirmation (cycle=1)
  → Rejected → Running → … → Consensus → Confirmation (cycle=2)
  → …
  → cycle > max_confirmation_cycles → Principal 选择 Force / Continue / Abort
```

详见 [confirmation.md](../domain/confirmation.md) 与 [ADR-0004](../architecture/ADR-0004-principal-confirmation.md)。
