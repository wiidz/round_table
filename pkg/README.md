# pkg/

将来可对外复用的公共库（Meeting Engine SDK、纯 Domain 类型导出等）。

v0.1 保持为空：领域代码在 `apps/server/internal/domain`，待 API 稳定后再提升到 `pkg/`（可独立 go.mod，由 `go.work` 引用）。
