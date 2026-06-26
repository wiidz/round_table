# apps/

Monorepo 可部署应用。各 app 独立构建，共享 `docs/` 领域与 ADR。

| 目录 | 说明 | 状态 |
|------|------|------|
| `server/` | Go Meeting Engine（HTTP/WS、Event Store、Engine） | Phase 1 |
| `web/` | React + Vite（Principal UI） | 规划 Phase 6 |
| `android/` | Android 客户端 | 规划 |
| `ios/` | iOS 客户端 | 规划 |

**Go 代码仅存在于 `server/`**。Web / Mobile 通过 `server` 的 Transport API 对接，业务逻辑不进 apps 客户端。
