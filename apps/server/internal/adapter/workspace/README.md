# adapter/workspace/

Meeting 文件产出区（ADR-0009）。每个 Meeting 一个目录，Markdown 读写。

| 包 | 职责 | 状态 |
|----|------|------|
| `port.go` | `Port` interface、标准文件名常量 | ✅ |
| `fs/` | 本地文件系统实现 + path jail | ✅ |

运行时目录默认 `./data/workspaces/{meeting_id}/`（gitignore）。

Domain 不得 import 本包；Engine / Participant adapter 通过 `workspace.Port` 调用。
