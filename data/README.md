# data/

RoundTable **运行时数据根目录**（大部分 gitignore，模板在 `_templates/` 入库）。

三层存储（参考 OpenClaw，Meeting-first 拆分）：

```
data/
├── _templates/              # ✅ 入库 — 首次 Ensure 时 seed
├── workspaces/{meeting_id}/ # Meeting 产出（ADR-0009）
├── profiles/                # Agent 身份（ADR-0010）
│   ├── participants/{id}/
│   ├── principals/{id}/
│   └── moderator/
└── knowledge/               # 长期记忆（ADR-0006）
    ├── participants/{id}/
    ├── principals/{id}/
    └── shared/              # 可选共享池
```

| 层 | ADR | 生命周期 |
|----|-----|----------|
| workspaces | ADR-0009 | 单次 Meeting |
| profiles | ADR-0010 | 长期，跨 Meeting |
| knowledge | ADR-0006 | 长期，默认按 owner 隔离 |

配置见 `apps/server/configs/server.yaml` 的 `workspace` / `profile` / `knowledge`。

Adapter：`internal/adapter/{workspace,profile,knowledge}`。
