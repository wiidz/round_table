# internal/adapter/

端口适配层，实现 Domain/Engine 定义的 interface。

| 包 | 职责 | 状态 |
|----|------|------|
| `storage/` | EventStore（memory → sqlite） | memory ✅ |
| `participant/` | ParticipantPort（stub → LLM） | stub ✅ / llm ✅ |
| `principal/` | Principal Confirmation 关 | stub ✅ |
| `model/` | Model Provider Adapter | openai_compat ✅ |
| `runtime/` | Agent Runtime Adapter | 规划 |
| `transport/` | Discord Principal 绑定 + 指令 | discord ✅ |
| `workspace/` | Meeting Markdown 产出读写 | fs ✅ |
| `profile/` | SOUL / AGENTS / USER 身份文件 | fs ✅ |
| `knowledge/` | MEMORY / memory/ 长期知识 | fs ✅ |

Domain 层不得 import 本目录。
