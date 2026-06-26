# Knowledge

Knowledge 是 **跨 Meeting 的持久知识**（对外不叫 Memory，见 [NAMING.md](../NAMING.md)）。

OpenClaw 的 `MEMORY.md` + `memory/YYYY-MM-DD.md` 对应本概念；物理存储见 ADR-0006。

---

## 已决议（ADR-0006）

### 作用域

| 作用域 | 路径 | 说明 |
|--------|------|------|
| Participant | `data/knowledge/participants/{id}/` | **默认隔离** |
| Principal | `data/knowledge/principals/{id}/` | **默认隔离** |
| Shared | `data/knowledge/shared/` | 显式共享，Meeting 通过 `KnowledgeRef` 引用 |

### 文件布局

```
MEMORY.md              # 策展长期事实、偏好、决策
memory/YYYY-MM-DD.md   # 按日 append 日志
```

### 与 Meeting 的关系

- Meeting Event 流 **不包含** Knowledge 正文（ADR-0003 D-E06）
- Meeting State 持 `KnowledgeRef[]`，Engine 按需加载
- Meeting workspace 的 Minutes/Artifacts **不是** Knowledge

### 加载建议（Participant）

1. 自身 Participant scope 的 MEMORY + 近 1–2 日 memory/
2. Meeting 引用的 shared / principal refs
3. 当前 Meeting workspace 的 MEETING.md + rounds/（短期上下文，非 Knowledge）

---

## 关联

- [ADR-0006](../architecture/ADR-0006-knowledge-scope.md)
- [participant.md](./participant.md) — D-P01 ✅
- [meeting.md](./meeting.md) — KnowledgeRef
- [workspace.md](./workspace.md) — 勿混淆产出与记忆
- [profile.md](./profile.md) — SOUL/AGENTS 非 Knowledge
