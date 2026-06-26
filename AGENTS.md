# RoundTable — Agent Instructions

Read **`.cursorrules`** and **`docs/CONSTITUTION.md`** before architectural or implementation work.

**Generic solutions only:** do not specialize Engine/synthesis/parsing for a single workspace or scenario. See **`.cursorrules` → Generic Solutions (Mandatory)** and **ADR-0011** §合成质量原则.

When creating git commits, read **`docs/COMMITS.md`** and use the structured commit format.

When running Go tooling for this repo (`go test`, `go mod tidy`, etc.), use:

```bash
export GOPROXY=https://goproxy.cn,direct
```

Or prefer **`make test`** / **`make tidy`** from the repo root (Makefile sets `GOPROXY` automatically).

When debugging **Discord** or other blocked external APIs, set the local proxy (see **`.cursorrules` → Local network proxy**):

```bash
export https_proxy=http://127.0.0.1:7897 http_proxy=http://127.0.0.1:7897 all_proxy=socks5://127.0.0.1:7897
```

Or use **`make run-discord`** (proxy included).
