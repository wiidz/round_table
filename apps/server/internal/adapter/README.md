# internal/adapter/

端口适配层，实现 Domain/Engine 定义的 interface。

| 包 | 职责 | 状态 |
|----|------|------|
| `storage/` | EventStore（memory → sqlite） | memory ✅ |
| `participant/` | ParticipantPort（stub → LLM） | stub ✅ |
| `model/` | Model Provider Adapter | 规划 |
| `runtime/` | Agent Runtime Adapter | 规划 |
| `transport/` | HTTP / WebSocket | 规划 Phase 4 |

Domain 层不得 import 本目录。
