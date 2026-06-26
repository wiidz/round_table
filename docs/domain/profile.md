# Profile

Profile 是 **Agent 身份与运行手册**的文件层，跨 Meeting 复用。对应 OpenClaw 的 SOUL / AGENTS / USER / TOOLS，**不属于** Meeting workspace（ADR-0009）或 Knowledge（ADR-0006）。

---

## 已决议（ADR-0010）

### 路径

```
data/profiles/
├── participants/{id}/SOUL.md | AGENTS.md | TOOLS.md
├── principals/{id}/USER.md
└── moderator/AGENTS.md
```

模板（入库）：`data/_templates/profiles/`

### 与 Domain 的分工

| 层 | 职责 |
|----|------|
| **Domain** | `Role`、`Expertise`、`Goal`、Participant State |
| **Profile 文件** | 人格语气、行为细则、工具约定（Adapter 注入 prompt） |
| **Knowledge** | 长期事实与日志（MEMORY / memory/） |
| **Workspace** | 单次 Meeting 产出 |

### 加载时机

- `ParticipantInvited` → 读 Participant profile + Knowledge refs
- `Preparing` → 读 Principal USER.md

---

## OpenClaw 对照

| OpenClaw | RoundTable Profile |
|----------|-------------------|
| SOUL.md | `participants/{id}/SOUL.md` |
| AGENTS.md | `participants/{id}/AGENTS.md` |
| TOOLS.md | `participants/{id}/TOOLS.md` |
| USER.md | `principals/{id}/USER.md` |
| IDENTITY.md | 注册元数据（可选文件） |

---

## 关联

- [ADR-0010](../architecture/ADR-0010-agent-profiles.md)
- [participant.md](./participant.md)
- [principal.md](./principal.md)
- [knowledge.md](./knowledge.md)
- [workspace.md](./workspace.md)
