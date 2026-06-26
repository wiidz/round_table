# adapter/knowledge/

跨 Meeting 长期 Knowledge（ADR-0006）。默认 **按 owner 隔离**，可选 `shared/` 共享池。

| Scope | 路径 |
|-------|------|
| `participants` | `knowledge/participants/{id}/` |
| `principals` | `knowledge/principals/{id}/` |
| `shared` | `knowledge/shared/` |

文件：`MEMORY.md` + `memory/YYYY-MM-DD.md`

模板：`data/_templates/knowledge/`
